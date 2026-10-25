package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	NIM       string    `json:"nim"`
	Password  string    `json:"-"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	OwnerID   int       `json:"owner_id"`
}

// CreateStudentRequest dengan validasi deklaratif
type CreateStudentRequest struct {
	Name     string  `json:"name" validate:"required,min=2,max=100"`
	Email    string  `json:"email" validate:"required,email,max=120"`
	NIM      string  `json:"nim" validate:"required,nim"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	Password string  `json:"password" validate:"required,min=8,max=72,nospace"`
}

// ReplaceStudentRequest untuk PUT (seluruh field wajib diisi)
type ReplaceStudentRequest struct {
	Name     string  `json:"name" validate:"required,min=2,max=100"`
	Email    string  `json:"email" validate:"required,email,max=120"`
	NIM      string  `json:"nim" validate:"required,nim"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}

// UpdateStudentRequest untuk PATCH (pointer field dengan omitnil)
type UpdateStudentRequest struct {
	Name     *string  `json:"name,omitempty" validate:"omitnil,min=2,max=100"`
	Email    *string  `json:"email,omitempty" validate:"omitnil,email,max=120"`
	NIM      *string  `json:"nim,omitempty" validate:"omitnil,nim"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,gte=0,lte=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// Amplop baku untuk semua respon
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

type Meta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

// CursorMeta menggantikan Meta pada endpoint yang memakai cursor.
//
// Perhatikan tidak adanya Total dan TotalPages. Keduanya tidak dapat
// disediakan tanpa COUNT(*) atas seluruh tabel — persis biaya yang ingin
// dihindari oleh pagination berbasis cursor.
type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Cursor adalah penanda posisi pada keyset pagination.
type Cursor struct {
	CreatedAt time.Time
	ID        int
}

// CursorQuery adalah parameter untuk FindAfterCursor.
type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

// User model dari Modul 6 — digunakan oleh auth dan user service.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // tidak pernah dikirim ke client
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Mulai pertemuan ini, aturan validasi ditulis sebagai tag pada struct.
// Aturan dan bentuk data berada pada baris yang sama, sehingga menambah
// satu field tanpa aturannya menjadi kelalaian yang langsung terlihat.
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace"`
}

type ReplaceUserRequest struct {
	Username string `json:"username"  validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email"     validate:"required,email,max=120"`
	IsActive bool   `json:"is_active"`
}

// Pada PATCH, pointer membedakan "tidak dikirim" (nil) dari "dikirim
// bernilai kosong". omitnil dipilih karena ia menyatakan maksud yang
// sebenarnya: lewati hanya bila nil.
// [Perbaikan Kesalahan #7]: di modul tertulis Username string (bukan *string),
// sehingga compiler menolak req.Username != nil dan *req.Username di user_rules.go.
type PatchUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3,max=30,alphanum"`
	Email    *string `json:"email,omitempty"   validate:"omitnil,email,max=120"`
	IsActive *bool   `json:"is_active,omitempty"`
}
