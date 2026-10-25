package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"modul9/app/model"
	"modul9/helper"
)

func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		// Catatan: out chan<- string (diperbaiki dari out <-chan string yang salah di modul)
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}
	close(start) // buka gerbang
	wg.Wait()   // tunggu SEMUA goroutine selesai
	close(results)
	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

type APIResponse struct {
	Success bool            `json:"success"`
	Code    string          `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func main() {
	mode := flag.String("mode", "crowd", "Mode pengujian: crowd | credits | swap")
	baseURL := flag.String("url", "http://localhost:3000", "Base URL backend")
	flag.Parse()

	// Load .env
	_ = godotenv.Load(".env")
	_ = godotenv.Load(filepath.Join("..", "..", ".env"))

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-jwt-key-minimum-32-chars-long!"
	}
	jwtManager := helper.NewJWTManager(jwtSecret, "loadtest", 2*time.Hour)

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbUser := os.Getenv("DB_USER")
		if dbUser == "" {
			dbUser = "postgres"
		}
		dbPass := os.Getenv("DB_PASSWORD")
		if dbPass == "" {
			dbPass = "postgres"
		}
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
		dbPort := os.Getenv("DB_PORT")
		if dbPort == "" {
			dbPort = "5432"
		}
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "praktikum_backend"
		}
		dbDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer pool.Close()

	// Ambil daftar user uji mhs001 - mhs060
	rows, err := pool.Query(ctx, "SELECT id, username, email, role FROM users WHERE username LIKE 'mhs%' ORDER BY username")
	if err != nil {
		log.Fatalf("Gagal mengambil user mhs: %v", err)
	}
	defer rows.Close()

	var testUsers []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role); err != nil {
			log.Fatalf("Scan user: %v", err)
		}
		testUsers = append(testUsers, u)
	}

	if len(testUsers) < 60 {
		log.Printf("Peringatan: hanya ditemukan %d user mahasiswa (disarankan jalankan migration 006)", len(testUsers))
	}

	// Buat token JWT untuk setiap user
	tokens := make(map[string]string)
	for _, u := range testUsers {
		token, err := jwtManager.GenerateAccessToken(u)
		if err != nil {
			log.Fatalf("Gagal membuat token: %v", err)
		}
		tokens[u.Username] = token
	}

	client := &http.Client{Timeout: 10 * time.Second}

	switch *mode {
	case "crowd":
		runCrowd(ctx, pool, client, *baseURL, testUsers, tokens)
	case "credits":
		runCredits(ctx, pool, client, *baseURL, testUsers, tokens)
	case "swap":
		runSwap(ctx, pool, client, *baseURL, testUsers, tokens)
	default:
		log.Fatalf("Mode tidak dikenal: %s (pilihan: crowd, credits, swap)", *mode)
	}
}

func sendRequest(client *http.Client, method, url, token string, body any) string {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return "REQ_ERROR"
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "NET_ERROR"
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	_ = json.NewDecoder(resp.Body).Decode(&apiResp)

	if resp.StatusCode == http.StatusOK {
		return "200 OK"
	}
	if resp.StatusCode == http.StatusCreated {
		return "201 OK"
	}
	if apiResp.Code != "" {
		return fmt.Sprintf("%d %s", resp.StatusCode, apiResp.Code)
	}
	return fmt.Sprintf("%d %s", resp.StatusCode, resp.Status)
}

func printTally(tally map[string]int) {
	keys := make([]string, 0, len(tally))
	for k := range tally {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println("hasil request:")
	for _, k := range keys {
		fmt.Printf("  %-25s %d\n", k, tally[k])
	}
}

func runCrowd(ctx context.Context, pool *pgxpool.Pool, client *http.Client, baseURL string, users []model.User, tokens map[string]string) {
	fmt.Println("60 mahasiswa berebut kelas IF399 secara serentak...")

	var courseID, quota, enrolledCount int
	err := pool.QueryRow(ctx, "SELECT id, quota, enrolled_count FROM courses WHERE code = 'IF399'").Scan(&courseID, &quota, &enrolledCount)
	if err != nil {
		log.Fatalf("Gagal menemukan kelas IF399: %v", err)
	}

	n := 60
	if len(users) < n {
		n = len(users)
	}

	start := time.Now()
	tally := fire(n, func(i int) string {
		u := users[i]
		url := fmt.Sprintf("%s/api/v1/courses/%d/enroll", baseURL, courseID)
		return sendRequest(client, http.MethodPost, url, tokens[u.Username], nil)
	})
	elapsed := time.Since(start)

	printTally(tally)
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())

	// Verifikasi database
	var countRows int
	_ = pool.QueryRow(ctx, "SELECT enrolled_count FROM courses WHERE id = $1", courseID).Scan(&enrolledCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", courseID).Scan(&countRows)

	status := "-> TIDAK KONSISTEN"
	if enrolledCount == countRows && enrolledCount <= quota {
		status = "-> KONSISTEN"
	}
	fmt.Printf("IF399 kuota=%d  enrolled_count=%d  baris enrollments=%d  %s\n", quota, enrolledCount, countRows, status)
}

func runCredits(ctx context.Context, pool *pgxpool.Pool, client *http.Client, baseURL string, users []model.User, tokens map[string]string) {
	fmt.Println("10 mahasiswa, masing-masing mendaftar ke 12 kelas 3 SKS secara serentak (batas 24 SKS)...")

	rows, err := pool.Query(ctx, "SELECT id FROM courses WHERE credits = 3 ORDER BY id LIMIT 12")
	if err != nil {
		log.Fatalf("Gagal mengambil courses: %v", err)
	}
	defer rows.Close()

	var courseIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			courseIDs = append(courseIDs, id)
		}
	}

	numStudents := 10
	if len(users) < numStudents {
		numStudents = len(users)
	}

	totalJobs := numStudents * len(courseIDs)

	type JobSpec struct {
		user     model.User
		courseID int
	}
	var jobs []JobSpec
	for i := 0; i < numStudents; i++ {
		for _, cID := range courseIDs {
			jobs = append(jobs, JobSpec{user: users[i], courseID: cID})
		}
	}

	start := time.Now()
	tally := fire(totalJobs, func(i int) string {
		job := jobs[i]
		url := fmt.Sprintf("%s/api/v1/courses/%d/enroll", baseURL, job.courseID)
		return sendRequest(client, http.MethodPost, url, tokens[job.user.Username], nil)
	})
	elapsed := time.Since(start)

	printTally(tally)
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())

	// Verifikasi database: setiap mahasiswa <= 24 SKS
	allConsistent := true
	for i := 0; i < numStudents; i++ {
		var totalCredits int
		_ = pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(c.credits), 0)
			FROM enrollments e
			JOIN courses c ON c.id = e.course_id
			WHERE e.user_id = $1`, users[i].ID).Scan(&totalCredits)
		if totalCredits > 24 {
			allConsistent = false
			fmt.Printf("User %s melebihi batas SKS: %d SKS\n", users[i].Username, totalCredits)
		}
	}

	if allConsistent {
		fmt.Println("seluruh 10 mahasiswa <= 24 SKS -> KONSISTEN")
	} else {
		fmt.Println("ada mahasiswa melebihi 24 SKS -> TIDAK KONSISTEN")
	}
}

func runSwap(ctx context.Context, pool *pgxpool.Pool, client *http.Client, baseURL string, users []model.User, tokens map[string]string) {
	fmt.Println("20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")

	var id101, id102, q101, q102 int
	_ = pool.QueryRow(ctx, "SELECT id, quota FROM courses WHERE code = 'IF101'").Scan(&id101, &q101)
	_ = pool.QueryRow(ctx, "SELECT id, quota FROM courses WHERE code = 'IF102'").Scan(&id102, &q102)

	if id101 == 0 || id102 == 0 {
		log.Fatalf("Kelas IF101 atau IF102 tidak ditemukan")
	}

	// Persiapkan state awal:
	// mhs001 - mhs020 terdaftar di IF101
	// mhs021 - mhs040 terdaftar di IF102
	for i := 0; i < 20 && i < len(users); i++ {
		_, _ = pool.Exec(ctx, "INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", id101, users[i].ID)
	}
	for i := 20; i < 40 && i < len(users); i++ {
		_, _ = pool.Exec(ctx, "INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", id102, users[i].ID)
	}
	_, _ = pool.Exec(ctx, "UPDATE courses SET enrolled_count = (SELECT COUNT(*) FROM enrollments WHERE course_id = courses.id) WHERE id IN ($1, $2)", id101, id102)

	// Ambil counter deadlock awal
	var deadlocksBefore int64
	_ = pool.QueryRow(ctx, "SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()").Scan(&deadlocksBefore)

	type SwapJob struct {
		user   model.User
		fromID int
		toID   int
	}
	var jobs []SwapJob
	// 20 mhs pindah IF101 -> IF102
	for i := 0; i < 20 && i < len(users); i++ {
		jobs = append(jobs, SwapJob{user: users[i], fromID: id101, toID: id102})
	}
	// 20 mhs pindah IF102 -> IF101
	for i := 20; i < 40 && i < len(users); i++ {
		jobs = append(jobs, SwapJob{user: users[i], fromID: id102, toID: id101})
	}

	start := time.Now()
	tally := fire(len(jobs), func(i int) string {
		job := jobs[i]
		url := fmt.Sprintf("%s/api/v1/enrollments/swap", baseURL)
		reqBody := map[string]int{
			"from_course_id": job.fromID,
			"to_course_id":   job.toID,
		}
		return sendRequest(client, http.MethodPost, url, tokens[job.user.Username], reqBody)
	})
	elapsed := time.Since(start)

	var deadlocksAfter int64
	_ = pool.QueryRow(ctx, "SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()").Scan(&deadlocksAfter)
	deadlocksDetected := deadlocksAfter - deadlocksBefore

	printTally(tally)
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", deadlocksDetected)

	var c101, rows101, c102, rows102 int
	_ = pool.QueryRow(ctx, "SELECT enrolled_count FROM courses WHERE id = $1", id101).Scan(&c101)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", id101).Scan(&rows101)
	_ = pool.QueryRow(ctx, "SELECT enrolled_count FROM courses WHERE id = $1", id102).Scan(&c102)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", id102).Scan(&rows102)

	s101 := "-> TIDAK KONSISTEN"
	if c101 == rows101 {
		s101 = "-> KONSISTEN"
	}
	s102 := "-> TIDAK KONSISTEN"
	if c102 == rows102 {
		s102 = "-> KONSISTEN"
	}

	fmt.Printf("IF101 kuota=%d  enrolled_count=%d  baris enrollments=%d  %s\n", q101, c101, rows101, s101)
	fmt.Printf("IF102 kuota=%d  enrolled_count=%d  baris enrollments=%d  %s\n", q102, c102, rows102, s102)
}
