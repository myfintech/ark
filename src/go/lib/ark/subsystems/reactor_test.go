package subsystems

import (
	"context"
	"testing"
	"time"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestReactorErrorHandling(t *testing.T) {
	t.Run("should reject non-retryable error cases", func(t *testing.T) {
		var err error = cqrs.NewRetryableError(errors.New("I am a retryable error"))
		err = errors.Wrap(err, "Encapsulated error")
		ctx := context.Background()
		msgFunc := cqrs.OnMessageFunc(func(ctx context.Context, msg cqrs.Envelope) error {
			return err
		})
		msgErrFunc := cqrs.OnMessageErrorFunc(func(ctx context.Context, msg cqrs.Envelope, err error) error {
			return nil
		})
		env := cqrs.NewDefaultEnvelope()
		reactorProcessMessage(msgFunc, ctx, env, msgErrFunc, logz.NoOpLogger{})
		select {
		case msgErr := <-env.Wait():
			require.Equal(t, msgErr, err)
		case <-time.After(time.Second * 5):
			t.Fail()
		}
	})
}
