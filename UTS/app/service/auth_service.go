package service

import (
	"context"
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type AuthService struct {
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
	jwtManager  *helper.JWTManager
}

func NewAuthService(
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	jwtManager *helper.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		studentRepo: studentRepo,
		jwtManager:  jwtManager,
	}
}

// Endpoint 1: POST /api/v1/auth/login
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx := context.Background()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format request JSON tidak valid")
	}

	// Validasi input
	valErrors := make(map[string][]string)

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		valErrors["email"] = append(valErrors["email"], "Email wajib diisi")
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		valErrors["email"] = append(valErrors["email"], "Format email tidak valid")
	}

	if req.Password == "" {
		valErrors["password"] = append(valErrors["password"], "Password wajib diisi")
	} else if len(req.Password) < 8 {
		valErrors["password"] = append(valErrors["password"], "Password minimal 8 karakter")
	}

	if len(valErrors) > 0 {
		return helper.Validation(valErrors)
	}

	// Cari user berdasarkan email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return helper.Internal(err)
	}
	if user == nil {
		return helper.Unauthorized("Email atau password salah")
	}

	// Cek password hash
	if !helper.ComparePassword(user.Password, req.Password) {
		return helper.Unauthorized("Email atau password salah")
	}

	// Jika role mahasiswa, pastikan tidak sedang di-soft delete
	if user.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return helper.Internal(err)
		}
		if student == nil || student.DeletedAt != nil {
			return helper.Unauthorized("Akun mahasiswa telah dinonaktifkan")
		}
	}

	// Generate JWT Access Token
	token, expiresIn, err := s.jwtManager.Generate(helper.AuthUser{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	})
	if err != nil {
		return helper.Internal(err)
	}

	res := model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User: model.UserSummary{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", res)
}

// Endpoint 2: GET /api/v1/auth/me
func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx := context.Background()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	user, err := s.userRepo.FindByID(ctx, authUser.UserID)
	if err != nil {
		return helper.Internal(err)
	}
	if user == nil {
		return helper.NotFound("Pengguna tidak ditemukan")
	}

	res := model.AuthMeResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	// Jika mahasiswa, sertakan data students (nim, nama, prodi, angkatan)
	if user.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return helper.Internal(err)
		}
		if student != nil {
			res.Student = &model.StudentMeSummary{
				ID:       student.ID,
				NIM:      student.NIM,
				Nama:     student.Nama,
				Prodi:    student.Prodi,
				Angkatan: student.Angkatan,
			}
		}
	}

	return helper.Success(c, fiber.StatusOK, "Profil pengguna berhasil diambil", res)
}
