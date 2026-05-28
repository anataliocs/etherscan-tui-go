package errorview

import (
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	"errors"
	"strings"
	"testing"
)

func TestErrorView(t *testing.T) {
	ctx := &context.ProgramContext{
		Theme: theme.DefaultTheme(),
	}

	t.Run("New", func(t *testing.T) {
		err := errors.New("test error")
		m := New(ctx, err)
		if !errors.Is(m.err, err) {
			t.Errorf("expected error %v, got %v", err, m.err)
		}
		if m.ctx != ctx {
			t.Error("context not set correctly")
		}
	})

	t.Run("View with error", func(t *testing.T) {
		err := errors.New("test error")
		m := New(ctx, err)
		view := m.View()
		if !strings.Contains(view, "Error") {
			t.Errorf("view should contain 'Error', got: %s", view)
		}
		if !strings.Contains(view, "test error") {
			t.Errorf("view should contain error message, got: %s", view)
		}
	})

	t.Run("View without error", func(t *testing.T) {
		m := New(ctx, nil)
		view := m.View()
		if view != "" {
			t.Errorf("expected empty view when error is nil, got: %q", view)
		}
	})

	t.Run("SetError", func(t *testing.T) {
		m := New(ctx, nil)
		newErr := errors.New("new error")
		m.SetError(newErr)
		if !errors.Is(m.err, newErr) {
			t.Errorf("expected error %v, got %v", newErr, m.err)
		}
	})

	t.Run("UpdateProgramContext", func(t *testing.T) {
		m := New(ctx, nil)
		newCtx := &context.ProgramContext{
			ScreenWidth: 100,
		}
		m.UpdateProgramContext(newCtx)
		if m.ctx != newCtx {
			t.Error("context not updated correctly")
		}
	})

	t.Run("Update", func(t *testing.T) {
		m := New(ctx, nil)
		m2, cmd := m.Update(nil)
		if cmd != nil {
			t.Error("Update should return nil cmd")
		}
		// m is a value receiver in Update, but errorview.Model is a struct.
		// Since it's currently a no-op, we just check it doesn't crash.
		if m2.ctx != m.ctx {
			t.Error("Update should return same model")
		}
	})
}
