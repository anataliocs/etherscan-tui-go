package etherscan

import (
	"testing"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		hex      string
		expected string
	}{
		{"0xde0b6b3a7640000", "♦ 1 ETH"},
		{"0x0", "♦ 0 ETH"},
		{"", ""},
	}

	for _, tt := range tests {
		got := formatValue(tt.hex)
		if got != tt.expected {
			t.Errorf("formatValue(%s) = %s; want %s", tt.hex, got, tt.expected)
		}
	}
}

func TestFormatGwei(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"0x3b9aca00", "1"},
		{"", ""},
		{"0x0", "0"},
	}

	for _, tt := range tests {
		got := formatGwei(tt.hex)
		if got != tt.want {
			t.Errorf("formatGwei(%s) = %s; want %s", tt.hex, got, tt.want)
		}
	}
}

func TestFormatGasPrice(t *testing.T) {
	tests := []struct {
		hex      string
		expected string
	}{
		{"0x3b9aca00", "⛽ 1 Gwei (0.000000001 ETH)"},
		{"", ""},
	}

	for _, tt := range tests {
		got := formatGasPrice(tt.hex)
		if got != tt.expected {
			t.Errorf("formatGasPrice(%s) = %s; want %s", tt.hex, got, tt.expected)
		}
	}
}

func TestFormatTransactionFee(t *testing.T) {
	tests := []struct {
		gasUsed  string
		gasPrice string
		expected string
	}{
		{"0x5208", "0x3b9aca00", "0.000021 ETH"}, // 21000 * 1 Gwei
		{"", "0x1", ""},
		{"0x1", "", ""},
	}

	for _, tt := range tests {
		got := formatTransactionFee(tt.gasUsed, tt.gasPrice)
		if got != tt.expected {
			t.Errorf("formatTransactionFee(%s, %s) = %s; want %s", tt.gasUsed, tt.gasPrice, got, tt.expected)
		}
	}
}

func TestFormatTransactionType(t *testing.T) {
	tests := []struct {
		hex      string
		expected string
	}{
		{"0x0", "0 (Legacy)"},
		{"0x1", "1 (Access List)"},
		{"0x2", "2 (EIP-1559)"},
		{"0x02", "2 (EIP-1559)"},
		{"0x3", "3 (EIP-4844)"},
		{"0xa", "10"},
		{"", "0 (Legacy)"},
		{"0x", "0 (Legacy)"},
	}

	for _, tt := range tests {
		got := formatTransactionType(tt.hex)
		if got != tt.expected {
			t.Errorf("formatTransactionType(%s) = %s; want %s", tt.hex, got, tt.expected)
		}
	}
}

func TestFormatLatestBlock(t *testing.T) {
	got := FormatLatestBlock("0xa")
	if got != "10" {
		t.Errorf("FormatLatestBlock(0xa) = %s; want 10", got)
	}
}
