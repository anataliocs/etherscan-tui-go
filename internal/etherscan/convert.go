package etherscan

import (
	"fmt"
	"math/big"
	"strings"
)

func hexToFloat(hexStr string, val float64) (*big.Float, string, bool) {
	if hexStr == "" || !strings.HasPrefix(hexStr, "0x") {
		return nil, hexStr, true
	}

	trimmed := strings.TrimPrefix(hexStr, "0x")
	if trimmed == "" {
		return nil, "0 ETH", true
	}

	bi := new(big.Int)
	if _, ok := bi.SetString(trimmed, 16); !ok {
		return nil, hexStr, true
	}

	// 1 ETH = 10^18 Wei
	eth := new(big.Float).SetInt(bi)
	eth.Quo(eth, big.NewFloat(val))
	return eth, "", false
}

func calculateBurntFees(gasUsedHex, baseFeeHex string) string {
	if gasUsedHex == "" || baseFeeHex == "" {
		return ""
	}

	gu := new(big.Int)
	if _, ok := gu.SetString(strings.TrimPrefix(gasUsedHex, "0x"), 16); !ok {
		return ""
	}

	bf := new(big.Int)
	if _, ok := bf.SetString(strings.TrimPrefix(baseFeeHex, "0x"), 16); !ok {
		return ""
	}

	// Burnt Fees = gasUsed * baseFee
	burntWei := new(big.Int).Mul(gu, bf)

	// 1 ETH = 10^18 Wei
	burntEth := new(big.Float).SetInt(burntWei)
	burntEth.Quo(burntEth, big.NewFloat(1e18))

	return fmt.Sprintf("%s ETH 🔥", burntEth.Text('f', -1))
}

func hexToDecimal(hexStr string) string {
	if hexStr == "" || !strings.HasPrefix(hexStr, "0x") {
		return hexStr
	}

	trimmed := strings.TrimPrefix(hexStr, "0x")
	if trimmed == "" {
		return "0"
	}

	// Use big.Int as Ethereum values can exceed uint64 (e.g., Value in Wei)
	bi := new(big.Int)
	if _, ok := bi.SetString(trimmed, 16); !ok {
		return hexStr
	}

	return bi.String()
}

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

	// confirmations = latest - tx + 1
	conf := new(big.Int).Add(diff, big.NewInt(1))
	return conf.String()
}

func stringToBigInt(s string) *big.Int {
	bi := new(big.Int)
	base := 10
	trimmed := s
	if strings.HasPrefix(s, "0x") {
		base = 16
		trimmed = strings.TrimPrefix(s, "0x")
	}

	if _, ok := bi.SetString(trimmed, base); !ok {
		return nil
	}
	return bi
}
