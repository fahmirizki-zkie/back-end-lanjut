package route

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"
)

type Dependencies struct {
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
	JWT               *helper.JWTManager
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// Health check endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return helper.Success(c, fiber.StatusOK, "SIAKAD Mini API is running healthy", fiber.Map{
			"status": "UP",
		})
	})

	// 1. Endpoint Auth
	auth := api.Group("/auth")
	// Endpoint 1: POST /api/v1/auth/login (Publik + Rate Limiter 5x/menit)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	// Endpoint 2: GET /api/v1/auth/me (Semua role terautentikasi)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// 2. Endpoint Students (Wajib Auth)
	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	// Endpoint 3: GET /api/v1/students (Admin)
	students.Get("/", middleware.RequireRole("admin"), deps.StudentService.List)
	// Endpoint 4: POST /api/v1/students (Admin)
	students.Post("/", middleware.RequireRole("admin"), deps.StudentService.Create)
	// Endpoint 5: GET /api/v1/students/:id (Admin, Mahasiswa data sendiri)
	students.Get("/:id", deps.StudentService.Get)
	// Endpoint 6: PUT /api/v1/students/:id (Admin)
	students.Put("/:id", middleware.RequireRole("admin"), deps.StudentService.Update)
	// Endpoint 7: DELETE /api/v1/students/:id (Admin - Soft delete)
	students.Delete("/:id", middleware.RequireRole("admin"), deps.StudentService.Delete)

	// 3. Endpoint Courses (Semua role terautentikasi)
	courses := api.Group("/courses", middleware.RequireAuth(deps.JWT))
	// Endpoint 8: GET /api/v1/courses
	courses.Get("/", deps.CourseService.List)

	// 4. Endpoint Enrollments (Mahasiswa)
	enrollments := api.Group("/enrollments", middleware.RequireAuth(deps.JWT))
	// Endpoint 9: POST /api/v1/enrollments (Mahasiswa)
	enrollments.Post("/", middleware.RequireRole("mahasiswa"), deps.EnrollmentService.Create)
	// Endpoint 10: DELETE /api/v1/enrollments/:id (Mahasiswa milik sendiri)
	enrollments.Delete("/:id", middleware.RequireRole("mahasiswa"), deps.EnrollmentService.Delete)
}
