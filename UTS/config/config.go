package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config menampung seluruh variabel konfigurasi runtime aplikasi
type Config struct {
	AppName    string
	AppEnv     string
	AppPort    string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
}

// Load membaca file .env dan memetakan ke struct Config
func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppName:    GetEnv("APP_NAME", "SIAKAD Mini API"),
		AppEnv:     GetEnv("APP_ENV", "development"),
		AppPort:    GetEnv("APP_PORT", "8080"),
		DBHost:     GetEnv("DB_HOST", "localhost"),
		DBPort:     GetEnv("DB_PORT", "5432"),
		DBUser:     GetEnv("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     GetEnv("DB_NAME", "siakad_mini_fahmi"),
		DBSSLMode:  GetEnv("DB_SSLMODE", "disable"),
		JWTSecret:  GetEnv("JWT_SECRET", "supersecretkeysiakadminifahmi2026jwt"),
	}
}
