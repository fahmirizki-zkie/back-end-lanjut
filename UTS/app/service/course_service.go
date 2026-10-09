package service

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type CourseService struct {
	courseRepo repository.CourseRepository
}

func NewCourseService(courseRepo repository.CourseRepository) *CourseService {
	return &CourseService{courseRepo: courseRepo}
}

// Endpoint 8: GET /api/v1/courses (Semua role)
func (s *CourseService) List(c *fiber.Ctx) error {
	ctx := context.Background()

	semester := c.Query("semester", "")
	search := c.Query("search", "")
	availableParam := c.Query("available", "false")
	availableOnly := availableParam == "true"

	courses, err := s.courseRepo.FindAll(ctx, semester, search, availableOnly)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", courses)
}
