package http_handlers

import (
	"encoding/json"

	api_errors2 "github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"

	"github.com/gofiber/fiber/v2"
	"github.com/myfintech/ark/src/go/lib/ark"
)

func NewConnectTargetHandler(store ark.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var edge ark.GraphEdge

		if err := json.Unmarshal(c.Body(), &edge); err != nil {
			return api_errors2.InternalServerError.
				WithErr(err)
		}

		if err := store.ConnectTargets(edge); err != nil {
			return api_errors2.InternalServerError.
				WithErr(err)
		}

		return c.JSON(edge)
	}
}
