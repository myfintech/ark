package fs_observer

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"github.com/reactivex/rxgo/v2"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/logz"
)

func TestNewSubsystem(t *testing.T) {
	ctx, shutdown := context.WithCancel(context.Background())
	logger := new(logz.MockLogger)
	broker := cqrs.NewMockBroker()
	observer := make(chan rxgo.Item)

	wg := new(sync.WaitGroup)
	eg, _ := errgroup.WithContext(ctx)

	logger.On("Child", mock.Anything)
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Info", mock.Anything, mock.Anything, mock.Anything)
	broker.On("Subscribe", topics.FSObserverEvents)

	wg.Add(1)
	eg.Go(NewSubsystem(logger, broker, observer).Factory(wg, ctx))
	wg.Wait()

	inbox, err := broker.Subscribe(ctx, topics.FSObserverEvents, nil)
	require.NoError(t, err)

	var testCases = map[string]struct {
		WantErr        bool
		File           rxgo.Item
		SetExpectation func()
		Wait           func()
	}{

		"should publish msg on success": {
			File: rxgo.Item{
				V: []*fs.File{{
					Name:          "/src/file.go",
					Exists:        true,
					New:           false,
					Type:          "f",
					Hash:          hex.EncodeToString(sha1.New().Sum(nil)),
					SymlinkTarget: "",
					RelName:       "src/file.go",
				}},
				E: nil,
			},
			SetExpectation: func() {
				broker.
					On("Publish", topics.FSObserverEvents)
			},
			Wait: func() {
				select {
				case <-inbox:
				case <-time.After(time.Second * 2):
					t.Error("failed to receive a message with in 2 seconds time out")
				}
			},
		},
		"should fail": {
			File: rxgo.Item{
				V: nil,
				E: errors.New(""),
			},
			SetExpectation: func() {
				logger.On("Error", mock.Anything)
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			tt.SetExpectation()
			observer <- tt.File
			if tt.Wait != nil {
				tt.Wait()
			}

			broker.AssertExpectations(t)
			// FIXME(rckgomz): we and to bring back the logger assertions so we can have more coverage
			// logger.AssertExpectations(t)
		})

	}

	// shutdown the subsystem
	shutdown()
	close(observer)
	require.NoError(t, eg.Wait())
}
