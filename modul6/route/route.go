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

// Dependencies berisi semua dependensi yang dibutuhkan route
func Register(app *fiber.App, pool *pgxpool.Pool, deps Dependencies) {
    api := app.Group("/api/v1")

    // publik
    api.Get("/health", healthCheck(pool))

    // autentikasi
    auth := api.Group("/auth", middleware.RequireJSON)
    auth.Post("/register", deps.AuthService.Register)
    auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
    auth.Post("/refresh", deps.AuthService.Refresh)
    auth.Post("/logout", deps.AuthService.Logout)
    auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

    // wajib login, hak akses diperiksa per endpoint
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