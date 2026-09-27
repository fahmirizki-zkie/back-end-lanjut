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
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(
	repo repository.StudentRepository,
	perms *helper.PermissionSet,
) *StudentService {
	return &StudentService{
		repo:  repo,
		perms: perms,
	}
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
		return helper.Internal(err)
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

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "student")
	}

	if !CanAccessStudent(
		current,
		student.OwnerID,
		s.perms,
		"student:read:any",
	) {
		return helper.Forbidden("tidak berhak mengakses data student lain")
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

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.NIM = strings.TrimSpace(req.NIM)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	newStudent, err := s.repo.Create(ctx, model.Student{
		Name:     req.Name,
		Email:    req.Email,
		NIM:      req.NIM,
		Grade:    req.Grade,
		Password: req.Password,
		IsActive: true,
		OwnerID:  current.UserID,
	})

	if err != nil {
		return translateError(err, "student")
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

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "student")
	}

	if !CanAccessStudent(
		current,
		student.OwnerID,
		s.perms,
		"student:update:any",
	) {
		return helper.Forbidden("tidak berhak mengubah data student lain")
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.NIM = strings.TrimSpace(req.NIM)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	result, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		NIM:      req.NIM,
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  student.OwnerID,
	})

	if err != nil {
		return translateError(err, "student")
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

	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "student")
	}

	if !CanAccessStudent(
		currentUser,
		student.OwnerID,
		s.perms,
		"student:update:any",
	) {
		return helper.Forbidden("tidak berhak mengubah data student lain")
	}

	updated := ApplyPatch(student, req)

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(err, "student")
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
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(err, "student")
	}

	return helper.NoContent(c)
}

// translateError mengubah error milik repository menjadi AppError.
func translateError(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(entity + " tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("NIM atau email sudah dipakai")
	default:
		return helper.Internal(err)
	}
}
