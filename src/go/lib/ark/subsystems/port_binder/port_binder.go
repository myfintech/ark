package port_binder

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/sources"
	v1 "k8s.io/api/core/v1"

	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/kube/portbinder"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"

	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
)

// NewSubsystem factory function to return a new graph_runner subsystem.
func NewSubsystem(
	broker cqrs.Broker,
	logger logz.FieldLogger,
	client kube.Client,
) *subsystems.Process {
	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": topics.PortBinder.String(),
	}))

	// client.OutputWriter = logger
	return &subsystems.Process{
		Name: topics.PortBinder.String(),
		Factory: subsystems.Reactor(
			topics.PortBinderCommands,
			broker,
			logger,
			newOnMessageFunc(broker, logger, client),
			newOnMessageErrFunc(broker, logger),
			nil,
		),
	}
}

func K8sEchoHandler(
	broker cqrs.Broker,
	logger logz.FieldLogger,
	client kube.Client,
) *subsystems.Process {
	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": topics.PortBinder.String(),
	}))

	// client.OutputWriter = logger
	return &subsystems.Process{
		Name: topics.K8sEchoEvents.String(),
		Factory: subsystems.Reactor(
			topics.K8sEchoEvents,
			broker,
			logger,
			newK8sEchoOnMessageFunc(broker, logger, client),
			newOnMessageErrFunc(broker, logger),
			nil,
		),
	}
}

func newOnMessageErrFunc(broker cqrs.Broker, logger logz.FieldLogger) cqrs.OnMessageErrorFunc {
	return func(ctx context.Context, msg cqrs.Envelope, err error) error {
		logger.Error(err)
		return broker.Publish(topics.PortBinderEvents, cqrs.NewDefaultEnvelope(
			sources.PortBinder,
			events.PortBinderFailedType,
			cqrs.WithSubject(cqrs.RouteKey(msg.Subject())),
			cqrs.WithData(cqrs.TextPlain, err.Error()),
		))
	}
}

type portBinderState struct {
	mutex  sync.Mutex
	ports  map[string]*kube.ForwardingOptions
	logger logz.FieldLogger
}

func newPortBinderState(logger logz.FieldLogger) *portBinderState {
	return &portBinderState{
		ports:  make(map[string]*kube.ForwardingOptions),
		logger: logger,
		// logger: logger.Child(logz.WithFields(logz.Fields{
		//	"system": "port.binder.state",
		// })),
	}
}

func (s *portBinderState) unbindExistingPorts(podName string, command *portbinder.BindPortCommand) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.logger.Debugf("attempting to unbind port state %s", command.PortMap.ToPairs())

	for _, binding := range command.PortMap {
		if port, ok := s.ports[binding.HostPort]; ok {
			s.logger.Debugf("attempting to close %s for pod %s", binding.HostPort, podName)
			port.Stop()

			select {
			case <-port.DoneChannel:
				s.logger.Infof(
					"successfully unbound port state %s for pod %s",
					command.PortMap.ToPairs(),
					podName,
				)
				delete(s.ports, binding.HostPort)
			case <-time.After(time.Second * 20): // TODO: maybe extract this to setting.json
				panic(errors.Errorf("fail to unbind port %s possible deadlock", binding.HostPort))
			}
		}
	}
}

func (s *portBinderState) addBindings(
	podName string,
	command *portbinder.BindPortCommand,
	opts *kube.ForwardingOptions,
) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debugf(
		"adding port forwarding state %s for pod %s",
		command.PortMap.ToPairs(),
		podName,
	)
	for _, binding := range command.PortMap {
		s.ports[binding.HostPort] = opts
	}
}

func newOnMessageFunc(
	broker cqrs.Broker,
	logger logz.FieldLogger,
	client kube.Client,
) cqrs.OnMessageFunc {
	logger.Info("ready")
	state := newPortBinderState(logger)
	return func(ctx context.Context, msg cqrs.Envelope) error {
		if msg.Error != nil {
			return errors.Wrap(msg.Error, "failed to deserialize incoming envelope")
		}

		command := new(portbinder.BindPortCommand)
		if err := msg.DataAs(command); err != nil {
			return errors.Wrap(err, "failed to unmarshal the incoming port binder")
		}

		state.unbindExistingPorts("dont know yet", command)

		pods, err := kube.GetPodsByLabel(
			client,
			command.Selector.Namespace,
			command.Selector.LabelKey,
			command.Selector.LabelValue)
		if err != nil {
			return err
		}

		if len(pods) == 0 {
			return errors.Errorf(
				"no pods found in namespace %s by %s=%s",
				command.Selector.Namespace,
				command.Selector.LabelKey,
				command.Selector.LabelValue,
			)
		}

		forwardOptions := &kube.ForwardingOptions{
			Namespace:    command.Selector.Namespace,
			Pod:          pods[0],
			Client:       client,
			Ports:        command.PortMap.ToPairs(),
			StopChannel:  make(chan struct{}),
			ReadyChannel: make(chan struct{}),
			DoneChannel:  make(chan error, 1),
		}

		state.addBindings("dont know yet", command, forwardOptions)

		go forwardPorts(forwardOptions, logger)
		go watchContextAndCancelForwarding(ctx, forwardOptions)

		select {
		case <-forwardOptions.ReadyChannel:
			logger.Infof("ports successfully bound %s", command.PortMap.ToPairs())
		case err = <-forwardOptions.DoneChannel:
			logger.Errorf("failed to bind ports %s %v", command.PortMap.ToPairs(), err)
			return err
		}

		return broker.Publish(topics.PortBinderEvents, cqrs.NewDefaultEnvelope(
			sources.PortBinder,
			events.PortBinderSuccessType,
			cqrs.WithSubject(cqrs.RouteKey(msg.Subject())),
			cqrs.WithData(cqrs.ApplicationJSON, command),
		))
	}
}

func newK8sEchoOnMessageFunc(
	broker cqrs.Broker,
	logger logz.FieldLogger,
	client kube.Client,
) cqrs.OnMessageFunc {
	logger.Info("ready")
	state := newPortBinderState(logger)
	return func(ctx context.Context, msg cqrs.Envelope) error {
		if msg.Error != nil {
			return errors.Wrap(msg.Error, "failed to deserialize incoming envelope")
		}

		event := new(messages.K8sEchoResourceChanged)
		if err := msg.DataAs(event); err != nil {
			return errors.Wrap(err, "failed to unmarshal the incoming port binder")
		}

		// skip port binding if pod doesn't have port binding enabled
		if event.Labels["ark.port.binding.enabled"] != "true" {
			logger.Debugf(
				"skipping port binding for pod: %s, label ark.port.binding.enabled is false",
				event.Name,
			)
			return nil
		}

		rawPortMap := []byte(event.Annotations["ark.port.binding"])

		// skip port binding if ark.port.binding annotation is empty
		if rawPortMap == nil || len(rawPortMap) <= 0 {
			logger.Debugf(
				"skipping port binding for pod: %s, annotation for ark.port.binding is empty",
				event.Name,
			)
			return nil
		}

		portMap := new(portbinder.PortMap)
		if err := json.Unmarshal(rawPortMap, &portMap); err != nil {
			return errors.Wrapf(err, "failed to unmarshal port map information %v", portMap)
		}

		// skip port binding if portMap is empty
		if len(*portMap) <= 0 {
			logger.Debugf(
				"skipping port binding for pod: %s, config for port binding not found",
				event.Name,
			)
			return nil
		}

		const labelKey = "ark.target.key"
		command := &portbinder.BindPortCommand{
			Selector: portbinder.Selector{
				Namespace:  event.Namespace,
				LabelKey:   labelKey,
				LabelValue: event.Labels[labelKey],
				Type:       "pod",
			},
			PortMap: *portMap,
		}

		var pod *v1.Pod
		if err := json.Unmarshal([]byte(event.Raw), &pod); err != nil {
			return errors.Wrap(err, "failed to unmarshal pod")
		}

		// skip if the pod has been mark for deletion since the new pod
		// will unbind and bind its ports
		if pod.DeletionTimestamp != nil {
			logger.Debugf("skipping port binding for pod: %s, marked for deletion", event.Name)
			return nil
		}

		// skip if pod is not in a running state due port bidning will fail
		if pod.Status.Phase != v1.PodRunning {
			logger.Debugf("skipping port binding for pod: %s is not running", event.Name)
			return nil
		}

		forwardOptions := &kube.ForwardingOptions{
			Namespace:    command.Selector.Namespace,
			Pod:          *pod,
			Client:       client,
			Ports:        command.PortMap.ToPairs(),
			StopChannel:  make(chan struct{}),
			ReadyChannel: make(chan struct{}),
			DoneChannel:  make(chan error, 1),
		}

		state.unbindExistingPorts(event.Name, command)
		state.addBindings(event.Name, command, forwardOptions)

		go forwardPorts(forwardOptions, logger)
		go watchContextAndCancelForwarding(ctx, forwardOptions)

		select {
		case <-forwardOptions.ReadyChannel:
			logger.Infof(
				"ports successfully bound %s for pod: %s",
				command.PortMap.ToPairs(),
				event.Name,
			)
		case err := <-forwardOptions.DoneChannel:
			logger.Errorf(
				"failed to bind ports %s for pod: %s, %v",
				command.PortMap.ToPairs(),
				event.Name,
				err,
			)
			return err
		}

		return broker.Publish(topics.PortBinderEvents, cqrs.NewDefaultEnvelope(
			sources.PortBinder,
			events.PortBinderSuccessType,
			cqrs.WithSubject(cqrs.RouteKey(msg.Subject())),
			cqrs.WithData(cqrs.ApplicationJSON, command),
		))
	}
}

func watchContextAndCancelForwarding(ctx context.Context, opts *kube.ForwardingOptions) {
	select {
	case <-ctx.Done():
		opts.Stop()
	}
}

func forwardPorts(opts *kube.ForwardingOptions, logger logz.FieldLogger) {
	defer close(opts.DoneChannel)
	logger.Infof("connecting to pod %s", opts.Pod.Name)
	opts.DoneChannel <- kube.PortForward(*opts)
}
