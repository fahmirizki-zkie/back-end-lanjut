package model

import (
	"time"
)

type Student struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type StudentItem struct {
	ID          int64   `json:"id"`
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim"`
	Nama        string   `json:"nama"`
	Email       string   `json:"email"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

type EnrolledCourseItem struct {
	EnrollmentID   int64  `json:"enrollment_id"`
	CourseID       int64  `json:"course_id"`
	KodeMK         string `json:"kode_mk"`
	NamaMK         string `json:"nama_mk"`
	SKS            int    `json:"sks"`
	Semester       int    `json:"semester"`
	TahunAkademik  string `json:"tahun_akademik"`
}

type StudentDetailResponse struct {
	ID          int64                `json:"id"`
	UserID      int64                `json:"user_id"`
	NIM         string               `json:"nim"`
	Nama        string               `json:"nama"`
	Prodi       string               `json:"prodi"`
	Angkatan    int                  `json:"angkatan"`
	IPKTerakhir float64              `json:"ipk_terakhir"`
	TotalSKS    int                  `json:"total_sks"`
	BatasSKS    int                  `json:"batas_sks"`
	Courses     []EnrolledCourseItem `json:"courses"`
}
