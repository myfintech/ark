package events

import (
	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
)

var (
	// GraphWalkerStarted event
	GraphWalkerStarted = topics.GraphWalkerEvents.With("started")

	// GraphWalkerStartedType
	GraphWalkerStartedType = cqrs.WithType(GraphWalkerStarted)

	// GraphWalkerFailed event
	GraphWalkerFailed = topics.GraphWalkerEvents.With("failed")

	// GraphWalkerFailedType
	GraphWalkerFailedType = cqrs.WithType(GraphWalkerFailed)

	// GraphWalkerDerivationComputed
	GraphWalkerDerivationComputed = topics.GraphWalkerEvents.With("derivation.computed")

	// GraphWalkerDerivationComputedType
	GraphWalkerDerivationComputedType = cqrs.WithType(GraphWalkerDerivationComputed)

	// GraphWalkerActionCached
	GraphWalkerActionCached = topics.GraphWalkerEvents.With("action.cached")

	// GraphWalkerActionCachedType
	GraphWalkerActionCachedType = cqrs.WithType(GraphWalkerActionCached)

	// GraphWalkerActionStarted
	GraphWalkerActionStarted = topics.GraphWalkerEvents.With("action.started")

	// GraphWalkerActionStartedType
	GraphWalkerActionStartedType = cqrs.WithType(GraphWalkerActionStarted)

	// GraphWalkerActionSuccess
	GraphWalkerActionSuccess = topics.GraphWalkerEvents.With("action.success")

	// GraphWalkerActionSuccessType
	GraphWalkerActionSuccessType = cqrs.WithType(GraphWalkerActionSuccess)

	// GraphWalkerArtifactPushStarted
	GraphWalkerArtifactPushStarted = topics.GraphWalkerEvents.With("artifact.push.started")

	// GraphWalkerArtifactPushStartedType
	GraphWalkerArtifactPushStartedType = cqrs.WithType(GraphWalkerArtifactPushStarted)

	// GraphRunnerStarted
	GraphRunnerStarted = topics.GraphRunnerEvents.With("started")

	// GraphRunnerStartedType
	GraphRunnerStartedType = cqrs.WithType(GraphRunnerStarted)

	// GraphRunnerFailed
	GraphRunnerFailed = topics.GraphRunnerEvents.With("failed")

	// GraphRunnerFailedType
	GraphRunnerFailedType = cqrs.WithType(GraphRunnerFailed)

	// GraphRunnerSuccess
	GraphRunnerSuccess = topics.GraphRunnerEvents.With("success")

	// GraphRunnerSuccessType
	GraphRunnerSuccessType = cqrs.WithType(GraphRunnerSuccess)

	// PortBinderSuccess
	PortBinderSuccess = topics.PortBinderEvents.With("success")
	// PortBinderSuccessType
	PortBinderSuccessType = cqrs.WithType(PortBinderSuccess)

	// PortBinderFailed
	PortBinderFailed = topics.PortBinderEvents.With("failed")
	// PortBinderFailedType
	PortBinderFailedType = cqrs.WithType(PortBinderFailed)
)

var (
	// FSObserverFileChanged a event
	FSObserverFileChanged = topics.FSObserverEvents.With("file.changed")
	// FSObserverFileChangedType event type
	FSObserverFileChangedType = cqrs.WithType(FSObserverFileChanged)
)

var (
	// K8sEchoPodChanged a event
	K8sEchoPodChanged = topics.K8sEchoEvents.With("pod.changed")
	// K8sEchoPodChangedType event type
	K8sEchoPodChangedType = cqrs.WithType(K8sEchoPodChanged)
)
