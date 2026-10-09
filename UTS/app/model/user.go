package model

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserSummary struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Request login: email (wajib, format email), password (wajib, min. 8 karakter)
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response login: access_token, token_type, expires_in, dan data user (id, email, role)
type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int         `json:"expires_in"` // detik
	User        UserSummary `json:"user"`
}

// Profil /api/v1/auth/me
type AuthMeResponse struct {
	ID      int64             `json:"id"`
	Email   string            `json:"email"`
	Role    string            `json:"role"`
	Student *StudentMeSummary `json:"student,omitempty"`
}

type StudentMeSummary struct {
	ID       int64  `json:"id"`
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}
