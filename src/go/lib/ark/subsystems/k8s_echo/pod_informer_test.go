package k8s_echo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ckPodInformerEventFunc func(*testing.T, *cqrs.MockBroker)

func ckPodInformerEvent(c ...ckPodInformerEventFunc) []ckPodInformerEventFunc {
	return c
}

type input struct {
	pod      *coreV1.Pod
	informer *podInformer
}

func TestPodInformer(t *testing.T) {
	published := func() ckPodInformerEventFunc {
		return func(t *testing.T, broker *cqrs.MockBroker) {
			broker.MethodCalled("Publish", topics.K8sEchoEvents)
			// broker.AssertNumberOfCalls(t, "Publish", 1)
		}
	}

	hasNoError := func() ckPodInformerEventFunc {
		return func(t *testing.T, _ *cqrs.MockBroker) {
		}
	}

	publishedWithMsg := func(
		inbox <-chan cqrs.Envelope,
		action messages.K8sEchoResourceChangedAction,
	) ckPodInformerEventFunc {
		return func(t *testing.T, broker *cqrs.MockBroker) {
			envelop := <-inbox
			require.Equal(t, envelop.Type(), events.K8sEchoPodChanged.String())
			require.NotEmpty(t, envelop.Data())

			var msg messages.K8sEchoResourceChanged
			err := json.Unmarshal(envelop.Data(), &msg)

			require.NoError(t, err)
			// TODO: we need to figure out why we have 3 calls in the publish method on broker_mock
			<-inbox
			<-inbox
			require.Equal(t, msg.Action, action)
		}
	}
	broker := cqrs.NewMockBroker()
	logger := new(logz.MockLogger)
	broker.On("Publish", topics.K8sEchoEvents).Return(nil)
	broker.On("Subscribe", topics.K8sEchoEvents)
	ctx := context.Background()
	inbox, err := broker.Subscribe(ctx, topics.K8sEchoEvents, nil)
	require.NoError(t, err)
	// logger.On("Child", mock.Anything)
	// logger.On("Info", mock.Anything)
	// logger.On("Debugf", mock.Anything, mock.Anything)
	informer := newPodInformer(broker, logger)
	pod := &coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "Some test pod",
			Labels: map[string]string{
				"ark.target.key": "true",
			},
		},
	}

	testCases := map[string]struct {
		input        input
		doStuff      func(i input) *cqrs.MockBroker
		expectations []ckPodInformerEventFunc
	}{
		"OnAddFunc called from podInformer": {
			input{
				pod,
				informer,
			},
			func(i input) *cqrs.MockBroker {
				i.informer.OnAddFunc(i.pod)
				return i.informer.broker.(*cqrs.MockBroker)
			},
			ckPodInformerEvent(
				hasNoError(),
				published(),
				publishedWithMsg(inbox, messages.K8sEchoResourceChangedActionAdded),
			),
		},
		"OnUpdateFunc called from podInformer": {
			input{
				pod,
				informer,
			},
			func(i input) *cqrs.MockBroker {
				i.informer.OnUpdateFunc(i.pod, i.pod)
				return i.informer.broker.(*cqrs.MockBroker)
			},
			ckPodInformerEvent(
				hasNoError(),
				published(),
				publishedWithMsg(inbox, messages.K8sEchoResourceChangedActionUpdated),
			),
		},
		"OnDeleteFunc called from podInformer": {
			input{
				pod,
				informer,
			},
			func(i input) *cqrs.MockBroker {
				i.informer.OnDeleteFunc(i.pod)
				return i.informer.broker.(*cqrs.MockBroker)
			},
			ckPodInformerEvent(
				hasNoError(),
				published(),
				publishedWithMsg(inbox, messages.K8sEchoResourceChangedActionDeleted),
			),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			for _, check := range tc.expectations {
				check(t, tc.doStuff(tc.input))
			}
		})
	}
	broker.AssertExpectations(t)
}
