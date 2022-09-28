package http_handlers

import (
	"encoding/json"
	"net/http"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/commands"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"
)

// NewRunHandler returns an
func NewRunHandler(store ark.Store, _ logz.FieldLogger, broker cqrs.Broker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		subscriptionId := uuid.New()
		var cmd messages.GraphRunnerExecuteCommand

		if err := json.Unmarshal(c.Body(), &cmd); err != nil {
			return api_errors.InternalServerError.
				WithErr(err)
		}

		for _, key := range cmd.TargetKeys {
			_, err := store.GetTargetByKey(key)
			if err != nil {
				// FIXME(rckgomz): we need to have better error types
				return api_errors.InternalServerError.
					WithErr(err)
			}
		}

		message := cqrs.NewDefaultEnvelope(
			commands.GraphRunnerExecuteType,
			cqrs.WithSource(cqrs.RouteKey(c.OriginalURL())),
			cqrs.WithSubject(cqrs.RouteKey(subscriptionId.String())),
			cqrs.WithData(cqrs.ApplicationJSON, cmd),
		)

		if message.Error != nil {
			return api_errors.InternalServerError.
				WithErr(message.Error)
		}

		if err := broker.Publish(topics.GraphRunnerCommands, message); err != nil {
			return api_errors.InternalServerError.
				WithErr(err)
		}

		return c.Status(http.StatusAccepted).JSON(messages.GraphRunnerExecuteCommandResponse{
			SubscriptionId: subscriptionId.String(),
		})
	}
}
