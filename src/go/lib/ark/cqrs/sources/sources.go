package sources

import (
	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
)

var (
	// GraphWalkerSource a message source derived from its topic
	GraphWalkerSource = cqrs.WithSource(topics.GraphWalker)

	// GraphRunner a message source derived from its topic
	GraphRunner = cqrs.WithSource(topics.GraphRunner)
)

var (
	// FSObserver a message source derived from its topic
	FSObserver = cqrs.WithSource(topics.FSObserver)
)

var (
	// PortBinder a message source for port binder events
	PortBinder = cqrs.WithSource(topics.PortBinder)
)

var (
	// K8sEcho a message source for the K8s echo
	K8sEcho = cqrs.WithSource(topics.K8sEcho)
)
