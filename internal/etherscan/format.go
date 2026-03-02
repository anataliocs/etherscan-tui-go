// Package etherscan provides formatting utilities for Ethereum-related data.
package etherscan

import (
	"fmt"
	"math/big"
	"strings"
)

// formatValue converts a hex string (Wei) to a human-readable ETH string.
// Parameters:
//   - hexStr: The hex value in Wei.
//
// Returns:
//   - A formatted string with the ETH symbol and value.
func formatValue(hexStr string) string {
	eth, s, done := hexToFloat(hexStr, 1e18)
	if done {
		return s
	}

	return fmt.Sprintf("♦ %s ETH", eth.Text('f', -1))
}

// formatGwei converts a hex string (Wei) to Gwei as a string.
// Parameters:
//   - hexStr: The hex value in Wei.
//
// Returns:
//   - The value in Gwei as a decimal string.
func formatGwei(hexStr string) string {
	if hexStr == "" {
		return ""
	}
	gwei, s, done := hexToFloat(hexStr, 1e9)
	if done {
		return s
	}
	return gwei.Text('f', -1)
}

// formatGasPrice converts a hex string (Wei) to a formatted Gwei and ETH gas price string.
// Parameters:
//   - hexStr: The hex value in Wei.
//
// Returns:
//   - A formatted string with gas pump emoji, Gwei value, and ETH value.
func formatGasPrice(hexStr string) string {
	gwei, s, done := hexToFloat(hexStr, 1e9)
	if done {
		return s
	}

	eth, _, _ := hexToFloat(hexStr, 1e18)

	return fmt.Sprintf("⛽ %s Gwei (%s ETH)", gwei.Text('f', -1), eth.Text('f', -1))
}

// formatTransactionFee calculates and formats the transaction fee in ETH.
// Parameters:
//   - gasUsedHex: The gas used in hex.
//   - gasPriceHex: The gas price in hex.
//
// Returns:
//   - The calculated fee in ETH as a formatted string.
func formatTransactionFee(gasUsedHex, gasPriceHex string) string {
	if gasUsedHex == "" || gasPriceHex == "" {
		return ""
	}

	gu := new(big.Int)
	if _, ok := gu.SetString(strings.TrimPrefix(gasUsedHex, "0x"), 16); !ok {
		return ""
	}

	gp := new(big.Int)
	if _, ok := gp.SetString(strings.TrimPrefix(gasPriceHex, "0x"), 16); !ok {
		return ""
	}

	// Fee = gasUsed * gasPrice
	feeWei := new(big.Int).Mul(gu, gp)

	// 1 ETH = 10^18 Wei
	feeEth := new(big.Float).SetInt(feeWei)
	feeEth.Quo(feeEth, big.NewFloat(1e18))

	return fmt.Sprintf("%s ETH", feeEth.Text('f', -1))
}

// formatTransactionType returns a human-readable description for an Ethereum transaction type.
// Parameters:
//   - hexStr: The transaction type in hex.
//
// Returns:
//   - A human-readable description (e.g., "2 (EIP-1559)").
func formatTransactionType(hexStr string) string {
	if hexStr == "" || hexStr == "0x" {
		return "0 (Legacy)"
	}
	bi := stringToBigInt(hexStr)
	if bi == nil {
		return hexStr
	}

	val := bi.Int64()
	switch val {
	case 0:
		return "0 (Legacy)"
	case 1:
		return "1 (Access List)"
	case 2:
		return "2 (EIP-1559)"
	case 3:
		return "3 (EIP-4844)"
	default:
		return fmt.Sprintf("%d", val)
	}
}
