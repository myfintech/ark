package subsystems

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"
)

type mockSubsystemProbe struct {
	mock.Mock
}

func (m *mockSubsystemProbe) OnInit() {
	m.Called()
}

func (m *mockSubsystemProbe) OnStart() {
	m.Called()
}

func (m *mockSubsystemProbe) OnCancel() {
	m.Called()
}

func TestManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	manager := NewManager(ctx)

	probe := new(mockSubsystemProbe)

	proc1 := &Process{
		Name:    "test",
		Factory: newMockSubsystemFactory(probe),
	}

	proc2 := &Process{
		Name:    "test2",
		Factory: newMockSubsystemFactory(probe),
	}

	probe.On("OnInit").Return()
	probe.On("OnStart").Return()
	probe.On("OnCancel").Return()

	t.Run("should successfully register multiple subsystems", func(t *testing.T) {
		err := manager.Register(proc1, proc2)
		require.NoError(t, err)
	})

	t.Run("should error when registering a subsystem with a conflicting name", func(t *testing.T) {
		err := manager.Register(proc1)
		require.Error(t, err)
	})

	t.Run("should start all known subsystems", func(t *testing.T) {
		err := manager.Start()
		require.NoError(t, err)
		probe.AssertNumberOfCalls(t, "OnInit", 2)
		probe.AssertNumberOfCalls(t, "OnStart", 2)
	})

	t.Run("should stop all known subsystems", func(t *testing.T) {
		time.AfterFunc(time.Second*1, cancel)
		err := manager.Wait()
		require.NoError(t, err)
		probe.AssertNumberOfCalls(t, "OnCancel", 2)
	})

	probe.AssertExpectations(t)
}

func TestManagerTimeout(t *testing.T) {
	manager := NewManager(context.Background())
	manager.StartTimeout = time.Second
	err := manager.Register(&Process{
		Name: "timeout",
		Factory: func(wg *sync.WaitGroup, ctx context.Context) func() error {
			return func() error {
				return nil
			}
		},
	})
	require.NoError(t, err)
	require.Error(t, manager.Start())
}

func newMockSubsystemFactory(probe *mockSubsystemProbe) func(wg *sync.WaitGroup, ctx context.Context) func() error {
	return func(wg *sync.WaitGroup, ctx context.Context) func() error {
		probe.OnInit()
		return func() error {
			probe.OnStart()
			wg.Done()
			<-ctx.Done()
			probe.OnCancel()
			return nil
		}
	}
}
