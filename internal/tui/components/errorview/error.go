package errorview

import (
	"awesomeProject/internal/tui/context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx *context.ProgramContext
	err error
}

func New(ctx *context.ProgramContext, err error) Model {
	return Model{
		ctx: ctx,
		err: err,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

func (m *Model) SetError(err error) {
	m.err = err
}

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
