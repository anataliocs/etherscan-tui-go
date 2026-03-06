package loader

import (
	"awesomeProject/internal/tui/context"
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx      *context.ProgramContext
	progress progress.Model
	text     string
}

func New(ctx *context.ProgramContext) Model {
	return Model{
		ctx:      ctx,
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case progress.FrameMsg:
		var pm tea.Model
		pm, cmd = m.progress.Update(msg)
		m.progress = pm.(progress.Model)
	}
	return m, cmd
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.progress.Width = m.ctx.ScreenWidth - 10
	if m.progress.Width > 80 {
		m.progress.Width = 80
	}
}

func (m *Model) SetText(text string) {
	m.text = text
}

func (m *Model) SetPercent(p float64) tea.Cmd {
	return m.progress.SetPercent(p)
}

func (m *Model) IncrPercent(p float64) tea.Cmd {
	return m.progress.IncrPercent(p)
}

func (m Model) Percent() float64 {
	return m.progress.Percent()
}

func (m Model) View() string {
	return fmt.Sprintf(
		"\n  Searching for %s...\n\n  %s",
		m.text,
		m.progress.View(),
	)
}
