package subsystems

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
)

// Reactor subscribes using the given topic and processes messages using the provided handlers
func Reactor(topic cqrs.RouteKey, broker cqrs.Broker, logger logz.FieldLogger, onMessage cqrs.OnMessageFunc, onError cqrs.OnMessageErrorFunc, ackDeadline *time.Duration) Factory {
	return func(wg *sync.WaitGroup, ctx context.Context) func() error {
		return func() error {
			stream, err := broker.Subscribe(ctx, topic, ackDeadline)
			if err != nil {
				return err
			}

			wg.Done()

			for {
				select {
				case msg := <-stream:
					go reactorProcessMessage(onMessage, ctx, msg, onError, logger)
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}

func reactorProcessMessage(onMessage cqrs.OnMessageFunc, ctx context.Context, msg cqrs.Envelope, onError cqrs.OnMessageErrorFunc, logger logz.FieldLogger) {
	msgErr := onMessage(ctx, msg)
	if msgErr == nil {
		msg.Ack()
		return
	}

	logger.Debugf("reactor received error from onMessageHandler %v", msgErr)

	if errors.As(msgErr, new(cqrs.RetryableError)) {
		msg.Reject(msgErr)
	} else {
		msg.Ack()
	}

	unrecoverableErr := onError(ctx, msg, msgErr)
	if unrecoverableErr != nil {
		logger.Error(unrecoverableErr)
	}
}
