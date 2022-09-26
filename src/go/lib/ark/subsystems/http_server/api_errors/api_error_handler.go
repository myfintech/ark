package api_errors

import (
	"github.com/gofiber/fiber/v2"
	"github.com/myfintech/ark/src/go/lib/logz"
)

// NewErrorHandler returns a fiber compatible error handler
func NewErrorHandler(logger logz.FieldLogger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		logger.WithFields(logz.Fields{
			"user-agent":   c.Get("User-Agent"),
			"method":       c.Method(),
			"original-url": c.OriginalURL(),
		}).Error(err)
		switch e := err.(type) {
		case APIError:
			return c.Status(e.Status).JSON(e)
		case *fiber.Error:
			return c.Status(e.Code).JSON(e)
		default:
			return c.Status(InternalServerError.Status).JSON(InternalServerError)
		}
	}
}
