// Package errorview provides a component for displaying error messages.
package errorview

import (
	"awesomeProject/internal/tui/context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the error view component state.
type Model struct {
	ctx *context.ProgramContext
	err error
}

// New creates a new error view component with the given context and error.
func New(ctx *context.ProgramContext, err error) Model {
	return Model{
		ctx: ctx,
		err: err,
	}
}

// Update updates the error view component state. Currently a no-op.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// UpdateProgramContext updates the error view's reference to the global program context.
func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

// SetError sets the error to be displayed.
func (m *Model) SetError(err error) {
	m.err = err
}

// View renders the error view component as a string.
func (m Model) View() string {
	if m.err == nil {
		return ""
	}
	return fmt.Sprintf(
		"%s\n\n%s",
		m.ctx.Theme.Title.Render("Error"),
		m.ctx.Theme.Error.Render(m.err.Error()),
	)
}
