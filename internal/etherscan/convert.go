// Package etherscan provides conversion utilities for Ethereum data types.
package etherscan

import (
	"fmt"
	"math/big"
	"strings"
)

const (
	weiInEth  = 1e18
	weiInGwei = 1e9
)

// stringToBigInt converts a hex (with "0x" prefix) or decimal string to a *big.Int.
func stringToBigInt(s string) *big.Int {
	if s == "" {
		return nil
	}
	bi := new(big.Int)
	base := 10
	trimmed := s
	if strings.HasPrefix(s, "0x") {
		base = 16
		trimmed = strings.TrimPrefix(s, "0x")
	}
	if trimmed == "" && base == 16 {
		return new(big.Int)
	}

	if _, ok := bi.SetString(trimmed, base); !ok {
		return nil
	}
	return bi
}

// weiToEth converts a big.Int Wei value to a big.Float ETH value.
func weiToEth(wei *big.Int) *big.Float {
	if wei == nil {
		return new(big.Float)
	}
	f := new(big.Float).SetInt(wei)
	return f.Quo(f, big.NewFloat(weiInEth))
}

// weiToGwei converts a big.Int Wei value to a big.Float Gwei value.
func weiToGwei(wei *big.Int) *big.Float {
	if wei == nil {
		return new(big.Float)
	}
	f := new(big.Float).SetInt(wei)
	return f.Quo(f, big.NewFloat(weiInGwei))
}

// hexToFloat converts a hex string to a big.Float using the given divisor.
// Deprecated: use stringToBigInt and weiToEth/weiToGwei instead.
func hexToFloat(hexStr string, val float64) (*big.Float, string, bool) {
	if hexStr == "" {
		return nil, "", true
	}
	if !strings.HasPrefix(hexStr, "0x") {
		return nil, hexStr, true
	}

	bi := stringToBigInt(hexStr)
	if bi == nil {
		return nil, hexStr, true
	}

	if hexStr == "0x" {
		return nil, "0 ETH", true
	}

	f := new(big.Float).SetInt(bi)
	f.Quo(f, big.NewFloat(val))
	return f, "", false
}

// calculateBurntFees calculates burnt fees in ETH given gas used and base fee.
func calculateBurntFees(gasUsedHex, baseFeeHex string) string {
	gu := stringToBigInt(gasUsedHex)
	bf := stringToBigInt(baseFeeHex)
	if gu == nil || bf == nil {
		return ""
	}

	burntWei := new(big.Int).Mul(gu, bf)
	burntEth := weiToEth(burntWei)

	return fmt.Sprintf("%s ETH 🔥", burntEth.Text('f', -1))
}

// calculateSavings calculates the ETH saved when MaxFeePerGas exceeds EffectiveGasPrice.
func calculateSavings(gasUsedHex, maxFeeHex, effectivePriceHex string) string {
	gu := stringToBigInt(gasUsedHex)
	mf := stringToBigInt(maxFeeHex)
	ep := stringToBigInt(effectivePriceHex)

	if gu == nil || mf == nil || ep == nil {
		return ""
	}

	savingsPerGas := new(big.Int).Sub(mf, ep)
	if savingsPerGas.Sign() <= 0 {
		return ""
	}

	totalSavingsWei := new(big.Int).Mul(savingsPerGas, gu)
	savingsEth := weiToEth(totalSavingsWei)

	return fmt.Sprintf("%s ETH 💸", savingsEth.Text('f', -1))
}

// hexToDecimal converts a hex string to its decimal string representation.
func hexToDecimal(hexStr string) string {
	bi := stringToBigInt(hexStr)
	if bi == nil {
		return hexStr
	}
	return bi.String()
}

// calculateConfirmations calculates the number of confirmations for a transaction block.
func calculateConfirmations(latestBlock, txBlock string) string {
	if latestBlock == "" || txBlock == "" || txBlock == "0x0" {
		return ""
	}

	latest := stringToBigInt(latestBlock)
	tx := stringToBigInt(txBlock)

	if latest == nil || tx == nil {
		return "error"
	}

	diff := new(big.Int).Sub(latest, tx)
	if diff.Sign() < 0 {
		return "0"
	}

	conf := new(big.Int).Add(diff, big.NewInt(1))
	return conf.String()
}
