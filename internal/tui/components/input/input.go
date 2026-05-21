// Package input provides a text input component for entering transaction hashes.
package input

import (
	"awesomeProject/internal/tui/context"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the input component state.
type Model struct {
	ctx       *context.ProgramContext
	textInput textinput.Model
}

// New creates a new input component with the given context.
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

// Update updates the input component state based on the received message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// UpdateProgramContext updates the input's reference to the global program context.
func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

// View renders the input component as a string.
func (m Model) View() string {
	return "Enter transaction hash:\n" + m.textInput.View()
}

// Value returns the current text value of the input.
func (m Model) Value() string {
	return m.textInput.Value()
}

// SetValue sets the current text value of the input.
func (m *Model) SetValue(s string) {
	m.textInput.SetValue(s)
}

// Blur removes focus from the input.
func (m *Model) Blur() {
	m.textInput.Blur()
}

// Focus sets focus on the input.
func (m *Model) Focus() tea.Cmd {
	return m.textInput.Focus()
}
