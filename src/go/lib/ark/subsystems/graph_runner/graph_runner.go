package graph_runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/myfintech/ark/src/go/lib/kube"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/commands"

	"github.com/docker/cli/cli/command"
	"github.com/myfintech/ark/src/go/lib/container"

	"github.com/myfintech/ark/src/go/lib/logz/transports"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/sources"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/graph"
	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/logz"
)

// NewSubsystem factory function to return a new graph_runner subsystem.
func NewSubsystem(store ark.Store, logger logz.FieldLogger, sharedClients shared_clients.Container, broker cqrs.Broker) *subsystems.Process {
	// logger = logger.Child(logz.WithFields(logz.Fields{
	// 	"system": topics.GraphRunner.String(),
	// }))
	return &subsystems.Process{
		Name: topics.GraphRunner.String(),
		Factory: subsystems.Reactor(
			topics.GraphRunnerCommands,
			broker,
			logger,
			newOnMessageFunc(store, sharedClients, broker, logger),
			newOnMessageErrFunc(broker),
			nil,
		),
	}
}

func newOnMessageErrFunc(broker cqrs.Broker) cqrs.OnMessageErrorFunc {
	return func(ctx context.Context, msg cqrs.Envelope, err error) error {
		return broker.Publish(topics.GraphRunnerEvents, cqrs.NewDefaultEnvelope(
			sources.GraphRunner,
			events.GraphRunnerFailedType,
			cqrs.WithSubject(msg.SubjectKey()),
			cqrs.WithData(cqrs.TextPlain, err.Error()),
		))
	}
}

type runnerState struct {
	mutex      sync.Mutex
	executions map[string]context.CancelFunc
}

func (r *runnerState) store(id string, ctx context.Context) context.Context {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	subCtx, cancel := context.WithCancel(ctx)
	r.executions[id] = cancel
	return subCtx
}

func (r *runnerState) stop(id string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if stop, ok := r.executions[id]; ok {
		stop()
		delete(r.executions, id)
	}
}

func newRunnerState() *runnerState {
	return &runnerState{executions: make(map[string]context.CancelFunc)}
}

func newOnMessageFunc(store ark.Store, sharedClients shared_clients.Container, broker cqrs.Broker, logger logz.FieldLogger) cqrs.OnMessageFunc {
	state := newRunnerState()
	return func(ctx context.Context, msg cqrs.Envelope) error {
		if msg.Error != nil {
			return errors.Wrap(msg.Error, "failed to deserialize incoming envelope")
		}

		logger.Debugf("command recieved from %s", msg.Subject())
		if msg.TypeKey() == commands.GraphRunnerCancel {
			err := errors.Errorf("%s canceled build %s", msg.Source(), msg.Subject())
			logger.Warn(err.Error())
			state.stop(msg.Subject())
			return err
		}

		cmd := new(messages.GraphRunnerExecuteCommand)
		if err := msg.DataAs(cmd); err != nil {
			return errors.Wrap(err, "failed to unmarshal the incoming command")
		}

		logger.Debugf("saving context for future cancellation %s", msg.Subject())
		ctx = state.store(msg.Subject(), ctx)
		defer state.stop(msg.Subject())

		if err := broker.Publish(topics.GraphRunnerEvents, cqrs.NewDefaultEnvelope(
			sources.GraphRunner,
			events.GraphRunnerStartedType,
			cqrs.WithSubject(cqrs.RouteKey(msg.Subject())),
		)); err != nil {
			return err
		}

		ctxLogger := logger.Child(
			logz.WithMux(transports.SuggestedLogFileWriter(
				fmt.Sprintf("ark/graph/%s", msg.Subject()),
				"run.log"),
			),
			logz.WithFields(logz.Fields{
				"system":         topics.GraphWalker.String(),
				"subscriptionID": msg.Subject(),
				"target_key":     cmd.TargetKeys[0],
			}),
		)

		if w, ok := ctxLogger.(*logz.Writer); ok {
			defer w.Close()

			if w.InitError() != nil {
				return errors.Wrap(w.InitError(), "failed to initialize child logger")
			}
		}

		dockerClient, err := container.NewDockerClient([]command.DockerCliOption{
			command.WithOutputStream(ctxLogger),
			command.WithErrorStream(ctxLogger),
			command.WithContentTrust(true),
		}...)
		if err != nil {
			return err
		}

		sharedClients.K8s.NamespaceOverride = kube.NormalizeNamespace(cmd.K8sNamespace)
		sharedClients.K8s.OutputWriter = ctxLogger

		sharedClients.Docker = *dockerClient
		sharedClients.Docker.OutputWriter = ctxLogger

		ctxLogger.Debug("starting graph execution")
		err = graph.Execute(graph.ExecuteOptions{
			Ctx:                         ctx,
			Store:                       store,
			SharedClients:               &sharedClients,
			RootTargetKey:               cmd.TargetKeys[0],
			Broker:                      broker,
			SubscriptionID:              msg.Subject(),
			ForceExecution:              cmd.ForceBuild,
			PushArtifactsAfterExecution: cmd.PushAfterBuild,
			SkipFilters:                 cmd.SkipFilters,
			K8sNamespace:                cmd.K8sNamespace,
			K8sContext:                  cmd.K8sContext,
			Logger:                      ctxLogger,
			MaxConcurrency:              cmd.MaxConcurrency,
		})
		ctxLogger.Debug("graph execution completed")

		if err != nil {
			return errors.Wrapf(err, "could not execute the graph for the given keys %s", cmd.TargetKeys)
		}

		return broker.Publish(topics.GraphRunnerEvents, cqrs.NewDefaultEnvelope(
			sources.GraphRunner,
			events.GraphRunnerSuccessType,
			cqrs.WithSubject(cqrs.RouteKey(msg.Subject())),
		))
	}
}
