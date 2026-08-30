package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul3/app/repository"
	"modul3/config"
	"modul3/database"
)

func main() {
	// Load konfigurasi dari .env
	cfg := config.Load()

	// Buat connection string PostgreSQL
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	// Buat connection pool PostgreSQL
	pool, err := database.NewPostgresPool(
		context.Background(),
		connString,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	app := fiber.New()

	// Middleware Content-Type
	app.Use(func(c *fiber.Ctx) error {
		method := c.Method()

		if method == fiber.MethodPost ||
			method == fiber.MethodPut ||
			method == fiber.MethodPatch {

			contentType := c.Get("Content-Type")

			if !strings.HasPrefix(contentType, "application/json") {
				return fail(
					c,
					fiber.StatusUnsupportedMediaType,
					"Content-Type harus application/json",
				)
			}
		}

		return c.Next()
	})

	// Repository
	studentRepo := repository.NewStudentRepository(pool)

	// Handler
	handler := NewHandler(studentRepo)

	// Route API
	api := app.Group("/api/v1")
	students := api.Group("/students")

	students.Get("/", handler.listStudent)
	students.Get("/:id", handler.getStudent)
	students.Post("/", handler.createStudent)
	students.Put("/:id", handler.replaceStudent)
	students.Delete("/:id", handler.deleteStudent)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		if err := pool.Ping(c.UserContext()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "unhealthy",
				"database": "down",
			})
		}

		return c.JSON(fiber.Map{
			"status":   "healthy",
			"database": "up",
		})
	})

	log.Fatal(app.Listen(":3000"))
}