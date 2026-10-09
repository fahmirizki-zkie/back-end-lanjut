package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type StudentService struct {
	pool           *pgxpool.Pool
	studentRepo    repository.StudentRepository
	userRepo       repository.UserRepository
	enrollmentRepo repository.EnrollmentRepository
}

func NewStudentService(
	pool *pgxpool.Pool,
	studentRepo repository.StudentRepository,
	userRepo repository.UserRepository,
	enrollmentRepo repository.EnrollmentRepository,
) *StudentService {
	return &StudentService{
		pool:           pool,
		studentRepo:    studentRepo,
		userRepo:       userRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

// Endpoint 3: GET /api/v1/students (Admin)
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx := context.Background()

	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	if perPage < 1 {
		perPage = 10
	} else if perPage > 50 {
		perPage = 50
	}

	prodi := c.Query("prodi", "")
	angkatan := c.Query("angkatan", "")
	search := c.Query("search", "")
	sort := c.Query("sort", "")

	students, total, err := s.studentRepo.FindAll(ctx, page, perPage, prodi, angkatan, search, sort)
	if err != nil {
		return helper.Internal(err)
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage == 0 {
		lastPage = 1
	}

	meta := &model.Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	}

	return helper.SuccessList(c, "Data mahasiswa berhasil diambil", students, meta)
}

// Endpoint 4: POST /api/v1/students (Admin)
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx := context.Background()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format request JSON tidak valid")
	}

	valErrors := make(map[string][]string)

	req.NIM = strings.TrimSpace(req.NIM)
	if req.NIM == "" {
		valErrors["nim"] = append(valErrors["nim"], "NIM wajib diisi")
	} else if len(req.NIM) != 12 {
		valErrors["nim"] = append(valErrors["nim"], "NIM harus tepat 12 digit")
	} else {
		// Validasi numerik
		if _, err := strconv.ParseUint(req.NIM, 10, 64); err != nil {
			valErrors["nim"] = append(valErrors["nim"], "NIM harus berupa angka")
		} else {
			// Cek keunikan NIM
			exists, err := s.studentRepo.CheckNIMExists(ctx, req.NIM)
			if err != nil {
				return helper.Internal(err)
			}
			if exists {
				valErrors["nim"] = append(valErrors["nim"], "NIM sudah terdaftar")
			}
		}
	}

	req.Nama = strings.TrimSpace(req.Nama)
	if req.Nama == "" {
		valErrors["nama"] = append(valErrors["nama"], "Nama wajib diisi")
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		valErrors["email"] = append(valErrors["email"], "Email wajib diisi")
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		valErrors["email"] = append(valErrors["email"], "Format email tidak valid")
	} else {
		// Cek keunikan email
		exists, err := s.userRepo.CheckEmailExists(ctx, req.Email)
		if err != nil {
			return helper.Internal(err)
		}
		if exists {
			valErrors["email"] = append(valErrors["email"], "Email sudah terdaftar")
		}
	}

	req.Prodi = strings.TrimSpace(req.Prodi)
	if req.Prodi == "" {
		valErrors["prodi"] = append(valErrors["prodi"], "Program studi wajib diisi")
	}

	currentYear := time.Now().Year()
	if req.Angkatan == 0 {
		valErrors["angkatan"] = append(valErrors["angkatan"], "Angkatan wajib diisi")
	} else if req.Angkatan < 1900 || req.Angkatan > currentYear {
		valErrors["angkatan"] = append(valErrors["angkatan"], fmt.Sprintf("Angkatan harus 4 digit dan <= tahun berjalan (%d)", currentYear))
	}

	var ipk float64
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
		if ipk < 0.00 || ipk > 4.00 {
			valErrors["ipk_terakhir"] = append(valErrors["ipk_terakhir"], "IPK terakhir harus di antara 0.00 dan 4.00")
		}
	}

	if len(valErrors) > 0 {
		return helper.Validation(valErrors)
	}

	// Password awal = NIM yang di-hash
	hashedPass, err := helper.HashPassword(req.NIM)
	if err != nil {
		return helper.Internal(err)
	}

	// Eksekusi dalam satu transaksi database
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	defer tx.Rollback(ctx)

	// 1. Buat record users (role mahasiswa)
	newUser := model.User{
		Email:    req.Email,
		Password: hashedPass,
		Role:     "mahasiswa",
	}
	if err := s.userRepo.Create(ctx, tx, &newUser); err != nil {
		return helper.Internal(err)
	}

	// 2. Buat record students
	newStudent := model.Student{
		UserID:      newUser.ID,
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: ipk,
	}
	if err := s.studentRepo.Create(ctx, tx, &newStudent); err != nil {
		return helper.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Internal(err)
	}

	res := model.StudentItem{
		ID:          newStudent.ID,
		NIM:         newStudent.NIM,
		Nama:        newStudent.Nama,
		Prodi:       newStudent.Prodi,
		Angkatan:    newStudent.Angkatan,
		IPKTerakhir: newStudent.IPKTerakhir,
	}

	return helper.Success(c, fiber.StatusCreated, "Mahasiswa dan akun pengguna berhasil dibuat", res)
}

// Endpoint 5: GET /api/v1/students/{id} (Admin, Mahasiswa data sendiri)
func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return helper.BadRequest("ID mahasiswa tidak valid")
	}

	student, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}
	if student == nil || student.DeletedAt != nil {
		return helper.NotFound("Mahasiswa tidak ditemukan atau sudah dihapus")
	}

	// Cek otorisasi kepemilikan
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	if authUser.Role == "mahasiswa" && authUser.UserID != student.UserID {
		return helper.Forbidden("Anda tidak berhak mengakses data mahasiswa lain")
	}

	// Ambil daftar mata kuliah yang diambil
	courses, err := s.enrollmentRepo.GetCoursesByStudentID(ctx, student.ID)
	if err != nil {
		return helper.Internal(err)
	}

	totalSKS, err := s.enrollmentRepo.GetTotalSKSByStudentID(ctx, student.ID)
	if err != nil {
		return helper.Internal(err)
	}

	// Hitung batas SKS berdasarkan IPK
	batasSKS := calculateBatasSKS(student.IPKTerakhir)

	res := model.StudentDetailResponse{
		ID:          student.ID,
		UserID:      student.UserID,
		NIM:         student.NIM,
		Nama:        student.Nama,
		Prodi:       student.Prodi,
		Angkatan:    student.Angkatan,
		IPKTerakhir: student.IPKTerakhir,
		TotalSKS:    totalSKS,
		BatasSKS:    batasSKS,
		Courses:     courses,
	}

	return helper.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", res)
}

// Endpoint 6: PUT /api/v1/students/{id} (Admin)
func (s *StudentService) Update(c *fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return helper.BadRequest("ID mahasiswa tidak valid")
	}

	student, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}
	if student == nil || student.DeletedAt != nil {
		return helper.NotFound("Mahasiswa tidak ditemukan")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format request JSON tidak valid")
	}

	valErrors := make(map[string][]string)

	req.Nama = strings.TrimSpace(req.Nama)
	if req.Nama == "" {
		valErrors["nama"] = append(valErrors["nama"], "Nama wajib diisi")
	}

	req.Prodi = strings.TrimSpace(req.Prodi)
	if req.Prodi == "" {
		valErrors["prodi"] = append(valErrors["prodi"], "Program studi wajib diisi")
	}

	currentYear := time.Now().Year()
	if req.Angkatan == 0 {
		valErrors["angkatan"] = append(valErrors["angkatan"], "Angkatan wajib diisi")
	} else if req.Angkatan < 1900 || req.Angkatan > currentYear {
		valErrors["angkatan"] = append(valErrors["angkatan"], fmt.Sprintf("Angkatan harus 4 digit dan <= tahun berjalan (%d)", currentYear))
	}

	if req.IPKTerakhir != nil {
		if *req.IPKTerakhir < 0.00 || *req.IPKTerakhir > 4.00 {
			valErrors["ipk_terakhir"] = append(valErrors["ipk_terakhir"], "IPK terakhir harus di antara 0.00 dan 4.00")
		} else {
			student.IPKTerakhir = *req.IPKTerakhir
		}
	}

	if len(valErrors) > 0 {
		return helper.Validation(valErrors)
	}

	student.Nama = req.Nama
	student.Prodi = req.Prodi
	student.Angkatan = req.Angkatan

	if err := s.studentRepo.Update(ctx, student); err != nil {
		return helper.Internal(err)
	}

	res := model.StudentItem{
		ID:          student.ID,
		NIM:         student.NIM,
		Nama:        student.Nama,
		Prodi:       student.Prodi,
		Angkatan:    student.Angkatan,
		IPKTerakhir: student.IPKTerakhir,
	}

	return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", res)
}

// Endpoint 7: DELETE /api/v1/students/{id} (Admin - Soft Delete)
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return helper.BadRequest("ID mahasiswa tidak valid")
	}

	err = s.studentRepo.SoftDelete(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("Mahasiswa tidak ditemukan atau sudah dihapus")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}

// calculateBatasSKS menghitung batas SKS berdasarkan IPK sesuai business rule UTS
func calculateBatasSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	} else if ipk >= 2.50 {
		return 21
	}
	return 18
}
