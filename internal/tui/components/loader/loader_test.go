package loader

import (
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	"strings"
	"testing"
)

func TestLoader(t *testing.T) {
	ctx := &context.ProgramContext{
		Theme:       theme.DefaultTheme(),
		ScreenWidth: 100,
	}

	t.Run("New", func(t *testing.T) {
		m := New(ctx)
		if m.ctx != ctx {
			t.Error("context not set correctly")
		}
	})

	t.Run("Percent", func(t *testing.T) {
		m := New(ctx)
		m.SetPercent(0.5)
		if m.Percent() != 0.5 {
			t.Errorf("expected percent 0.5, got %f", m.Percent())
		}
	})

	t.Run("SetText", func(t *testing.T) {
		m := New(ctx)
		text := "fetching transaction"
		m.SetText(text)
		if m.text != text {
			t.Errorf("expected text %q, got %q", text, m.text)
		}
		if !strings.Contains(m.View(), text) {
			t.Errorf("view should contain text, got: %s", m.View())
		}
	})

	t.Run("UpdateProgramContext", func(t *testing.T) {
		m := New(ctx)
		newCtx := &context.ProgramContext{ScreenWidth: 50}
		m.UpdateProgramContext(newCtx)
		if m.ctx != newCtx {
			t.Error("context not updated correctly")
		}
		expectedWidth := 50 - 10
		if m.progress.Width != expectedWidth {
			t.Errorf("expected progress width %d, got %d", expectedWidth, m.progress.Width)
		}
	})

	t.Run("UpdateProgramContext - Cap Width", func(t *testing.T) {
		m := New(ctx)
		newCtx := &context.ProgramContext{ScreenWidth: 200}
		m.UpdateProgramContext(newCtx)
		if m.progress.Width != 80 {
			t.Errorf("expected progress width to be capped at 80, got %d", m.progress.Width)
		}
	})
}
