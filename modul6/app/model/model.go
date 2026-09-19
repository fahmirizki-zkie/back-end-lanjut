package model

import "time"


type Student struct {
	ID        int    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	NIM       string    `json:"nim"`
	Password  string    `json:"-"`
	Grade	  float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`  
}

type Nilai struct {
	ID         int           `json:"id"`
	IDStudent  int           `json:"id_student"`
	Matakuliah string        `json:"mata_kuliah"`
	Nilai      float64       `json:"nilai"`
}

// Post semua filed wajib 
type CreateStudentRequest struct {
	Name	 string  `json:"name"`
	Email    string  `json:"email"`
	NIM      string  `json:"nim"`
	Grade    float64 `json:"grade"`
	Password string  `json:"password"`
}

// Put ganti seluruh isi, jadi field bertipe biasa dan wajib diisi semua 
type ReplaceStudentRequest struct {
	Name 	 string `json:"name"`
	Email    string `json:"email"`
	NIM      string `json:"nim"`
	Grade    float64 `json:"grade"`
	IsActive bool   `json:"is_active"`
}

// Patch ganti sebagian isi, jadi menggunakan pointer 
// supaya bisa membedakan antara tidak dikiirim (nil) dan dikirim bernilai kosong
type UpdateStudentRequest struct {
	Name 	 *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	NIM      *string `json:"nim,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Amplop baku untuk semua respone 
type WebResponse struct {
	Success		bool        `json:"success"`
	Message 	string      `json:"message"`
	Data    	any 		`json:"data,omitempty"`
	Meta		*Meta		`json:"meta,omitempty"`
	Error		any			`json:"error,omitempty"`
}

type Meta struct {
	Page 		int 	`json:"page"`
	Limit 		int 	`json:"limit"`
	Total 		int 	`json:"total"`
	TotalPage 	int 	`json:"total_page"`
}

type ListQuery struct {
	Page 		int 	
	Limit 	    int 	
	Search 	    string
	Sort 	    string
	Order 		string
	IsActive 	*bool
}

