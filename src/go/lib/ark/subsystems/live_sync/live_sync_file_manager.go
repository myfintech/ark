package live_sync

import (
	"bytes"
	"context"
	"fmt"
	"math"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/ark/components/entrypoint"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/logz"
)

func NewFSSync(broker cqrs.Broker, logger logz.FieldLogger, manager *ConnectionManager, config workspace.Config) *subsystems.Process {
	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": "live.sync.file.manager",
	}))
	return &subsystems.Process{
		Name:    topics.LiveSyncFSManager.String(),
		Factory: subsystems.Reactor(topics.FSObserverEvents, broker, logger, newFSSyncMessageHandler(manager, broker, config, logger), newFSSyncErrorHandler(logger), nil),
	}
}

func newFSSyncMessageHandler(manager *ConnectionManager, _ cqrs.Broker, config workspace.Config, logger logz.FieldLogger) cqrs.OnMessageFunc {
	logger.Info("ready")
	return func(ctx context.Context, msg cqrs.Envelope) error {
		if manager.Len() < 1 {
			logger.Debug("no live sync connections, skipping")
			return nil
		}

		if msg.TypeKey() != events.FSObserverFileChanged {
			return nil
		}

		if msg.Error != nil {
			return msg.Error
		}

		var change messages.FileSystemObserverFileChanged

		if err := msg.DataAs(&change); err != nil {
			return err
		}

		if len(change.Files) == 0 {
			return nil
		}

		var filePaths []string
		var entrypointNotification = new(entrypoint.FileChangeNotification)

		writer := new(bytes.Buffer)
		for _, f := range change.Files {
			entrypointNotification.Files = append(entrypointNotification.Files, &entrypoint.File{
				Name:          f.Name,
				Exists:        f.Exists,
				New:           f.New,
				Type:          f.Type,
				Hash:          f.Hash,
				SymlinkTarget: f.SymlinkTarget,
				RelName:       f.RelName,
			})

			filePaths = append(filePaths, f.Name)
			logger.Debugf(f.Name)
		}

		err := fs.GzipTarFiles(filePaths, config.Root(), writer, nil)
		if err != nil {
			return err
		}

		logger.Infof("compressed changes into %s", hsize(float64(writer.Len())))
		entrypointNotification.Archive = writer.Bytes()

		for _, connection := range manager.GetAllConnections() {
			go sendFileChange(connection, entrypointNotification)
		}

		return nil
	}
}

func sendFileChange(connection entrypoint.Sync_StreamFileChangeClient, entrypointNotification *entrypoint.FileChangeNotification) {
	if connErr := connection.Send(entrypointNotification); connErr != nil {
		log.Error(connErr)
	}
}

func newFSSyncErrorHandler(logger logz.FieldLogger) func(ctx context.Context, msg cqrs.Envelope, err error) error {
	return func(ctx context.Context, msg cqrs.Envelope, err error) error {
		logger.Error(err)
		return err
	}
}

func hsize(num float64) string {
	suffix := "B"
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(num) < 1024.0 {
			return fmt.Sprintf("%3.1f%s%s", num, unit, suffix)
		}
		num /= 1024.0
	}

	return fmt.Sprintf("%.1f%s%s", num, "Yi", suffix)
}
