package middleware

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/helper"
)

// RequireRole memeriksa apakah user memiliki salah satu role yang diizinkan
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("Belum terautentikasi")
		}

		if _, granted := allowed[user.Role]; !granted {
			return helper.Forbidden("Akses ditolak: role Anda tidak memiliki izin untuk aksi ini")
		}

		return c.Next()
	}
}
