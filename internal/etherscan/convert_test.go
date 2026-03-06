package etherscan

import (
	"math/big"
	"testing"
)

func TestHexToFloat(t *testing.T) {
	tests := []struct {
		name       string
		hex        string
		divisor    float64
		wantVal    string
		wantBackup string
		wantDone   bool
	}{
		{
			name:       "Empty",
			hex:        "",
			divisor:    1e18,
			wantBackup: "",
			wantDone:   true,
		},
		{
			name:       "NoPrefix",
			hex:        "123",
			divisor:    1e18,
			wantBackup: "123",
			wantDone:   true,
		},
		{
			name:       "PrefixOnly",
			hex:        "0x",
			divisor:    1e18,
			wantBackup: "0 ETH",
			wantDone:   true,
		},
		{
			name:     "ETH",
			hex:      "0xde0b6b3a7640000", // 1e18
			divisor:  1e18,
			wantVal:  "1",
			wantDone: false,
		},
		{
			name:     "Gwei",
			hex:      "0x3b9aca00", // 1e9
			divisor:  1e9,
			wantVal:  "1",
			wantDone: false,
		},
		{
			name:       "InvalidHex",
			hex:        "0xxyz",
			divisor:    1e18,
			wantBackup: "0xxyz",
			wantDone:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotBackup, gotDone := hexToFloat(tt.hex, tt.divisor)
			if gotDone != tt.wantDone {
				t.Errorf("hexToFloat() gotDone = %v, want %v", gotDone, tt.wantDone)
			}
			if gotBackup != tt.wantBackup {
				t.Errorf("hexToFloat() gotBackup = %v, want %v", gotBackup, tt.wantBackup)
			}
			if gotVal != nil {
				gotStr := gotVal.Text('f', -1)
				if gotStr != tt.wantVal {
					t.Errorf("hexToFloat() gotVal = %v, want %v", gotStr, tt.wantVal)
				}
			} else if tt.wantVal != "" {
				t.Errorf("hexToFloat() gotVal = nil, want %v", tt.wantVal)
			}
		})
	}
}

func TestHexToDecimal(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"0x1", "1"},
		{"0xa", "10"},
		{"0xff", "255"},
		{"", ""},
		{"123", "123"},
		{"0x", "0"},
		{"0xinvalid", "0xinvalid"},
	}

	for _, tt := range tests {
		got := hexToDecimal(tt.hex)
		if got != tt.want {
			t.Errorf("hexToDecimal(%s) = %s; want %s", tt.hex, got, tt.want)
		}
	}
}

func TestCalculateConfirmations(t *testing.T) {
	tests := []struct {
		latest string
		tx     string
		want   string
	}{
		{"10", "10", "1"},
		{"0xa", "0xa", "1"},
		{"10", "9", "2"},
		{"0xa", "0x9", "2"},
		{"10", "11", "0"},
		{"", "10", ""},
		{"10", "", ""},
		{"10", "0x0", ""},
		{"invalid", "10", "error"},
	}

	for _, tt := range tests {
		got := calculateConfirmations(tt.latest, tt.tx)
		if got != tt.want {
			t.Errorf("calculateConfirmations(%s, %s) = %s; want %s", tt.latest, tt.tx, got, tt.want)
		}
	}
}

func TestStringToBigInt(t *testing.T) {
	tests := []struct {
		s    string
		want *big.Int
	}{
		{"10", big.NewInt(10)},
		{"0xa", big.NewInt(10)},
		{"0x10", big.NewInt(16)},
		{"invalid", nil},
	}

	for _, tt := range tests {
		got := stringToBigInt(tt.s)
		if tt.want == nil {
			if got != nil {
				t.Errorf("stringToBigInt(%s) = %v; want nil", tt.s, got)
			}
		} else {
			if got == nil || got.Cmp(tt.want) != 0 {
				t.Errorf("stringToBigInt(%s) = %v; want %v", tt.s, got, tt.want)
			}
		}
	}
}

func TestCalculateBurntFees(t *testing.T) {
	tests := []struct {
		gasUsed  string
		baseFee  string
		expected string
	}{
		{"0x5208", "0x1d1a94a200", "0.002625 ETH 🔥"},
		{"", "0x1", ""},
		{"0x1", "", ""},
	}

	for _, tt := range tests {
		got := calculateBurntFees(tt.gasUsed, tt.baseFee)
		if got != tt.expected {
			t.Errorf("calculateBurntFees(%s, %s) = %s; want %s", tt.gasUsed, tt.baseFee, got, tt.expected)
		}
	}
}

func TestCalculateSavings(t *testing.T) {
	tests := []struct {
		gasUsed        string
		maxFee         string
		effectivePrice string
		expected       string
	}{
		// 10 Gwei = 10,000,000,000 Wei = 0x2540BE400
		// 5 Gwei = 5,000,000,000 Wei = 0x12A05F200
		// diff = 5 Gwei = 5,000,000,000
		// gasUsed = 21000 (0x5208)
		// total = 5,000,000,000 * 21,000 = 105,000,000,000,000 Wei
		// 105,000,000,000,000 / 1e18 = 0.000105 ETH
		{"0x5208", "0x2540be400", "0x12a05f200", "0.000105 ETH 💸"},
		{"0x5208", "0x12a05f200", "0x2540be400", ""}, // negative savings
		{"0x5208", "0x12a05f200", "0x12a05f200", ""}, // zero savings
		{"", "0x1", "0x1", ""},
	}

	for _, tt := range tests {
		got := calculateSavings(tt.gasUsed, tt.maxFee, tt.effectivePrice)
		if got != tt.expected {
			t.Errorf("calculateSavings(%s, %s, %s) = %s; want %s", tt.gasUsed, tt.maxFee, tt.effectivePrice, got, tt.expected)
		}
	}
}
