package http_handlers

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/gofiber/fiber/v2"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/derivation"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"
)

// NewAddTargetHandler returns an
func NewAddTargetHandler(store ark.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rawTarget ark.RawTarget

		if err := json.Unmarshal(c.Body(), &rawTarget); err != nil {
			return api_errors.InternalServerError.
				WithErr(err)
		}

		target, err := derivation.TargetFromRawTarget(rawTarget)
		if err != nil {
			return api_errors.InternalServerError.
				WithErr(err)
		}

		if err = target.Validate(); err != nil {
			return api_errors.InternalServerError.
				WithErr(errors.Errorf("%s failed validation %v", target.Key(), err))
		}

		artifact, err := store.AddTarget(rawTarget)
		if err != nil {
			return api_errors.InternalServerError.
				WithErr(err)
		}

		return c.JSON(artifact)
	}
}
