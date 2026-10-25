package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"modul10/app/repository"
	"modul10/app/service"
	"modul10/config"
	"modul10/database"
	"modul10/helper"
	"modul10/route"
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

	connString := config.GetEnv("DB_DSN", "")
	if connString == "" {
		connString = "postgres://" +
			cfg.DBUser + ":" +
			cfg.DBPassword + "@" +
			cfg.DBHost + ":" +
			cfg.DBPort + "/" +
			cfg.DBName +
			"?sslmode=" + cfg.DBSSLMode
	}

	pool, err := database.NewPostgresPool(
		context.Background(),
		connString,
	)
	if err != nil {
		log.Println("gagal terhubung ke database:", err)
		os.Exit(1)
	}
	defer pool.Close()

	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// Disconnect menutup seluruh koneksi di pool milik driver. Diberi batas
	// waktu sendiri, karena context milik request sudah tidak berlaku saat
	// aplikasi berhenti.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()
	mongoDB := mongoClient.Database(config.GetEnv("MONGO_DATABASE", "praktikum_backend"))
	_ = mongoDB

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
	courseRepository := repository.NewCourseRepository(pool)

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

	txManager := repository.NewTxManager(pool)
	courseService := service.NewCourseService(courseRepository, txManager)

	// App
	app := config.NewApp(
		logger,
		pool,
		route.Dependencies{
			StudentService: studentService,
			AuthService:    authService,
			CourseService:  courseService,
			Mongo:          mongoClient,
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
