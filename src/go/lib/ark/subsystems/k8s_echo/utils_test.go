package k8s_echo

import (
	"context"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/stretchr/testify/require"
)

type ckPublishMsgFunc func(*testing.T, *cqrs.MockBroker)

func ckPublishMsg(c ...ckPublishMsgFunc) []ckPublishMsgFunc {
	return c
}

func TestPublishMsg(t *testing.T) {
	published := func() ckPublishMsgFunc {
		return func(t *testing.T, broker *cqrs.MockBroker) {
			broker.MethodCalled("Publish", topics.K8sEchoEvents)
		}
	}
	publishedWithMsg := func(inbox <-chan cqrs.Envelope, action messages.K8sEchoResourceChangedAction) ckPublishMsgFunc {
		return func(t *testing.T, broker *cqrs.MockBroker) {
			envelope := <-inbox
			require.Equal(t, envelope.Type(), events.K8sEchoPodChanged.String())
			require.NotEmpty(t, envelope.Data())
		}
	}
	broker := cqrs.NewMockBroker()
	logger := new(logz.MockLogger)

	broker.On("Subscribe", topics.K8sEchoEvents)
	broker.On("Publish", topics.K8sEchoEvents)
	ctx := context.Background()
	inbox, err := broker.Subscribe(ctx, topics.K8sEchoEvents, nil)
	require.NoError(t, err)
	// logger.On("Child", mock.Anything)
	// logger.On("Info", mock.Anything)
	// logger.On("Debugf", mock.Anything, mock.Anything)

	testCases := map[string]struct {
		msg         messages.K8sEchoResourceChanged
		expetations []ckPublishMsgFunc
	}{
		"publishMsg publishes from a valid data": {
			messages.K8sEchoResourceChanged{
				Name:   "test pod",
				Action: messages.K8sEchoResourceChangedActionAdded,
			},
			ckPublishMsg(
				published(),
				publishedWithMsg(inbox, messages.K8sEchoResourceChangedActionAdded),
			),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			for _, check := range tc.expetations {
				publishMsg(broker, logger, events.K8sEchoPodChangedType, tc.msg)
				check(t, broker)
			}
		})
	}
}
