package helper

import (
	"github.com/gofiber/fiber/v2"

	"modul6/app/model"
)

// LocalsAuthUser adalah kunci untuk menyimpan/membaca identitas user di Fiber context
const LocalsAuthUser = "authUser"

// CurrentUser membaca identitas user yang sudah diset oleh middleware RequireAuth
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}

// RequestID membaca request id dari c.Locals("requestid")
func RequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("requestid").(string); ok {
		return id
	}
	return ""
}
