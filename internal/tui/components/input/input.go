package input

import (
	"awesomeProject/internal/tui/context"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx       *context.ProgramContext
	textInput textinput.Model
}

func New(ctx *context.ProgramContext) Model {
	ti := textinput.New()
	ti.Placeholder = "0x..."
	ti.Focus()
	ti.CharLimit = 66
	ti.Width = 70

	return Model{
		ctx:       ctx,
		textInput: ti,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

func (m Model) View() string {
	return "Enter transaction hash:\n" + m.textInput.View()
}

func (m Model) Value() string {
	return m.textInput.Value()
}

func (m *Model) SetValue(s string) {
	m.textInput.SetValue(s)
}

func (m *Model) Blur() {
	m.textInput.Blur()
}

func (m *Model) Focus() tea.Cmd {
	return m.textInput.Focus()
}
