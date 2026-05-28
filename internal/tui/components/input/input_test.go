package input

import (
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInput(t *testing.T) {
	ctx := &context.ProgramContext{
		Theme: theme.DefaultTheme(),
	}

	t.Run("New", func(t *testing.T) {
		m := New(ctx)
		if m.ctx != ctx {
			t.Error("context not set correctly")
		}
		if !m.textInput.Focused() {
			t.Error("expected textInput to be focused")
		}
	})

	t.Run("Value/SetValue", func(t *testing.T) {
		m := New(ctx)
		val := "0x123"
		m.SetValue(val)
		if m.Value() != val {
			t.Errorf("expected value %q, got %q", val, m.Value())
		}
	})

	t.Run("Focus/Blur", func(t *testing.T) {
		m := New(ctx)
		m.Blur()
		if m.textInput.Focused() {
			t.Error("expected textInput to be blurred")
		}
		m.Focus()
		if !m.textInput.Focused() {
			t.Error("expected textInput to be focused after Focus()")
		}
	})

	t.Run("View", func(t *testing.T) {
		m := New(ctx)
		view := m.View()
		if !strings.Contains(view, "Enter transaction hash:") {
			t.Error("view should contain prompt")
		}
	})

	t.Run("Update", func(t *testing.T) {
		m := New(ctx)
		// Simulate a key press
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
		m2, _ := m.Update(msg)
		if m2.Value() != "a" {
			t.Errorf("expected value 'a', got %q", m2.Value())
		}
	})

	t.Run("UpdateProgramContext", func(t *testing.T) {
		m := New(ctx)
		newCtx := &context.ProgramContext{ScreenWidth: 100}
		m.UpdateProgramContext(newCtx)
		if m.ctx != newCtx {
			t.Error("context not updated correctly")
		}
	})
}
