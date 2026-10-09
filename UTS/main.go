package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"siakad-mini/app/repository"
	"siakad-mini/app/service"
	"siakad-mini/config"
	"siakad-mini/database"
	"siakad-mini/helper"
	"siakad-mini/route"
)

func main() {
	cfg := config.Load()
	logger := config.NewLogger()

	logger.Info("Memulai SIAKAD Mini API server...")

	// Konfigurasi koneksi PostgreSQL
	var connString string
	if cfg.DBPassword == "" {
		connString = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
			cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	} else {
		connString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	}

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, connString)
	if err != nil {
		logger.Error("Gagal terhubung ke database PostgreSQL",
			slog.String("error", err.Error()),
			slog.String("dsn", connString),
		)
		log.Fatalf("Database connection error: %v", err)
	}
	defer pool.Close()

	logger.Info("Berhasil terhubung ke database PostgreSQL", slog.String("database", cfg.DBName))

	// Jalankan migrasi tabel otomatis jika belum ada
	if err := database.RunMigration(ctx, pool); err != nil {
		logger.Error("Gagal menjalankan migrasi database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Migrasi database berhasil dijalankan")

	// Jalankan seeder database (minimal 1 admin, 20 mahasiswa, 10 mata kuliah)
	if err := database.SeedDatabase(ctx, pool); err != nil {
		logger.Error("Gagal menjalankan seeder database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Inisialisasi JWT Manager
	jwtTTL := time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 1440)) * time.Minute
	jwtManager := helper.NewJWTManager(cfg.JWTSecret, config.GetEnv("JWT_ISSUER", "siakad-mini"), jwtTTL)

	// Inisialisasi Repository
	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	// Inisialisasi Service
	authService := service.NewAuthService(userRepo, studentRepo, jwtManager)
	studentService := service.NewStudentService(pool, studentRepo, userRepo, enrollmentRepo)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(pool, enrollmentRepo, courseRepo, studentRepo)

	deps := route.Dependencies{
		AuthService:       authService,
		StudentService:    studentService,
		CourseService:     courseService,
		EnrollmentService: enrollmentService,
		JWT:               jwtManager,
	}

	// Inisialisasi Fiber App
	app := config.NewApp(logger, deps)

	port := config.GetEnv("APP_PORT", "8080")
	logger.Info(fmt.Sprintf("Server berjalan di port :%s", port))

	if err := app.Listen(":" + port); err != nil {
		logger.Error("Server berhenti dengan error", slog.String("error", err.Error()))
	}
}
