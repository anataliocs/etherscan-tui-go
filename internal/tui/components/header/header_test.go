package header

import (
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	"strings"
	"testing"
)

func TestHeader(t *testing.T) {
	ctx := &context.ProgramContext{
		Theme: theme.DefaultTheme(),
	}

	t.Run("New", func(t *testing.T) {
		m := New(ctx, 1)
		if m.chainID != 1 {
			t.Errorf("expected chainID 1, got %d", m.chainID)
		}
		if !m.isFetchingBlock {
			t.Error("expected isFetchingBlock to be true")
		}
	})

	t.Run("SetLatestBlock", func(t *testing.T) {
		m := New(ctx, 1)
		m.SetLatestBlock("12345", "0xabc")
		if m.latestBlock != "12345" {
			t.Errorf("expected latestBlock 12345, got %s", m.latestBlock)
		}
		if m.latestTxHash != "0xabc" {
			t.Errorf("expected latestTxHash 0xabc, got %s", m.latestTxHash)
		}
		if m.isFetchingBlock {
			t.Error("expected isFetchingBlock to be false")
		}
		if m.LatestTxHash() != "0xabc" {
			t.Errorf("LatestTxHash() returned %s, expected 0xabc", m.LatestTxHash())
		}
	})

	t.Run("SetChainID", func(t *testing.T) {
		m := New(ctx, 1)
		m.SetLatestBlock("12345", "0xabc")
		m.SetChainID(11155111)
		if m.chainID != 11155111 {
			t.Errorf("expected chainID 11155111, got %d", m.chainID)
		}
		if !m.isFetchingBlock {
			t.Error("expected isFetchingBlock to be true after SetChainID")
		}
	})

	t.Run("View - Mainnet", func(t *testing.T) {
		m := New(ctx, 1)
		m.SetLatestBlock("100", "0xhash")
		view := m.View()
		if !strings.Contains(view, "Mainnet") {
			t.Error("view should contain 'Mainnet'")
		}
		if !strings.Contains(view, "100") {
			t.Error("view should contain block number")
		}
		if !strings.Contains(view, "0xhash") {
			t.Error("view should contain tx hash")
		}
	})

	t.Run("View - Sepolia", func(t *testing.T) {
		m := New(ctx, 11155111)
		view := m.View()
		if !strings.Contains(view, "Sepolia") {
			t.Error("view should contain 'Sepolia'")
		}
	})

	t.Run("UpdateProgramContext", func(t *testing.T) {
		m := New(ctx, 1)
		newCtx := &context.ProgramContext{ScreenWidth: 50}
		m.UpdateProgramContext(newCtx)
		if m.ctx != newCtx {
			t.Error("context not updated correctly")
		}
	})
}
