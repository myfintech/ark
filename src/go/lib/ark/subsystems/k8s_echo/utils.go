package k8s_echo

import (
	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/sources"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/pkg/errors"
)

func publishMsg(
	broker cqrs.Broker,
	logger logz.FieldLogger,
	envelopOpt cqrs.EnvelopeOption,
	msg messages.K8sEchoResourceChanged,
) error {
	if err := broker.Publish(topics.K8sEchoEvents, cqrs.NewDefaultEnvelope(
		sources.K8sEcho,
		envelopOpt,
		cqrs.WithData(cqrs.ApplicationJSON, msg),
	)); err != nil {
		logger.Error(errors.Wrap(err, "failed to publish change notification"))
	}

	return nil
}
