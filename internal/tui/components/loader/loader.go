// Package loader provides a loading screen component with a progress bar.
package loader

import (
	"awesomeProject/internal/tui/context"
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the loader component state.
type Model struct {
	ctx      *context.ProgramContext
	progress progress.Model
	text     string
}

// New creates a new loader component with the given context.
func New(ctx *context.ProgramContext) Model {
	return Model{
		ctx:      ctx,
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

// Update updates the loader component state based on the received message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if msg, ok := msg.(progress.FrameMsg); ok {
		var pm tea.Model
		pm, cmd = m.progress.Update(msg)
		if p, ok := pm.(progress.Model); ok {
			m.progress = p
		}
	}
	return m, cmd
}

// UpdateProgramContext updates the loader's reference to the global program context.
func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.progress.Width = m.ctx.ScreenWidth - 10
	if m.progress.Width > 80 {
		m.progress.Width = 80
	}
}

// SetText sets the descriptive text displayed above the progress bar.
func (m *Model) SetText(text string) {
	m.text = text
}

// SetPercent sets the progress bar percentage (0.0 to 1.0).
func (m *Model) SetPercent(p float64) tea.Cmd {
	return m.progress.SetPercent(p)
}

// IncrPercent increments the progress bar percentage by the given amount.
func (m *Model) IncrPercent(p float64) tea.Cmd {
	return m.progress.IncrPercent(p)
}

// Percent returns the current progress bar percentage.
func (m Model) Percent() float64 {
	return m.progress.Percent()
}

// View renders the loader component as a string.
func (m Model) View() string {
	return fmt.Sprintf(
		"\n  Searching for %s...\n\n  %s",
		m.text,
		m.progress.View(),
	)
}
