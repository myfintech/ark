package subsystems

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
)

// Manager handles the registration and lifecycle of Process
type Manager struct {
	eg                 *errgroup.Group
	wg                 *sync.WaitGroup
	ctx                context.Context
	systems            sync.Map
	StartTimeout       time.Duration
	DisabledSubsystems sync.Map
}

// Process a unit that can be started and managed
type Process struct {
	Name    string
	Factory Factory
}

// Factory a produces a function that can start a process
// The callback is responsible for calling wg.Done() to complete process registration
type Factory func(wg *sync.WaitGroup, ctx context.Context) func() error

// Register store a Process on the Manager
// An error is returned if a process with that name already exists
func (m *Manager) Register(subsystems ...*Process) error {
	for _, process := range subsystems {
		if _, ok := m.systems.Load(process.Name); ok {
			return errors.Errorf("subsystem with name %s already exists", process.Name)
		}
		m.systems.Store(process.Name, process)
	}
	return nil
}

// Start uses all registered subsystem factories to create subsystems and kicks them off in their own go routines
// This method blocks until all subsystems have successfully started by synchronizing on a wait group
func (m *Manager) Start() error {
	waitForAllSubsystems := make(chan bool)

	m.systems.Range(func(key, value interface{}) bool {
		system := value.(*Process)
		if _, ok := m.DisabledSubsystems.Load(system.Name); ok {
			// the subsystem is disabled, so return here without increasing the delta or spinning off the subsystem
			return true
		}
		m.wg.Add(1)
		m.eg.Go(system.Factory(m.wg, m.ctx))
		return true
	})

	go func() {
		m.wg.Wait()
		close(waitForAllSubsystems)
	}()

	select {
	case <-waitForAllSubsystems:
		return nil
	case <-time.After(m.StartTimeout):
		return errors.Errorf("failed to start subsystems within configured timeout %s", m.StartTimeout)
	}
}

// Wait blocks until all subsystem go routines have exited or until one returns an error
func (m *Manager) Wait() error {
	return m.eg.Wait()
}

// NewManager creates a new Manager instance
// StartTimeout defaults to 10 seconds
func NewManager(ctx context.Context) *Manager {
	eg, gctx := errgroup.WithContext(ctx)
	return &Manager{
		eg:           eg,
		ctx:          gctx,
		wg:           new(sync.WaitGroup),
		StartTimeout: time.Second * 10,
	}
}
