package footer

import (
	"awesomeProject/internal/tui/context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx  *context.ProgramContext
	help string
}

func New(ctx *context.ProgramContext, help string) Model {
	return Model{
		ctx:  ctx,
		help: help,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

func (m *Model) SetHelp(help string) {
	m.help = help
}

func (m Model) Help() string {
	return m.help
}

func (m Model) View() string {
	if m.ctx.ScreenWidth <= 0 {
		return ""
	}
	separator := m.ctx.Theme.Separator.Render(strings.Repeat("─", m.ctx.ScreenWidth))
	return separator + "\n" + m.ctx.Theme.Help.Render(m.help)
}
