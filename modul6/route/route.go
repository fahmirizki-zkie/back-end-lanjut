package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul6/app/service"
	"modul6/helper"
	"modul6/middleware"
)

// Dependencies berisi semua dependensi yang dibutuhkan route.
type Dependencies struct {
	StudentService *service.StudentService
	AuthService    *service.AuthService
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
}

// Register mendaftarkan seluruh route aplikasi.
func Register(app *fiber.App, pool *pgxpool.Pool, deps Dependencies) {
	api := app.Group("/api/v1")

	// Publik
	api.Get("/health", healthCheck(pool))

	// Autentikasi
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// Wajib login, hak akses diperiksa per endpoint.
	students := api.Group(
		"/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)

	perms := deps.Permissions

	// Hak dapat diputuskan tanpa melihat isi data → middleware.
	students.Get(
		"/",
		middleware.RequirePermission(perms, "student:list"),
		deps.StudentService.List,
	)

	students.Post(
		"/",
		middleware.RequirePermission(perms, "student:create"),
		deps.StudentService.Create,
	)

	students.Delete(
		"/:id",
		middleware.RequirePermission(perms, "student:delete"),
		deps.StudentService.Delete,
	)

	// Hak bergantung pada kepemilikan data → diperiksa di service.
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)
}

// healthCheck mengecek koneksi database.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(
			c.Context(),
			2*time.Second,
		)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak tersedia",
			)
		}

		return helper.Success(
			c,
			fiber.StatusOK,
			"service sehat",
			fiber.Map{
				"status": "ok",
			},
		)
	}
}
