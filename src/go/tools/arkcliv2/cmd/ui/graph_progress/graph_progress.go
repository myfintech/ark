package graph_progress

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
)

type state int

const (
	queued state = iota
	running
	complete
	cached
	failed
	deployed
)

var (
	hashStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#008080"))

	nameStyle = lipgloss.NewStyle().
		MarginLeft(1).
		MarginRight(1).
		Bold(true).
		Foreground(lipgloss.Color("#3F6DAA"))

	eventStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444"))
)

func status(state state) string {
	switch state {
	case queued:
		return "🕒"
	case running:
		return "🔄"
	case failed:
		return "❌"
	case cached:
		return "♻️"
	case complete:
		return "✅"
	case deployed:
		return "🚀"
	default:
		return ""
	}
}

type GraphRenderModel struct {
	Stream    <-chan tea.Msg
	Targets   []*TargetModel
	TargetIdx map[string]*TargetModel
	viewport  viewport.Model
	ready     bool
}

type stop struct{}

func Stop() stop {
	return stop{}
}
func nextEventInStream(stream <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-stream
	}
}

func (m GraphRenderModel) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		nextEventInStream(m.Stream),
	)
}

func (m *GraphRenderModel) updateViewport(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch cm := msg.(type) {
	case tea.KeyMsg:
		switch cm.String() {
		case "down", "j":
			m.viewport.LineDown(1)
		case "up", "k":
			m.viewport.LineUp(1)
		case "pgup", "u":
			m.viewport.GotoTop()
		case "pgdown", "d":
			m.viewport.GotoBottom()
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.Model{
				Width:  cm.Width,
				Height: cm.Height - 2,
			}
			m.ready = true
		} else {
			m.viewport.Width = cm.Width
			m.viewport.Height = cm.Height - 2
			m.viewport.GotoBottom()
		}
	}
	m.viewport.SetContent(m.viewportContent().String())
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (m *GraphRenderModel) quitOnInterrupt(msg tea.Msg) tea.Cmd {
	switch cm := msg.(type) {
	case tea.KeyMsg:
		switch cm.String() {
		case "ctrl+c", "q", "esc":
			return tea.Quit
		}
	case stop:
		return tea.Quit
	}
	return nil
}

func (m *GraphRenderModel) advanceSpinnerOnTick(msg tea.Msg) tea.Cmd {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg.(type) {
	case spinner.TickMsg:
		for _, target := range m.Targets {
			_, cmd = target.Update(msg)
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, nextEventInStream(m.Stream))
	}

	return tea.Batch(cmds...)
}

func (m *GraphRenderModel) onCQRSEnvelope(msg tea.Msg) tea.Cmd {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch cm := msg.(type) {
	case cqrs.Envelope:
		var d ark.Derivative
		routeKey := cqrs.RouteKey(cm.Type())

		if err := cm.DataAs(&d); err != nil {
			cmds = append(cmds, nextEventInStream(m.Stream))
			break
		}

		if routeKey == events.GraphWalkerDerivationComputed {
			target := &TargetModel{
				state:     queued,
				name:      d.RawTarget.Key(),
				hash:      d.RawArtifact.Hash,
				startTime: time.Now(),
				lastEvent: cqrs.RouteKey(cm.Type()),
				spinner: spinner.Model{
					Spinner: spinner.MiniDot,
				},
			}

			m.Targets = append(m.Targets, target)
			m.TargetIdx[d.RawTarget.Key()] = target
		}

		if target, exists := m.TargetIdx[d.RawTarget.Key()]; exists {
			_, cmd = target.Update(msg)
			cmds = append(cmds, cmd)
		}

		m.viewport.GotoBottom()
		cmds = append(cmds, spinner.Tick)
	}

	return tea.Batch(cmds...)
}

func (m GraphRenderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	cmds = append(cmds, m.quitOnInterrupt(msg))
	cmds = append(cmds, m.advanceSpinnerOnTick(msg))
	cmds = append(cmds, m.onCQRSEnvelope(msg))
	cmds = append(cmds, m.updateViewport(msg))

	return m, tea.Batch(cmds...)
}

func (m GraphRenderModel) viewportContent() *strings.Builder {
	view := new(strings.Builder)
	for _, target := range m.Targets {
		view.WriteString(target.View() + "\n")
	}
	return view
}

func (m GraphRenderModel) View() string {
	return m.viewport.View()
}

type TargetModel struct {
	state         state
	name          string
	hash          string
	hideIcon      bool
	hideSpinner   bool
	startTime     time.Time
	lastEventTime time.Time
	lastEvent     cqrs.RouteKey
	spinner       spinner.Model
}

func (t TargetModel) Init() tea.Cmd {
	return spinner.Tick
}

func (t *TargetModel) updateTime() {
	switch t.state {
	case cached, deployed, complete, failed:
		return
	default:
		t.lastEventTime = time.Now()
		return
	}
}

func (t *TargetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	t.updateTime()
	var cmd tea.Cmd
	switch cm := msg.(type) {
	case spinner.TickMsg:
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd
	case cqrs.Envelope:
		return t.updateByCQRSEnvelope(cm)
	default:
		return t, nil
	}
}

func (t *TargetModel) updateByCQRSEnvelope(envelop cqrs.Envelope) (tea.Model, tea.Cmd) {
	t.lastEvent = cqrs.RouteKey(envelop.Type())
	switch t.lastEvent {
	case events.GraphRunnerFailed, events.GraphWalkerFailed:
		t.state = failed
		t.spinner.Finish()
		t.hideSpinner = true
		return t, nil
	case events.GraphWalkerActionStarted:
		t.state = running
		return t, nil
	case events.GraphWalkerActionCached:
		t.state = cached
		t.spinner.Finish()
		t.hideSpinner = true
		return t, nil
	case events.GraphWalkerActionSuccess:
		t.state = complete
		t.spinner.Finish()
		t.hideSpinner = true
		return t, nil
	default:
		return t, nil
	}
}

func (t TargetModel) status() string {
	if t.hideIcon {
		return ""
	}
	return status(t.state)
}

func (t TargetModel) shortHash() string {
	if len(t.hash) < 8 {
		return t.hash
	}
	return t.hash[0:7]
}

func (t TargetModel) renderSpinner() string {
	if t.hideSpinner {
		return ""
	}
	return fmt.Sprintf("[%s]", t.spinner.View())
}
func (t TargetModel) View() string {
	return fmt.Sprintf(
		"%s %s %s %s (%s)[%s]",
		t.status(),
		t.renderSpinner(),
		hashStyle.Render(t.shortHash()),
		nameStyle.Render(t.name),
		eventStyle.Render(t.lastEvent.String()),
		duration.HumanDuration(t.lastEventTime.Sub(t.startTime)),
	)
}

func New(stream <-chan tea.Msg) func() error {
	return func() error {
		return tea.NewProgram(&GraphRenderModel{
			Stream:    stream,
			Targets:   make([]*TargetModel, 0),
			TargetIdx: make(map[string]*TargetModel),
		}, tea.WithMouseCellMotion()).Start()
	}
}
