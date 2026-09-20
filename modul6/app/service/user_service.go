package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul6/app/model"
	"modul6/app/repository"
	"modul6/helper"
)

// StudentService memegang dua tanggung jawab sekaligus pada struktur baku
// mata kuliah ini: menerima *fiber.Ctx (peran controller) dan menjalankan
// business rules (peran use case).
type StudentService struct {
	repo repository.StudentRepository
}

// NewStudentService menerima INTERFACE, bukan struct konkret.
func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	params := repository.ListParams{
		Search:   q.Search,
		IsActive: q.IsActive,
		Sort:     q.Sort,
		Order:    q.Order,
		Limit:    q.Limit,
		Offset:   (q.Page - 1) * q.Limit,
	}

	students, total, err := s.repo.FindAll(ctx, params)
	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
		)
	}

	return helper.SuccessList(
		c,
		"daftar student berhasil diambil",
		students,
		&model.Meta{
			Page:      q.Page,
			Limit:     q.Limit,
			Total:     total,
			TotalPage: CountTotalPages(total, q.Limit),
		},
	)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data student")
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"student ditemukan",
		student,
	)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.NIM = strings.TrimSpace(req.NIM)

	// Business rulesnya dipanggil, bukan ditulis ulang di sini.
	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	newStudent, err := s.repo.Create(ctx, model.Student{
		Name:     req.Name,
		Email:    req.Email,
		NIM:      req.NIM,
		Grade:    req.Grade,
		Password: req.Password,
		IsActive: true,
	})

	if err != nil {
		return translateError(c, err, "gagal menyimpan student")
	}

	return helper.Created(
		c,
		"student berhasil dibuat",
		newStudent,
		"/api/v1/students/"+strconv.Itoa(newStudent.ID),
	)
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.NIM = strings.TrimSpace(req.NIM)

	if errs := ValidateReplace(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		NIM:      req.NIM,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})

	if err != nil {
		return translateError(c, err, "gagal memperbarui student")
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"student berhasil diganti seluruhnya",
		result,
	)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	var req model.UpdateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	if IsEmptyPatch(req) {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"tidak ada field yang diubah",
		)
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data student")
	}

	updated, errs := ApplyPatch(current, req)

	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(c, err, "gagal memperbarui student")
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"student berhasil diperbarui sebagian",
		result,
	)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err, "gagal menghapus student")
	}

	return helper.NoContent(c)
}

// translateError memetakan error milik repository menjadi status HTTP.
func translateError(c *fiber.Ctx, err error, generalMessage string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)

	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(
			c,
			fiber.StatusConflict,
			"NIM atau email sudah dipakai",
		)

	default:
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			generalMessage,
		)
	}
}