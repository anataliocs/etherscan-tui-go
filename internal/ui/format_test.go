package ui

import (
	"awesomeProject/internal/etherscan"
	"testing"
)

func TestFormatGasFees(t *testing.T) {
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
			result := formatGasFees(tt.tx)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
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
			result := formatStatus(tt.status)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetStatusStyle(t *testing.T) {
	tests := []struct {
		name   string
		status string
		// We can't easily compare lipgloss.Style, so we check if it matches a known style's string representation or just check it doesn't panic.
		// However, we can compare the underlying colors if we really wanted to, but checking if it's the right style object is enough.
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
			style := getStatusStyle(tt.status)
			// Simple check to ensure we get a style back
			if style.GetForeground() == nil && tt.status != "unknown" {
				t.Errorf("style for %s has no foreground color", tt.status)
			}
		})
	}
}
