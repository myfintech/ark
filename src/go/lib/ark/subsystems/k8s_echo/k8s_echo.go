package k8s_echo

import (
	"context"
	"sync"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/logz"
)

// NewSubsystem factory function to return a new k8s_echo subsystem.
func NewSubsystem(
	broker cqrs.Broker,
	logger logz.FieldLogger,
	client kube.Client,
	config workspace.Config,
) *subsystems.Process {
	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": topics.K8sEcho.String(),
	}))

	return &subsystems.Process{
		Name:    topics.K8sEcho.String(),
		Factory: factory(logger, broker, client, config.K8s),
	}
}

func factory(
	logger logz.FieldLogger,
	broker cqrs.Broker,
	client kube.Client,
	k8sConfig workspace.KubernetesConfig,
) subsystems.Factory {
	return func(wg *sync.WaitGroup, ctx context.Context) func() error {
		return func() error {

			p := newPodInformer(broker, logger)
			informers := kube.InformerHandlers{
				kube.Pod: p,
			}

			opts := kube.SubscribeOptions{
				InformerHnadlers: informers,
				Duration:         k8sConfig.Informers.ResyncPeriod,
			}

			wg.Done()
			if err := kube.SubscribeToResources(ctx, client, opts); err != nil {
				return err
			}

			<-ctx.Done()
			logger.Info("shutdown received")
			return nil
		}
	}
}
