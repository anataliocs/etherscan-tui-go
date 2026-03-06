package ui

import (
	"awesomeProject/internal/etherscan"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTransaction(t *testing.T) {
	// Ensure lipgloss colors don't interfere with string matching
	// (though usually lipgloss is smart enough in tests)

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

	result := renderTransaction(tx)

	// Check for key components in the rendered output
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
	tx := &etherscan.Transaction{Gas: "100000"}
	item := struct {
		label string
		value string
		style lipgloss.Style
	}{
		label: "Gas Usage",
		value: "50000",
		style: lipgloss.NewStyle(),
	}

	result := renderGasUsage(tx, item, "")
	if !strings.Contains(result, "(50.00%)") {
		t.Errorf("expected gas usage percentage '(50.00%%)', got %q", result)
	}

	// Zero gas limit case
	tx.Gas = "0"
	result = renderGasUsage(tx, item, "")
	if strings.Contains(result, "%%") {
		t.Errorf("should not contain percentage when gas limit is 0, got %q", result)
	}
}

func TestRenderBlockNumber(t *testing.T) {
	tx := &etherscan.Transaction{Confirmations: "10"}
	item := struct {
		label string
		value string
		style lipgloss.Style
	}{
		label: "Block Number",
		value: "100",
		style: lipgloss.NewStyle(),
	}

	result := renderBlockNumber(tx, "100", item)
	if !strings.Contains(result, "(10 confirmations)") {
		t.Errorf("expected '(10 confirmations)', got %q", result)
	}

	tx.Confirmations = "Pending"
	result = renderBlockNumber(tx, "100", item)
	if !strings.Contains(result, "(Pending)") {
		t.Errorf("expected '(Pending)', got %q", result)
	}
}

func TestRenderTimestamp(t *testing.T) {
	// Use a fixed timestamp for testing relative time
	now := time.Now()
	past := now.Add(-10 * time.Minute)
	timestampStr := past.Format(time.RFC3339)

	item := struct {
		label string
		value string
		style lipgloss.Style
	}{
		label: "Timestamp",
		value: timestampStr,
		style: lipgloss.NewStyle(),
	}

	result := renderTimestamp(item, "")
	if !strings.Contains(result, "10m") {
		t.Errorf("expected '10m' in relative timestamp, got %q", result)
	}
}
