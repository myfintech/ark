package topics

import "github.com/myfintech/ark/src/go/lib/ark/cqrs"

var (
	// GraphRunner a system level topic
	GraphRunner = cqrs.RouteKey("graph.runner")

	// GraphRunnerCommands the GraphRunner commands topic
	GraphRunnerCommands = GraphRunner.With("commands")

	// GraphRunnerEvents the GraphRunner events topic
	GraphRunnerEvents = GraphRunner.With("events")

	// GraphWalker a system level topic
	GraphWalker = cqrs.RouteKey("graph.walker")

	// GraphWalkerEvents the GraphWalker events topic
	GraphWalkerEvents = GraphWalker.With("events")

	// HTTPServer a system level topic
	HTTPServer = cqrs.RouteKey("http.server")

	// EmbeddedBroker a system level topic
	EmbeddedBroker = cqrs.RouteKey("embedded.broker")

	// LiveSyncConnectionManager a system level topic
	LiveSyncConnectionManager = cqrs.RouteKey("live.sync.connection.manager")

	// LiveSyncFSManager a system level topic
	LiveSyncFSManager = cqrs.RouteKey("live.sync.fs.manager")

	// LiveSyncConnectionManagerEvents a system level topic
	LiveSyncConnectionManagerEvents = LiveSyncConnectionManager.With("events")

	// PortBinder a system level topic
	PortBinder = cqrs.RouteKey("port.binder")

	// PortBinderCommands a system level topic
	PortBinderCommands = PortBinder.With("commands")

	// PortBinderEvents a system level topic
	PortBinderEvents = PortBinder.With("events")
)

var (
	// FSObserver a system level topic
	FSObserver = cqrs.RouteKey("fs.observer")

	// FSObserverEvents the FSObserver events topic
	FSObserverEvents = FSObserver.With("events")
)

var (
	// K8sEcho a system level topic
	K8sEcho = cqrs.RouteKey("k8s.echo")

	// K8sEchoEvents a system level topic
	K8sEchoEvents = K8sEcho.With("events")
)
