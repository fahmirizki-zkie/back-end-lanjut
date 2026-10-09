package model

// WebResponse adalah format respons standar untuk semua endpoint SIAKAD Mini
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta pagination sesuai contoh soal UTS:
// { "current_page": 1, "per_page": 10, "total": 20, "last_page": 2 }
type Meta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}
