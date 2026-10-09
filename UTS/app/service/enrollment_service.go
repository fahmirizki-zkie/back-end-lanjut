package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type EnrollmentService struct {
	pool           *pgxpool.Pool
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
	studentRepo    repository.StudentRepository
}

func NewEnrollmentService(
	pool *pgxpool.Pool,
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
	studentRepo repository.StudentRepository,
) *EnrollmentService {
	return &EnrollmentService{
		pool:           pool,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		studentRepo:    studentRepo,
	}
}

// Regex format tahun akademik, contoh: "2026/2027-Ganjil" atau "2026/2027-Genap"
var tahunAkademikRegex = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

// Endpoint 9: POST /api/v1/enrollments (Mahasiswa)
func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	ctx := context.Background()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	if authUser.Role != "mahasiswa" {
		return helper.Forbidden("Hanya mahasiswa yang dapat mengambil mata kuliah")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format request JSON tidak valid")
	}

	valErrors := make(map[string][]string)

	if req.CourseID <= 0 {
		valErrors["course_id"] = append(valErrors["course_id"], "course_id wajib diisi")
	}

	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)
	if req.TahunAkademik == "" {
		valErrors["tahun_akademik"] = append(valErrors["tahun_akademik"], "tahun_akademik wajib diisi")
	} else if !tahunAkademikRegex.MatchString(req.TahunAkademik) {
		valErrors["tahun_akademik"] = append(valErrors["tahun_akademik"], "Format tahun_akademik harus seperti 2026/2027-Ganjil atau 2026/2027-Genap")
	}

	if len(valErrors) > 0 {
		return helper.Validation(valErrors)
	}

	// Cari student record dari user_id yang sedang login
	student, err := s.studentRepo.FindByUserID(ctx, authUser.UserID)
	if err != nil {
		return helper.Internal(err)
	}
	if student == nil || student.DeletedAt != nil {
		return helper.NotFound("Data mahasiswa tidak ditemukan atau telah dinonaktifkan")
	}

	// MULAI TRANSAKSI DATABASE DENGAN ROW LOCKING
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	defer tx.Rollback(ctx)

	// 1. Cek duplikasi: tidak dapat mengambil MK yang sama dua kali pada tahun akademik yang sama
	alreadyEnrolled, err := s.enrollmentRepo.CheckDuplicate(ctx, tx, student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return helper.Internal(err)
	}
	if alreadyEnrolled {
		return helper.Conflict("Mata kuliah sudah pernah diambil pada tahun akademik yang sama")
	}

	// 2. Kunci baris mata kuliah dengan ROW LOCKING (SELECT ... FOR UPDATE)
	course, err := s.courseRepo.GetForUpdate(ctx, tx, req.CourseID)
	if err != nil {
		return helper.Internal(err)
	}
	if course == nil {
		return helper.NotFound("Mata kuliah tidak ditemukan")
	}

	// 3. Cek sisa kuota mata kuliah
	terisi, err := s.courseRepo.CountEnrolled(ctx, tx, course.ID, req.TahunAkademik)
	if err != nil {
		return helper.Internal(err)
	}
	if terisi >= course.Kuota {
		return helper.Unprocessable("Kuota mata kuliah sudah penuh")
	}

	// 4. Cek batas SKS berdasarkan IPK mahasiswa
	// Hitung total SKS yang sudah diambil pada tahun akademik tersebut
	totalSKSSaatIni, err := s.enrollmentRepo.GetTotalSKS(ctx, tx, student.ID, req.TahunAkademik)
	if err != nil {
		return helper.Internal(err)
	}

	batasSKS := calculateBatasSKS(student.IPKTerakhir)
	sisaSKS := batasSKS - totalSKSSaatIni

	if totalSKSSaatIni+course.SKS > batasSKS {
		msg := fmt.Sprintf("Total SKS melebihi batas (%d SKS). Anda saat ini telah mengambil %d SKS, sisa SKS yang dapat diambil adalah %d SKS, sedangkan mata kuliah ini berbobot %d SKS",
			batasSKS, totalSKSSaatIni, sisaSKS, course.SKS)
		return helper.Unprocessable(msg)
	}

	// 5. Simpan record enrollment
	enrollment := model.Enrollment{
		StudentID:     student.ID,
		CourseID:      course.ID,
		TahunAkademik: req.TahunAkademik,
	}
	if err := s.enrollmentRepo.Create(ctx, tx, &enrollment); err != nil {
		return helper.Internal(err)
	}

	// COMMIT TRANSAKSI
	if err := tx.Commit(ctx); err != nil {
		return helper.Internal(err)
	}

	res := model.EnrollmentResponse{
		ID:            enrollment.ID,
		StudentID:     enrollment.StudentID,
		CourseID:      enrollment.CourseID,
		TahunAkademik: enrollment.TahunAkademik,
		CreatedAt:     enrollment.CreatedAt,
		Course: &model.CourseResponse{
			ID:        course.ID,
			KodeMK:    course.KodeMK,
			NamaMK:    course.NamaMK,
			SKS:       course.SKS,
			Semester:  course.Semester,
			Kuota:     course.Kuota,
			Terisi:    terisi + 1,
			SisaKuota: course.Kuota - (terisi + 1),
		},
	}

	return helper.Success(c, fiber.StatusCreated, "Mata kuliah berhasil ditambahkan ke KRS", res)
}

// Endpoint 10: DELETE /api/v1/enrollments/{id} (Mahasiswa pemilik KRS)
func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return helper.BadRequest("ID enrollment tidak valid")
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	if authUser.Role != "mahasiswa" {
		return helper.Forbidden("Hanya mahasiswa yang dapat membatalkan KRS")
	}

	// Cari student record
	student, err := s.studentRepo.FindByUserID(ctx, authUser.UserID)
	if err != nil {
		return helper.Internal(err)
	}
	if student == nil {
		return helper.Forbidden("Data mahasiswa tidak ditemukan")
	}

	// Ambil data enrollment
	enrollment, err := s.enrollmentRepo.FindByID(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}
	if enrollment == nil {
		return helper.NotFound("Enrollment tidak ditemukan")
	}

	// Cek kepemilikan: hanya boleh membatalkan KRS miliknya sendiri
	if enrollment.StudentID != student.ID {
		return helper.Forbidden("Anda tidak berhak membatalkan KRS mahasiswa lain")
	}

	// Hapus enrollment
	if err := s.enrollmentRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("Enrollment tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}
