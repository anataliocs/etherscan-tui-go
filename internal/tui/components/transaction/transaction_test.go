package transaction

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestFormatGasFees(t *testing.T) {
	ctx := &context.ProgramContext{Theme: theme.DefaultTheme()}
	m := New(ctx, nil)

	tests := []struct {
		name     string
		tx       *etherscan.Transaction
		expected string
	}{
		{
			name: "All Fees Present",
			tx: &etherscan.Transaction{
				BaseFeePerGas:        "10",
				MaxFeePerGas:         "20",
				MaxPriorityFeePerGas: "2",
			},
			expected: "⛽ Base: 10 Gwei | Max: 20 Gwei | Max Priority: 2 Gwei",
		},
		{
			name: "Only Base Fee",
			tx: &etherscan.Transaction{
				BaseFeePerGas: "10",
			},
			expected: "⛽ Base: 10 Gwei | Max: n/a Gwei | Max Priority: n/a Gwei",
		},
		{
			name: "All Fees Empty",
			tx: &etherscan.Transaction{
				BaseFeePerGas:        "",
				MaxFeePerGas:         "",
				MaxPriorityFeePerGas: "",
			},
			expected: "n/a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.formatGasFees(tt.tx)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	ctx := &context.ProgramContext{Theme: theme.DefaultTheme()}
	m := New(ctx, nil)

	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"Success", "success", "✔ success"},
		{"Success Upper", "SUCCESS", "✔ success"},
		{"Failed", "failed", "✘ failed"},
		{"Pending", "pending", "⧖ Pending"},
		{"Dropped", "dropped", "↓ dropped"},
		{"Replaced", "replaced", "↺ replaced"},
		{"Unknown", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.formatStatus(tt.status)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetStatusStyle(t *testing.T) {
	ctx := &context.ProgramContext{Theme: theme.DefaultTheme()}
	m := New(ctx, nil)

	tests := []struct {
		name   string
		status string
	}{
		{"Success", "success"},
		{"Failed", "failed"},
		{"Pending", "pending"},
		{"Dropped", "dropped"},
		{"Replaced", "replaced"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := m.getStatusStyle(tt.status)
			if style.GetForeground() == nil && tt.status != "unknown" {
				t.Errorf("style for %s has no foreground color", tt.status)
			}
		})
	}
}

func TestRenderTransaction(t *testing.T) {
	ctx := &context.ProgramContext{Theme: theme.DefaultTheme()}
	tx := &etherscan.Transaction{
		Status:         "success",
		Hash:           "0x123",
		Type:           "2 (EIP-1559)",
		Timestamp:      "2024-02-20T20:12:48Z",
		BlockNumber:    "11",
		Value:          "0 ETH",
		Gas:            "21000",
		GasUsed:        "21000",
		GasPrice:       "10 Gwei (0.00000001 ETH)",
		TransactionFee: "0.00021 ETH",
		Confirmations:  "100",
		MaxFeePerGas:   "20",
		BaseFeePerGas:  "10",
		ToAccountType:  "EOA",
	}
	m := New(ctx, tx)

	result := m.View()

	expectedSubstrings := []string{
		"Transaction Details",
		"✔ success",
		"0x123",
		"2 (EIP-1559)",
		"11",
		"(100 confirmations)",
		"21000",
		"(100.00%)",
		"EOA",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(result, sub) {
			t.Errorf("rendered output missing expected substring: %q", sub)
		}
	}
}

func TestRenderGasUsage(t *testing.T) {
	ctx := &context.ProgramContext{Theme: theme.DefaultTheme()}
	m := New(ctx, nil)

	tx := &etherscan.Transaction{Gas: "100000"}
	result := m.renderGasUsage(tx, "50000", lipgloss.NewStyle())
	if !strings.Contains(result, "(50.00%)") {
		t.Errorf("expected gas usage percentage '(50.00%%)', got %q", result)
	}

	tx.Gas = "0"
	result = m.renderGasUsage(tx, "50000", lipgloss.NewStyle())
	if strings.Contains(result, "%%") {
		t.Errorf("should not contain percentage when gas limit is 0, got %q", result)
	}
}

func TestRenderBlockNumber(t *testing.T) {
	ctx := &context.ProgramContext{Theme: theme.DefaultTheme()}
	m := New(ctx, nil)

	tx := &etherscan.Transaction{Confirmations: "10"}
	result := m.renderBlockNumber(tx, "100", lipgloss.NewStyle())
	if !strings.Contains(result, "(10 confirmations)") {
		t.Errorf("expected '(10 confirmations)', got %q", result)
	}

	tx.Confirmations = "Pending"
	result = m.renderBlockNumber(tx, "100", lipgloss.NewStyle())
	if !strings.Contains(result, "(Pending)") {
		t.Errorf("expected '(Pending)', got %q", result)
	}
}

func TestRenderTimestamp(t *testing.T) {
	ctx := &context.ProgramContext{Theme: theme.DefaultTheme()}
	m := New(ctx, nil)

	// Use a fixed timestamp for testing relative time
	now := time.Now()
	past := now.Add(-10 * time.Minute)
	timestampStr := past.Format(time.RFC3339)

	result := m.renderTimestamp(timestampStr, lipgloss.NewStyle())
	if !strings.Contains(result, "10m") {
		t.Errorf("expected '10m' in relative timestamp, got %q", result)
	}
}
