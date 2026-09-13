package helper

import (
	"github.com/gofiber/fiber/v2"

	"modul4/app/model"
)

// LocalsAuthUser adalah kunci untuk menyimpan/membaca identitas user di Fiber context
const LocalsAuthUser = "authUser"

// CurrentUser membaca identitas user yang sudah diset oleh middleware RequireAuth
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}
