package service

import (
	"testing"

	"modul4/app/model"
)

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"terlalu pendek", "abc123", true},
		{"hanya huruf", "abcdefgh", true},
		{"hanya angka", "12345678", true},
		{"password umum", "password1", true},
		{"password valid", "Mahasiswa99", false},
		{"password valid dengan simbol", "Belajar!23", false},
		{"tepat 8 karakter valid", "abcd1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := checkPasswordStrength(tt.password)
			if tt.wantErr && msg == "" {
				t.Errorf("diharapkan error untuk password %q, tapi tidak ada", tt.password)
			}
			if !tt.wantErr && msg != "" {
				t.Errorf("tidak diharapkan error untuk password %q, tapi dapat: %s", tt.password, msg)
			}
		})
	}
}

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name    string
		req     model.RegisterRequest
		wantKey string // field yang seharusnya error, kosong jika valid
	}{
		{
			name:    "username kosong",
			req:     model.RegisterRequest{Username: "", Email: "test@example.com", Password: "Valid99x"},
			wantKey: "username",
		},
		{
			name:    "username terlalu pendek",
			req:     model.RegisterRequest{Username: "ab", Email: "test@example.com", Password: "Valid99x"},
			wantKey: "username",
		},
		{
			name:    "username karakter tidak valid",
			req:     model.RegisterRequest{Username: "user name!", Email: "test@example.com", Password: "Valid99x"},
			wantKey: "username",
		},
		{
			name:    "email tidak valid",
			req:     model.RegisterRequest{Username: "fahmi", Email: "bukan-email", Password: "Valid99x"},
			wantKey: "email",
		},
		{
			name:    "password lemah",
			req:     model.RegisterRequest{Username: "fahmi", Email: "fahmi@example.com", Password: "password1"},
			wantKey: "password",
		},
		{
			name:    "semua valid",
			req:     model.RegisterRequest{Username: "fahmi", Email: "fahmi@example.com", Password: "Kuat99xx"},
			wantKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateRegister(tt.req)
			if tt.wantKey == "" {
				if len(errs) > 0 {
					t.Errorf("tidak diharapkan error, tapi dapat: %v", errs)
				}
			} else {
				if _, ok := errs[tt.wantKey]; !ok {
					t.Errorf("diharapkan error pada field %q, tapi tidak ada. Errors: %v", tt.wantKey, errs)
				}
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	errs := ValidateLogin(model.LoginRequest{Username: "", Password: ""})
	if _, ok := errs["username"]; !ok {
		t.Error("diharapkan error username saat kosong")
	}
	if _, ok := errs["password"]; !ok {
		t.Error("diharapkan error password saat kosong")
	}

	errs2 := ValidateLogin(model.LoginRequest{Username: "fahmi", Password: "apapun"})
	if len(errs2) > 0 {
		t.Errorf("tidak diharapkan error: %v", errs2)
	}
}
