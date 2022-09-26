package k8s_echo

import (
	"encoding/json"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/pkg/errors"
	coreV1 "k8s.io/api/core/v1"
)

type podInformer struct {
	broker  cqrs.Broker
	logger  logz.FieldLogger
	indexes map[string]func(interface{}) ([]string, error)
}

func newPodInformer(broker cqrs.Broker, logger logz.FieldLogger) *podInformer {
	i := podInformerIndex{}
	return &podInformer{
		broker,
		logger,
		map[string]func(interface{}) ([]string, error){
			"DeployedByArk": i.deployedByArk,
			"WithLiveSync":  i.withLiveSync,
		},
	}
}

// OnAddFunc handler publishes a messages whe a pod is added
func (i podInformer) OnAddFunc(obj interface{}) {
	pod := obj.(*coreV1.Pod)
	_ = i.publishPodInfo(pod, messages.K8sEchoResourceChangedActionAdded)
}

// OnUpdateFunc handler publishes a message when a pod is updated
func (i podInformer) OnUpdateFunc(_, newObj interface{}) {
	pod := newObj.(*coreV1.Pod)
	_ = i.publishPodInfo(pod, messages.K8sEchoResourceChangedActionUpdated)
}

// OnDeleteFunc handler publishes a mesages when a pod is mark deletion
func (i podInformer) OnDeleteFunc(obj interface{}) {
	pod := obj.(*coreV1.Pod)
	_ = i.publishPodInfo(pod, messages.K8sEchoResourceChangedActionDeleted)
}

// GetIndexes returns a map of al lthe indexes defiend for pod informer
func (i podInformer) GetIndexes() map[string]func(obj interface{}) ([]string, error) {
	return i.indexes
}

// AddIndex add an index for the pod informer
func (i podInformer) AddIndex(key string, fn func(obj interface{}) ([]string, error)) {
	i.indexes[key] = fn
}

func (i podInformer) publishPodInfo(
	pod *coreV1.Pod,
	action messages.K8sEchoResourceChangedAction,
) error {
	if _, ok := pod.GetLabels()["ark.target.key"]; ok {

		data, err := json.Marshal(pod)
		if err != nil {
			return errors.Wrap(err, "fail to marshal pod to json")
		}

		msg := messages.K8sEchoResourceChanged{
			Name:        pod.GetName(),
			Action:      action,
			Namespace:   pod.GetNamespace(),
			Labels:      pod.GetLabels(),
			Annotations: pod.GetAnnotations(),
			Raw:         string(data),
		}
		return publishMsg(i.broker, i.logger, events.K8sEchoPodChangedType, msg)
	}
	return nil
}
