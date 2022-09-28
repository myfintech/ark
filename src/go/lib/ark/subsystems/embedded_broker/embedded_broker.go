package embedded_broker

import (
	"context"
	"sync"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/protocols/nats"
	"github.com/nats-io/nats-server/v2/server"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/logz"
)

func NewSubsystem(brokerType, _ string, logger logz.FieldLogger, natsd *server.Server, broker cqrs.Broker) *subsystems.Process {
	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": topics.EmbeddedBroker.String(),
		"type":   brokerType,
	}))
	return &subsystems.Process{
		Name: topics.EmbeddedBroker.String(),
		Factory: func(wg *sync.WaitGroup, ctx context.Context) func() error {
			return func() error {

				if brokerType != "nats-embedded" {
					return errors.Errorf("requested broker type %s is not supported", brokerType)
				}

				if natsd == nil {
					s, err := nats.RunServer(nil)
					if err != nil {
						return err
					}
					natsd = s
				}

				// subscribe to ALL events
				observer, err := broker.Subscribe(ctx, ">", nil)
				if err != nil {
					return err
				}

				go func() {
					for envelope := range observer {
						logger.WithFields(logz.Fields{
							"id":      envelope.ID(),
							"time":    envelope.Time(),
							"source":  envelope.Source(),
							"subject": envelope.Subject(),
							"type":    envelope.Type(),
						}).Trace(envelope.Data())
					}
				}()

				wg.Done()

				logger.Infof("listening on %s", natsd.ClientURL())

				<-ctx.Done()
				logger.Info("recieved shutdown signal")
				natsd.Shutdown()
				return nil
			}
		},
	}
}
