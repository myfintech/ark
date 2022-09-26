package http_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/myfintech/ark/src/go/lib/ark"
	api_errors2 "github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"
)

func NewListTargetsHandler(store ark.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {

		if targetKey := c.Query("targetKey", ""); targetKey != "" {

			target, err := store.GetTargetByKey(targetKey)
			if err != nil {
				return api_errors2.NotFoundError.
					WithErr(err)
			}

			return c.JSON(target)
		}

		targets, err := store.GetTargets()
		if err != nil {
			return api_errors2.InternalServerError.
				WithErr(err)
		}

		return c.JSON(targets)
	}
}
