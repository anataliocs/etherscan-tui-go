package block

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	"strings"
	"testing"
)

func TestBlock(t *testing.T) {
	ctx := &context.ProgramContext{
		Theme: theme.DefaultTheme(),
	}

	mockBlock := &etherscan.Block{
		Hash:          "0x123",
		Number:        "100",
		Timestamp:     "2023-01-01T00:00:00Z",
		BaseFeePerGas: "10 gwei",
		FeeRecipient:  "0xabc",
		Status:        "Finalized",
		Transactions:  []string{"tx1", "tx2"},
		Slot:          "10",
		Epoch:         "1",
		BlockReward:   "5 ETH",
	}

	t.Run("New", func(t *testing.T) {
		m := New(ctx, mockBlock)
		if m.ctx != ctx {
			t.Error("context not set correctly")
		}
		if m.block != mockBlock {
			t.Error("block not set correctly")
		}
	})

	t.Run("Update", func(t *testing.T) {
		m := New(ctx, mockBlock)
		updated, cmd := m.Update(nil)
		if updated != m {
			t.Error("expected no state change in Update")
		}
		if cmd != nil {
			t.Error("expected no command in Update")
		}
	})

	t.Run("UpdateProgramContext", func(t *testing.T) {
		m := New(ctx, mockBlock)
		newCtx := &context.ProgramContext{ScreenWidth: 50}
		m.UpdateProgramContext(newCtx)
		if m.ctx != newCtx {
			t.Error("context not updated correctly")
		}
	})

	t.Run("View - With Block", func(t *testing.T) {
		m := New(ctx, mockBlock)
		view := m.View()
		if !strings.Contains(view, "Block Details") {
			t.Error("view should contain 'Block Details'")
		}
		if !strings.Contains(view, "0x123") {
			t.Error("view should contain hash")
		}
		if !strings.Contains(view, "100") {
			t.Error("view should contain block number")
		}
		if !strings.Contains(view, "2") { // Number of transactions
			t.Error("view should contain transaction count")
		}
		if !strings.Contains(view, "Slot 10, Epoch 1") {
			t.Error("view should contain slot and epoch")
		}
		if !strings.Contains(view, "Block Reward") {
			t.Error("view should contain 'Block Reward'")
		}
		if !strings.Contains(view, "5 ETH") {
			t.Error("view should contain block reward value")
		}
	})

	t.Run("View - Nil Block", func(t *testing.T) {
		m := New(ctx, nil)
		view := m.View()
		if view != "" {
			t.Error("view should be empty for nil block")
		}
	})
}
