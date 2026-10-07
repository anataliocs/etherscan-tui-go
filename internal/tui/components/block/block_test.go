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
		BaseFeePerGas: "0x2540be400",
		FeeRecipient:  "0xabc",
		Status:        "Finalized",
		Transactions:  []string{"tx1", "tx2"},
		Slot:          "10",
		Epoch:         "1",
		BlockReward:   "0xde0b6b3a7640000",
		Size:          "1024",
		GasUsed:       "500000",
		GasLimit:      "1000000",
		ExtraData:     "0x1234",
		BurntFees:     "1 ETH",
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
		if !strings.Contains(view, "Block Header Hash") {
			t.Error("view should contain 'Block Header Hash'")
		}
		if !strings.Contains(view, "0x123") {
			t.Error("view should contain hash value")
		}
		if !strings.Contains(view, "100") {
			t.Error("view should contain block number")
		}
		if !strings.Contains(view, "Base Fee Per Gas") || !strings.Contains(view, "0.00000001 ETH") || !strings.Contains(view, "(10 Gwei)") {
			t.Error("view should contain base fee per gas and gwei")
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
		if !strings.Contains(view, "♦ 1 ETH") {
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
		if !strings.Contains(view, "Extra Data") || !strings.Contains(view, "0x1234") {
			t.Error("view should contain extra data")
		}
		if !strings.Contains(view, "Burnt Fees") || !strings.Contains(view, "1 ETH") {
			t.Error("view should contain burnt fees")
		}
	})

	t.Run("View - Empty Burnt Fees", func(t *testing.T) {
		blockWithNoBurntFees := &etherscan.Block{
			Hash:      "0x123",
			Number:    "100",
			Timestamp: "2023-01-01T00:00:00Z",
		}
		m := New(ctx, blockWithNoBurntFees)
		view := m.View()
		if !strings.Contains(view, "Burnt Fees") || !strings.Contains(view, "0") {
			t.Error("view should contain burnt fees with value 0")
		}
	})

	t.Run("View - Nil Block", func(t *testing.T) {
		m := New(ctx, nil)
		view := m.View()
		if view != "" {
			t.Error("view should be empty for nil block")
		}
	})

	t.Run("View - Extra Data", func(t *testing.T) {
		blockWithExtraData := &etherscan.Block{
			Hash:      "0x123",
			Number:    "100",
			ExtraData: "0x48656c6c6f", // "Hello"
		}
		m := New(ctx, blockWithExtraData)
		view := m.View()
		if !strings.Contains(view, "Hello") || !strings.Contains(view, "0x48656c6c6f") {
			t.Error("view should contain decoded extra data and raw hex")
		}
	})
}
