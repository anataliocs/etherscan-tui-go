// Package footer provides a footer component for displaying help and keybindings.
package footer

import (
	"awesomeProject/internal/tui/context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the footer component state.
type Model struct {
	ctx  *context.ProgramContext
	help string
}

// New creates a new footer component with the given context and help text.
func New(ctx *context.ProgramContext, help string) Model {
	return Model{
		ctx:  ctx,
		help: help,
	}
}

// Update updates the footer component state. Currently a no-op.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// UpdateProgramContext updates the footer's reference to the global program context.
func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

// SetHelp updates the help text displayed in the footer.
func (m *Model) SetHelp(help string) {
	m.help = help
}

// Help returns the current help text.
func (m Model) Help() string {
	return m.help
}

// View renders the footer component as a string.
func (m Model) View() string {
	if m.ctx.ScreenWidth <= 0 {
		return ""
	}
	width := m.ctx.FooterWidth
	if width <= 0 {
		width = m.ctx.ScreenWidth
	}
	separator := m.ctx.Theme.Separator.Render(strings.Repeat("─", width))
	return separator + "\n" + m.ctx.Theme.Help.Render(m.help)
}
