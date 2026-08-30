package main

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"modul3/app/model"
	"modul3/app/repository"
)

type Handler struct {
	repo repository.StudentRepository
}

func NewHandler(repo repository.StudentRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) getStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	student, err := h.repo.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil student",
		)
	}

	return ok(c, "student ditemukan", student)
}

func (h *Handler) createStudent(c *fiber.Ctx) error {
	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa json yang valid",
		)
	}

	req.Name = strings.TrimSpace(req.Name)
	req.NIM = strings.TrimSpace(req.NIM)

	errs := map[string]string{}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "grade harus antara 0 dan 100"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student := model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	result, err := h.repo.Create(ctx, student)

	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return fail(
				c,
				fiber.StatusConflict,
				"nim sudah digunakan",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal membuat student",
		)
	}

	return created(
		c,
		"student berhasil dibuat",
		result,
		"/api/v1/students/"+strconv.Itoa(result.ID),
	)
}

func (h *Handler) replaceStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa json yang valid",
		)
	}

	req.Name = strings.TrimSpace(req.Name)
	req.NIM = strings.TrimSpace(req.NIM)

	errs := map[string]string{}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "grade harus antara 0 dan 100"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student := model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	result, err := h.repo.Update(ctx, student)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(
				c,
				fiber.StatusConflict,
				"nim sudah digunakan",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengubah student",
		)
	}

	return ok(
		c,
		"student berhasil diubah",
		result,
	)
}

func (h *Handler) deleteStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	err = h.repo.Delete(ctx, id)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal menghapus student",
		)
	}

	return noContent(c)
}

func (h *Handler) listStudent(c *fiber.Ctx) error {
	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	// Search
	search := strings.TrimSpace(c.Query("search"))

	// Sorting
	sortField := c.Query("sort", "id")
	order := strings.ToLower(c.Query("order", "asc"))

	// Filter is_active
	var isActive *bool

	isActiveParam := c.Query("is_active")

	if isActiveParam != "" {
		value, err := strconv.ParseBool(isActiveParam)

		if err != nil {
			return fail(
				c,
				fiber.StatusBadRequest,
				"is_active harus true atau false",
			)
		}

		isActive = &value
	}

	// Whitelist kolom sorting
	allowedSort := map[string]bool{
		"id":         true,
		"name":       true,
		"nim":        true,
		"grade":      true,
		"is_active":  true,
		"created_at": true,
	}

	if !allowedSort[sortField] {
		return fail(
			c,
			fiber.StatusBadRequest,
			"field sort tidak valid",
		)
	}

	if order != "asc" && order != "desc" {
		return fail(
			c,
			fiber.StatusBadRequest,
			"order harus asc atau desc",
		)
	}

	// Hitung OFFSET
	offset := (page - 1) * limit

	// Parameter untuk repository
	params := repository.ListParams{
		Search:   search,
		IsActive: isActive,
		Sort:     sortField,
		Order:    order,
		Limit:    limit,
		Offset:   offset,
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	// Repository sekarang mengembalikan data + total
	students, total, err := h.repo.FindAll(ctx, params)

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil daftar student",
		)
	}

	// Hitung total halaman
	totalPage := 0

	if total > 0 {
		totalPage = (total + limit - 1) / limit
	}

	return oklist(
		c,
		"daftar student berhasil diambil",
		students,
		&model.Meta{
			Page:      page,
			Limit:     limit,
			Total:     total,
			TotalPage: totalPage,
		},
	)
}