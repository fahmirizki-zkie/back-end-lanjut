package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/helper"
)

type seedStudentData struct {
	NIM         string
	Nama        string
	Email       string
	Prodi       string
	Angkatan    int
	IPKTerakhir float64
}

type seedCourseData struct {
	KodeMK   string
	NamaMK   string
	SKS      int
	Semester int
	Kuota    int
}

func SeedDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	// Cek apakah sudah ada admin
	var adminCount int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
	if err != nil {
		return err
	}

	if adminCount > 0 {
		log.Println("[SEEDER] Database sudah memiliki data awal, skipping seeder.")
		return nil
	}

	log.Println("[SEEDER] Memulai seeding data awal...")

	// 1. Seed Admin
	adminPasswordHash, err := helper.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("gagal hash password admin: %w", err)
	}

	var adminID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password, role, created_at, updated_at)
		VALUES ($1, $2, 'admin', NOW(), NOW())
		RETURNING id
	`, "admin@siakad.ac.id", adminPasswordHash).Scan(&adminID)
	if err != nil {
		return fmt.Errorf("gagal seed admin: %w", err)
	}
	log.Printf("[SEEDER] Admin berhasil dibuat (ID: %d, email: admin@siakad.ac.id)\n", adminID)

	// 2. Seed 20 Mahasiswa
	studentsData := []seedStudentData{
		{"202301010001", "Ahmad Fauzi", "ahmad.fauzi@student.siakad.ac.id", "Teknik Informatika", 2023, 3.75},
		{"202301010002", "Budi Santoso", "budi.santoso@student.siakad.ac.id", "Teknik Informatika", 2023, 2.85},
		{"202301010003", "Citra Lestari", "citra.lestari@student.siakad.ac.id", "Sistem Informasi", 2023, 2.40},
		{"202301010004", "Dewi Anggraini", "dewi.anggraini@student.siakad.ac.id", "Sistem Informasi", 2023, 3.50},
		{"202301010005", "Eko Prasetyo", "eko.prasetyo@student.siakad.ac.id", "Teknik Komputer", 2023, 3.10},
		{"202201010006", "Fajar Hidayat", "fajar.hidayat@student.siakad.ac.id", "Teknik Informatika", 2022, 2.95},
		{"202201010007", "Gita Gutawa", "gita.gutawa@student.siakad.ac.id", "Sistem Informasi", 2022, 3.85},
		{"202201010008", "Hadi Saputra", "hadi.saputra@student.siakad.ac.id", "Teknik Komputer", 2022, 2.20},
		{"202201010009", "Indah Permata", "indah.permata@student.siakad.ac.id", "Teknik Informatika", 2022, 3.65},
		{"202201010010", "Joko Widodo", "joko.widodo@student.siakad.ac.id", "Sistem Informasi", 2022, 2.70},
		{"202401010011", "Kartika Sari", "kartika.sari@student.siakad.ac.id", "Teknik Informatika", 2024, 3.90},
		{"202401010012", "Lukman Hakim", "lukman.hakim@student.siakad.ac.id", "Teknik Komputer", 2024, 2.45},
		{"202401010013", "Mega Utami", "mega.utami@student.siakad.ac.id", "Sistem Informasi", 2024, 3.20},
		{"202401010014", "Nanda Pratama", "nanda.pratama@student.siakad.ac.id", "Teknik Informatika", 2024, 2.60},
		{"202401010015", "Oki Setiana", "oki.setiana@student.siakad.ac.id", "Teknik Komputer", 2024, 3.40},
		{"202301010016", "Putri Rahayu", "putri.rahayu@student.siakad.ac.id", "Sistem Informasi", 2023, 2.80},
		{"202301010017", "Qori Sandioriva", "qori.sandioriva@student.siakad.ac.id", "Teknik Informatika", 2023, 3.60},
		{"202301010018", "Rian Ardianto", "rian.ardianto@student.siakad.ac.id", "Teknik Komputer", 2023, 2.30},
		{"202301010019", "Siti Nurhaliza", "siti.nurhaliza@student.siakad.ac.id", "Sistem Informasi", 2023, 3.35},
		{"202301010020", "Taufik Hidayat", "taufik.hidayat@student.siakad.ac.id", "Teknik Informatika", 2023, 3.70},
	}

	for _, s := range studentsData {
		// Sesuai soal: password awal mahasiswa = NIM yang di-hash
		studentPassHash, err := helper.HashPassword(s.NIM)
		if err != nil {
			return fmt.Errorf("gagal hash password mahasiswa %s: %w", s.NIM, err)
		}

		var uID int64
		err = pool.QueryRow(ctx, `
			INSERT INTO users (email, password, role, created_at, updated_at)
			VALUES ($1, $2, 'mahasiswa', NOW(), NOW())
			RETURNING id
		`, s.Email, studentPassHash).Scan(&uID)
		if err != nil {
			return fmt.Errorf("gagal insert user mahasiswa %s: %w", s.Email, err)
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		`, uID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir)
		if err != nil {
			return fmt.Errorf("gagal insert student %s: %w", s.NIM, err)
		}
	}
	log.Println("[SEEDER] 20 Mahasiswa berhasil dibuat.")

	// 3. Seed 10 Mata Kuliah
	coursesData := []seedCourseData{
		{"IF101", "Dasar Pemrograman", 3, 1, 35},
		{"IF102", "Matematika Diskrit", 3, 1, 30},
		{"IF201", "Struktur Data & Algoritma", 4, 2, 30},
		{"IF202", "Basis Data", 3, 2, 40},
		{"IF301", "Pemrograman Berorientasi Objek", 3, 3, 35},
		{"IF302", "Arsitektur Komputer", 3, 3, 30},
		{"IF401", "Pemrograman Web Lanjut", 3, 4, 30},
		{"IF402", "Pemrograman Back-End", 4, 4, 25},
		{"IF501", "Kecerdasan Buatan", 3, 5, 30},
		{"IF601", "Keamanan Sistem & Jaringan", 3, 6, 25},
	}

	for _, c := range coursesData {
		_, err = pool.Exec(ctx, `
			INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (kode_mk) DO NOTHING
		`, c.KodeMK, c.NamaMK, c.SKS, c.Semester, c.Kuota)
		if err != nil {
			return fmt.Errorf("gagal insert course %s: %w", c.KodeMK, err)
		}
	}
	log.Println("[SEEDER] 10 Mata Kuliah berhasil dibuat.")

	log.Println("[SEEDER] Seeding selesai dengan sukses!")
	return nil
}
