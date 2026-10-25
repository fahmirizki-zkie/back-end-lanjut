package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"modul10/app/service"
	"modul10/helper"
	"modul10/middleware"
)

// Dependencies berisi semua dependensi yang dibutuhkan route.
type Dependencies struct {
	StudentService *service.StudentService
	AuthService    *service.AuthService
	CourseService  *service.CourseService
	Mongo          *mongo.Client
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
}

// Register mendaftarkan seluruh route aplikasi.
func Register(app *fiber.App, pool *pgxpool.Pool, deps Dependencies) {
	api := app.Group("/api/v1")

	// Publik
	api.Get("/health", healthCheck(pool, deps.Mongo))

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

	// --- kelas dan KRS ---
	courses := api.Group("/courses",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai
	// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id.
	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id",
		middleware.RequirePermission(perms, "course:manage"),
		deps.CourseService.UpdateQuota)

	enrollments := api.Group("/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	enrollments.Post("/swap", deps.CourseService.Swap)
}

// healthCheck kini memeriksa DUA database. Aplikasi dianggap sehat hanya
// bila keduanya dapat dihubungi, dan pesan kegagalannya menyebut yang mana.
func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}
		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}
