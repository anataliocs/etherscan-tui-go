package footer

import (
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	"strings"
	"testing"
)

func TestFooter(t *testing.T) {
	ctx := &context.ProgramContext{
		Theme:       theme.DefaultTheme(),
		ScreenWidth: 80,
	}

	t.Run("New", func(t *testing.T) {
		help := "q: quit"
		m := New(ctx, help)
		if m.help != help {
			t.Errorf("expected help %q, got %q", help, m.help)
		}
		if m.ctx != ctx {
			t.Error("context not set correctly")
		}
	})

	t.Run("View", func(t *testing.T) {
		help := "q: quit"
		m := New(ctx, help)
		view := m.View()
		if !strings.Contains(view, help) {
			t.Errorf("view should contain help text, got: %s", view)
		}
		if !strings.Contains(view, "─") {
			t.Error("view should contain separator line")
		}
	})

	t.Run("View with zero ScreenWidth", func(t *testing.T) {
		ctxZero := &context.ProgramContext{
			Theme:       theme.DefaultTheme(),
			ScreenWidth: 0,
		}
		m := New(ctxZero, "help")
		if m.View() != "" {
			t.Error("expected empty view when ScreenWidth is 0")
		}
	})

	t.Run("SetHelp", func(t *testing.T) {
		m := New(ctx, "old help")
		newHelp := "new help"
		m.SetHelp(newHelp)
		if m.help != newHelp {
			t.Errorf("expected help %q, got %q", newHelp, m.help)
		}
		if m.Help() != newHelp {
			t.Errorf("Help() returned %q, expected %q", m.Help(), newHelp)
		}
	})

	t.Run("UpdateProgramContext", func(t *testing.T) {
		m := New(ctx, "help")
		newCtx := &context.ProgramContext{
			ScreenWidth: 100,
		}
		m.UpdateProgramContext(newCtx)
		if m.ctx != newCtx {
			t.Error("context not updated correctly")
		}
	})

	t.Run("Update", func(t *testing.T) {
		m := New(ctx, "help")
		m2, cmd := m.Update(nil)
		if cmd != nil {
			t.Error("Update should return nil cmd")
		}
		if m2.help != m.help {
			t.Error("Update should return same model")
		}
	})
}
