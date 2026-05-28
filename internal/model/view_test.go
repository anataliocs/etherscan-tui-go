package model

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/components/transaction"
	"fmt"
	"strings"
	"testing"
)

func TestView_States(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)
	m.ctx.ScreenWidth = 100

	tests := []struct {
		name     string
		state    sessionState
		setup    func(*Model)
		contains []string
	}{
		{
			name:  "inputState",
			state: inputState,
			setup: func(m *Model) {
				m.header.SetLatestBlock("123", "0xabc")
			},
			contains: []string{"Ethereum Transaction Explorer", "Enter transaction hash:"},
		},
		{
			name:  "loadingState",
			state: loadingState,
			setup: func(m *Model) {
				m.loader.SetText("0x123")
			},
			contains: []string{"Searching for 0x123..."},
		},
		{
			name:  "resultState",
			state: resultState,
			setup: func(m *Model) {
				m.tx = &etherscan.Transaction{Hash: "0xabc", Value: "100"}
				// We need to recreate the transaction component with the tx
				m.transaction = transaction.New(m.ctx, m.tx)
			},
			contains: []string{"Transaction Details", "0xabc"},
		},
		{
			name:  "errorState",
			state: errorState,
			setup: func(m *Model) {
				m.errorView.SetError(fmt.Errorf("not found"))
			},
			contains: []string{"Error", "not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.state = tt.state
			if tt.setup != nil {
				tt.setup(&m)
			}
			view := m.View()
			for _, s := range tt.contains {
				if !strings.Contains(view, s) {
					t.Errorf("expected view to contain %q, but it didn't\nView:\n%s", s, view)
				}
			}
		})
	}
}

func TestView_FooterWidth(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)
	m.tx = &etherscan.Transaction{Hash: "0xabc"}
	m.state = resultState

	// Large screen
	m.ctx.ScreenWidth = 100
	_ = m.View()
	expectedWidth := 60 // 0.6 * 100
	if m.ctx.FooterWidth != expectedWidth {
		t.Errorf("expected FooterWidth %d, got %d", expectedWidth, m.ctx.FooterWidth)
	}

	// Small screen
	m.ctx.ScreenWidth = 50
	_ = m.View()
	expectedWidth = 50
	if m.ctx.FooterWidth != expectedWidth {
		t.Errorf("expected FooterWidth %d, got %d", expectedWidth, m.ctx.FooterWidth)
	}
}
