package http_handlers

import (
	"github.com/gofiber/fiber/v2"
)

func NewHealthListener() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(map[string]interface{}{
			"ok": true,
		})
	}
}
