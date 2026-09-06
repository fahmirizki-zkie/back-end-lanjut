package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"modul4/app/repository"
	"modul4/app/service"
	"modul4/config"
	"modul4/database"
)

func main() {
	// 1. Load konfigurasi
	cfg := config.Load()
	logger := config.NewLogger()

	// 2. Database
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

	// 3. Repository -> Service
	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	// 4. Aplikasi
	app := config.NewApp(logger, pool, studentService)

	// 5. Jalankan server
	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Println("server berhenti:", err)
		}
	}()

	log.Println("server berjalan di port 3000")

	// 6. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
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