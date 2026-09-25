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
		Size:          "1024",
		GasUsed:       "500000",
		GasLimit:      "1000000",
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
		if !strings.Contains(view, "Base Fee Per Gas") || !strings.Contains(view, "10 gwei") {
			t.Error("view should contain base fee per gas")
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
		if !strings.Contains(view, "1024 bytes") {
			t.Error("view should contain size")
		}
		if !strings.Contains(view, "Gas Limit") || !strings.Contains(view, "1000000") {
			t.Error("view should contain gas limit")
		}
		if !strings.Contains(view, "Gas Used") || !strings.Contains(view, "500000 (50.00%)") {
			t.Error("view should contain gas used and percentage")
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
