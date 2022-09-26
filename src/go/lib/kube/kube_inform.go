package kube

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

//Resource is base type of a enum
type Resource int64

const (
	//Pod is a resource type  that matches k8s pod resource type
	Pod Resource = iota
)

//ResourceEventHandler is the contract that all resources need to implement in order to be used as an informer
type ResourceEventHandler interface {
	OnAddFunc(obj interface{})
	OnUpdateFunc(oldObj, newObj interface{})
	OnDeleteFunc(obj interface{})
	GetIndexes() map[string]func(interface{}) ([]string, error)
	AddIndex(string, func(interface{}) ([]string, error))
}

//InformerHandlers a map for resources and handlers
type InformerHandlers = map[Resource]ResourceEventHandler

// SubscribeOptions holds the handlers that will be used and the refresh rate of the informers
type SubscribeOptions struct {
	InformerHnadlers InformerHandlers
	Duration         time.Duration
}

// SubscribeToResources start informers
func SubscribeToResources(
	ctx context.Context,
	client Client,
	subscribeOpts SubscribeOptions,
) error {
	clientset, err := client.Factory.KubernetesClientSet()
	if err != nil {
		return errors.Wrap(err, "fail to create the kubernetes client set")
	}
	factory := informers.NewSharedInformerFactory(clientset, subscribeOpts.Duration)

	for k, v := range subscribeOpts.InformerHnadlers {
		switch k {
		case Pod:
			informer := factory.Core().V1().Pods().Informer()
			// registering callbacks for events
			informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    v.OnAddFunc,
				UpdateFunc: v.OnUpdateFunc,
				DeleteFunc: v.OnDeleteFunc,
			})

			// adding indexes if defined
			indexes := make(map[string]cache.IndexFunc)
			for name, index := range v.GetIndexes() {
				indexes[name] = index
			}
			if len(indexes) > 0 {
				if err := informer.AddIndexers(indexes); err != nil {
					return errors.Wrap(err, "fial to add indexers")
				}
			}

			// execute  the informers
			go func() {
				informer.Run(ctx.Done())
			}()
		}
	}
	return nil
}
