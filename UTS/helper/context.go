package helper

import "github.com/gofiber/fiber/v2"

const LocalsAuthUser = "auth_user"

func CurrentUser(c *fiber.Ctx) (*AuthUser, bool) {
	val := c.Locals(LocalsAuthUser)
	if val == nil {
		return nil, false
	}
	user, ok := val.(*AuthUser)
	return user, ok
}
