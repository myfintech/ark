package graph

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/myfintech/ark/src/go/lib/container"
	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/logz/transports"

	"github.com/myfintech/ark/src/go/lib/ark/targets/deploy"

	"golang.org/x/sync/semaphore"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/sources"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/derivation"
	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"
	"github.com/myfintech/ark/src/go/lib/dag"
	"github.com/pkg/errors"
)

// ExecuteOptions parameters to drive the graph engine
type ExecuteOptions struct {
	Ctx                         context.Context
	Store                       ark.Store
	SharedClients               *shared_clients.Container
	RootTargetKey               string
	PushArtifactsAfterExecution bool
	SubscriptionID              string
	Broker                      cqrs.Broker
	SkipFilters                 []string
	K8sNamespace                string
	K8sContext                  string
	ForceExecution              bool
	Logger                      logz.FieldLogger
	MaxConcurrency              int
}

var topic = topics.GraphWalkerEvents

// Execute executes a parallel walk of the graph derived from the supplied ark.Store
// If ExecuteOptions.MaxConcurrency is 0 it will be set to runtime.GOMAXPROCS(0)
func Execute(opts ExecuteOptions) error {
	if opts.MaxConcurrency == 0 {
		opts.MaxConcurrency = runtime.GOMAXPROCS(0)
	}

	graph, err := opts.Store.GetGraph()
	if err != nil {
		return err
	}

	rootVertex, err := opts.Store.GetTargetByKey(opts.RootTargetKey)
	if err != nil {
		return err
	}

	graph = graph.Isolate(rootVertex)
	if err = graph.WalkWithErr(validationWalk(opts)); err != nil {
		return err
	}

	if err = opts.Broker.Publish(topic, cqrs.NewDefaultEnvelope(
		sources.GraphWalkerSource,
		events.GraphWalkerStartedType,
		cqrs.WithSubject(cqrs.RouteKey(opts.SubscriptionID)),
		cqrs.WithData(cqrs.ApplicationJSON, graph),
	)); err != nil {
		return err
	}

	return graph.WalkWithErr(newExecutionWalkFunc(opts))
}

func validationWalk(opts ExecuteOptions) func(vertex dag.Vertex) (err error) {
	return func(vertex dag.Vertex) (err error) {
		var rawTarget ark.RawTarget

		// the graph isolation function produces a graph of pointers
		// this allows us to support static copies and pointer data in the graph
		switch t := vertex.(type) {
		case ark.RawTarget:
			rawTarget = t
		case *ark.RawTarget:
			rawTarget = *t
		default:
			err = errors.Errorf("graph walk cannot continue %T is not type ark.RawTarget", vertex)
			return
		}
		if rawTarget.Type != deploy.Type {
			return nil
		}
		if opts.K8sNamespace == "" {
			return errors.New("namespace is a required field when running a deploy target")
		}
		return nil
	}
}

func newExecutionWalkFunc(opts ExecuteOptions) dag.WalkFuncWithErr {
	sem := semaphore.NewWeighted(int64(opts.MaxConcurrency))
	subject := cqrs.WithSubject(cqrs.RouteKey(opts.SubscriptionID))
	return func(vertex dag.Vertex) (err error) {
		err = sem.Acquire(opts.Ctx, 1)
		if err != nil {
			err = errors.Wrap(err, "failed to acquire semaphore")
			return
		}
		defer sem.Release(1)

		var rawTarget ark.RawTarget

		// the graph isolation function produces a graph of pointers
		// this allows us to support static copies and pointer data in the graph
		switch t := vertex.(type) {
		case ark.RawTarget:
			rawTarget = t
		case *ark.RawTarget:
			rawTarget = *t
		default:
			err = errors.Errorf("graph walk cannot continue %T is not type ark.RawTarget", vertex)
			return
		}
		target, artifact, err := derivation.TargetAndArtifactFromRawTarget(rawTarget)
		if err != nil {
			return
		}

		derivative := ark.Derivation{
			Target:   target,
			Artifact: artifact,
		}
		defer func() {
			if err != nil {
				_ = opts.Broker.Publish(topic, cqrs.NewDefaultEnvelope(
					subject,
					sources.GraphWalkerSource,
					events.GraphWalkerFailedType,
					cqrs.WithData(cqrs.ApplicationJSON, derivative),
				))
			}
		}()

		if err = opts.Broker.Publish(topic, cqrs.NewDefaultEnvelope(
			subject,
			sources.GraphWalkerSource,
			events.GraphWalkerDerivationComputedType,
			cqrs.WithData(cqrs.ApplicationJSON, derivative),
		)); err != nil {
			return err
		}

		// injects artifacts with shared clients before verification
		opts.SharedClients.Inject(artifact)

		cached, err := verifyArtifact(opts.Ctx, artifact)
		if err != nil {
			return
		}

		if cached && !opts.ForceExecution {
			if err = opts.Broker.Publish(topic, cqrs.NewDefaultEnvelope(
				subject,
				sources.GraphWalkerSource,
				events.GraphWalkerActionCachedType,
				cqrs.WithData(cqrs.ApplicationJSON, derivative),
			)); err != nil {
				return
			}
			return
		}

		action, err := derivation.ActionFromTargetAndArtifact(target, artifact)
		if err != nil {
			return
		}

		rawArtifact, err := derivation.RawArtifactFromArtifact(artifact)
		if err != nil {
			return
		}

		opts.SharedClients.Inject(action)
		input := injectOrSkipLoggerInput{
			action:         action,
			logger:         opts.Logger,
			subscriptionID: opts.SubscriptionID,
			key:            target.Key(),
			name:           rawTarget.Name,
			hashCode:       rawArtifact.Hash,
			dockerClient:   opts.SharedClients.Docker,
			k8sClient:      opts.SharedClients.K8s,
		}

		// this function return a cleanup function and we defer its execution
		defer injectOrSkipLogger(input)()

		// inject actions with shared clients before execution
		if err = opts.Broker.Publish(topic, cqrs.NewDefaultEnvelope(
			subject,
			sources.GraphWalkerSource,
			events.GraphWalkerActionStartedType,
			cqrs.WithData(cqrs.ApplicationJSON, derivative),
		)); err != nil {
			return err
		}

		if err = action.Execute(opts.Ctx); err != nil {
			return err
		}

		if err = artifact.WriteState(); err != nil {
			return err
		}

		if opts.PushArtifactsAfterExecution {
			if err = opts.Broker.Publish(topic, cqrs.NewDefaultEnvelope(
				subject,
				sources.GraphWalkerSource,
				events.GraphWalkerArtifactPushStartedType,
				cqrs.WithData(cqrs.ApplicationJSON, derivative),
			)); err != nil {
				return err
			}
			if err = artifact.Push(opts.Ctx); err != nil {
				return
			}
		}

		return opts.Broker.Publish(topic, cqrs.NewDefaultEnvelope(
			subject,
			sources.GraphWalkerSource,
			events.GraphWalkerActionSuccessType,
			cqrs.WithData(cqrs.ApplicationJSON, derivative),
		))
	}
}

type injectOrSkipLoggerInput struct {
	action         interface{}
	logger         logz.FieldLogger
	subscriptionID string
	hashCode       string
	key            string
	name           string
	dockerClient   container.Docker
	k8sClient      kube.Client
}

func injectOrSkipLogger(input injectOrSkipLoggerInput) func() {

	emptyFn := func() {}

	loggerInjector, ok := input.action.(logz.Injector)
	if !ok {
		return emptyFn
	}

	dockerInjector, dockerInjectorCast := input.action.(shared_clients.DockerClientUser)
	k8sInjector, k8sInjectorCast := input.action.(shared_clients.K8sClientUser)

	hashCode := input.hashCode

	// normalized key name to not include forwarslash "/" since that will create
	// a folder structure.
	normalizedKey := strings.Replace(input.key, "/", "__", -1)

	if len(input.hashCode) > 8 {
		hashCode = input.hashCode[0:7]
	}
	childLogger := input.logger.Child(
		// this out wil be something like
		// "ark/graph/[some-guid]/[short-hash]_[some__build__ts__file.ts:some-target.log"
		logz.WithMux(transports.SuggestedLogFileWriter(
			fmt.Sprintf("ark/graph/%s", input.subscriptionID),
			fmt.Sprintf("%s_%s.log", hashCode, normalizedKey)),
		),
		logz.WithFields(logz.Fields{
			"target_key": input.key,
			"hash_code":  hashCode,
		}),
	)

	loggerInjector.UseLogger(childLogger)
	// use and inject the new child logger to k8s and docker clients
	if dockerInjectorCast {
		dockerClient := input.dockerClient
		dockerClient.OutputWriter = childLogger
		dockerInjector.UseDockerClient(dockerClient)
	}
	if k8sInjectorCast {
		k8sClient := input.k8sClient
		k8sClient.OutputWriter = childLogger
		k8sInjector.UseK8sClient(k8sClient)
	}

	// fn used to cleanup or rollback the changes made
	return func() {
		if w, ok := childLogger.(*logz.Writer); ok {
			w.Close()
		}

		// resetting logger for k8s and docker clients
		if dockerInjectorCast {
			dockerInjector.UseDockerClient(input.dockerClient)
		}
		if k8sInjectorCast {
			k8sInjector.UseK8sClient(input.k8sClient)
		}
	}

}

func verifyArtifact(ctx context.Context, artifact ark.Artifact) (bool, error) {
	if !artifact.Cacheable() {
		return false, nil
	}

	locallyCached, err := artifact.LocallyCached(ctx)
	if err != nil {
		return false, err
	}

	if locallyCached {
		return true, nil
	}

	remotelyCached, err := artifact.RemotelyCached(ctx)
	if err != nil {
		return false, err
	}

	if remotelyCached {
		return true, artifact.Pull(ctx)
	}

	return false, nil
}
