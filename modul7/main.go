package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"modul6/app/repository"
	"modul6/app/service"
	"modul6/config"
	"modul6/database"
	"modul6/helper"
	"modul6/route"
)

// minSecretLength panjang minimum JWT_SECRET (32 karakter = 256 bit)
const minSecretLength = 32

func main() {
	cfg := config.Load()
	logger := config.NewLogger()

	// Periksa JWT_SECRET sebelum server menyala
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error(
			"JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength),
		)
		os.Exit(1)
	}

	connString := "postgres://" +
		cfg.DBUser + ":" +
		cfg.DBPassword + "@" +
		cfg.DBHost + ":" +
		cfg.DBPort + "/" +
		cfg.DBName +
		"?sslmode=" + cfg.DBSSLMode

	pool, err := database.NewPostgresPool(
		context.Background(),
		connString,
	)
	if err != nil {
		log.Println("gagal terhubung ke database:", err)
		os.Exit(1)
	}
	defer pool.Close()

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(
			config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15),
		)*time.Minute,
	)

	// Repository
	studentRepository := repository.NewStudentRepository(pool)
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)

	// Pemetaan role ke permission dibaca sekali saat aplikasi menyala.
	rawPermissions, err := roleRepository.LoadPermissions(
		context.Background(),
	)
	if err != nil {
		logger.Error(
			"gagal memuat permission",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	permissions := helper.NewPermissionSet(rawPermissions)

	logger.Info(
		"permission dimuat",
		slog.Any("roles", permissions.KnownRoles()),
	)

	// Service
	studentService := service.NewStudentService(
		studentRepository,
		permissions,
	)

	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		jwtManager,
		time.Duration(
			config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7),
		)*24*time.Hour,
	)

	// App
	app := config.NewApp(
		logger,
		pool,
		route.Dependencies{
			StudentService: studentService,
			AuthService:    authService,
			JWT:            jwtManager,
			Permissions:    permissions,
		},
	)

	// Jalankan server
	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Println("server berhenti:", err)
		}
	}()

	log.Println("server berjalan di port 3000")

	// Menunggu sinyal berhenti
	quit := make(chan os.Signal, 1)
	signal.Notify(
		quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-quit

	log.Println("sinyal berhenti diterima, menutup server")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Println("gagal menutup server:", err)
	}

	log.Println("server berhenti dengan rapi")
}
