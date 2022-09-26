package fs_observer

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/reactivex/rxgo/v2"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"
	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/sources"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/logz"
)

// Observable a interface
type Observable <-chan rxgo.Item

// NewSubsystem factory function to return a new fs_observer subsystem.
func NewSubsystem(logger logz.FieldLogger, broker cqrs.Broker, observer Observable) *subsystems.Process {
	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": topics.FSObserver,
	}))
	return &subsystems.Process{
		Name:    topics.FSObserver.String(),
		Factory: factory(logger, broker, observer),
	}
}

func factory(
	logger logz.FieldLogger,
	broker cqrs.Broker,
	observer Observable,
) subsystems.Factory {
	return func(wg *sync.WaitGroup, ctx context.Context) func() error {
		return func() error {
			wg.Done()
			for {
				select {
				case <-ctx.Done():
					logger.Info("shutdown recieved")
					return nil
				case fileStream := <-observer:

					if fileStream.E != nil {
						logger.Error(fileStream.E)
						continue
					}

					// we need to validate the fileStream value before we try to cast it.
					// if is nil we should ignore this message and move forward
					if fileStream.V == nil {
						continue
					}

					msg := messages.FileSystemObserverFileChanged{
						Files: fileStream.V.([]*fs.File),
					}

					logger.Debugf("files changed %d", len(msg.Files))

					if err := broker.Publish(topics.FSObserverEvents, cqrs.NewDefaultEnvelope(
						sources.FSObserver,
						events.FSObserverFileChangedType,
						cqrs.WithData(cqrs.ApplicationJSON, msg),
					)); err != nil {
						logger.Error(errors.Wrap(err, "failed to publish change notification"))
					}
				}
			}
		}
	}
}
