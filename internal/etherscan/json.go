// Package etherscan provides utilities for handling JSON responses from Etherscan.
package etherscan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// buildTransaction takes a raw transaction response and converts it to a Transaction struct.
// Parameters:
//   - ctx: The context for the request.
//   - hash: The transaction hash.
//   - proxyResp: The raw response from the Etherscan proxy.
//   - err: Any error that occurred during the initial request.
//   - c: The Etherscan client.
//
// Returns:
//   - The built Transaction struct.
//   - A pointer to Transaction (nil unless there's a specific return case).
//   - An error if building the transaction fails.
func buildTransaction(ctx context.Context, hash string, proxyResp *ProxyResponse[json.RawMessage], err error, c *Client) (Transaction, *Transaction, error) {
	if len(proxyResp.Result) == 0 || string(proxyResp.Result) == "null" {
		return Transaction{}, nil, errors.New("transaction not found or invalid response")
	}

	// Try to unmarshal Result as a Transaction object
	var tx Transaction
	if err := json.Unmarshal(proxyResp.Result, &tx); err != nil {
		// If it's not a Transaction object, check if it's a string (e.g., an error message)
		var msg string
		if json.Unmarshal(proxyResp.Result, &msg) == nil {
			// If the message contains "Error!" it's likely a transaction not found on this network
			if strings.Contains(msg, "Error!") {
				return Transaction{}, nil, fmt.Errorf("Etherscan API error: %s (Is the hash on the correct network?)", msg)
			}
			return Transaction{}, nil, fmt.Errorf("Etherscan API error: %s", msg)
		}
		return Transaction{}, nil, fmt.Errorf("unexpected response format for result: %w", err)
	}

	// Keep hex block number for timestamp fetching
	hexBlockNumber := tx.BlockNumber

	// Keep hex fields for fee calculation
	hexGasPrice := tx.GasPrice
	hexMaxFeePerGas := tx.MaxFeePerGas

	// Convert hex fields to decimal
	tx.BlockNumber = hexToDecimal(tx.BlockNumber)
	tx.Value = formatValue(tx.Value)
	tx.Gas = hexToDecimal(tx.Gas)
	tx.GasPrice = formatGasPrice(tx.GasPrice)
	tx.Nonce = hexToDecimal(tx.Nonce)
	tx.TransactionIndex = hexToDecimal(tx.TransactionIndex)
	tx.Type = formatTransactionType(tx.Type)

	latestBlock, err := c.FetchLatestBlockNumber(ctx)
	if err == nil {
		tx.Confirmations = calculateConfirmations(latestBlock, hexBlockNumber)
	} else {
		tx.Confirmations = err.Error()
	}

	status, gasUsed, effectiveGasPrice, _ := c.FetchTransactionReceipt(ctx, hash)
	tx.Status = status
	tx.GasUsed = hexToDecimal(gasUsed)
	tx.TransactionFee = formatTransactionFee(gasUsed, hexGasPrice)

	if hexMaxFeePerGas != "" {
		tx.Savings = calculateSavings(gasUsed, hexMaxFeePerGas, effectiveGasPrice)
	}

	if hexBlockNumber != "" && hexBlockNumber != "0x0" {
		timestamp, baseFee, _, err := c.FetchBlockDetails(ctx, hexBlockNumber)
		if err == nil {
			tx.Timestamp = timestamp
			tx.BaseFeePerGas = formatGwei(baseFee)
			tx.BurntFees = calculateBurntFees(gasUsed, baseFee)
		} else {
			tx.Timestamp = err.Error()
		}
	}

	if tx.MaxFeePerGas != "" {
		tx.MaxFeePerGas = formatGwei(tx.MaxFeePerGas)
	}
	if tx.MaxPriorityFeePerGas != "" {
		tx.MaxPriorityFeePerGas = formatGwei(tx.MaxPriorityFeePerGas)
	}

	// For legacy transactions, gas price = max fee = max priority fee (informally)
	// But Etherscan usually doesn't show them if they are not EIP-1559.
	// We'll leave them empty if not present in the original tx response.

	if tx.To != "" && tx.To != "0x0000000000000000000000000000000000000000" {
		isContract, err := c.IsContract(ctx, tx.To)
		if err == nil {
			if isContract {
				tx.ToAccountType = "Smart Contract"
			} else {
				tx.ToAccountType = "EOA"
			}
		}
	}
	return tx, nil, nil
}

// extractTransactionReceipt extracts status information from a transaction receipt.
// Parameters:
//   - proxyResp: The raw response from the Etherscan proxy for the receipt.
//
// Returns:
//   - The transaction status (success, failed, or Pending).
//   - An empty string (kept for signature compatibility).
//   - An empty string (kept for signature compatibility).
//   - An empty string (kept for signature compatibility).
//   - An error if extraction fails (currently always nil).
//   - A boolean indicating if the receipt is missing/pending.
func extractTransactionReceipt(proxyResp *ProxyResponse[receiptResult]) (string, string, string, string, error, bool) {
	if proxyResp.Result.Status == "" && proxyResp.Result.GasUsed == "" {
		return "", "Pending", "", "", nil, true
	}

	status := "Pending"
	if proxyResp.Result.Status == "0x1" {
		status = "success"
	} else if proxyResp.Result.Status == "0x0" {
		status = "failed"
	}
	return status, "", "", "", nil, false
}

// extractBlockDetails parses block details from a raw proxy response.
// Parameters:
//   - proxyResp: The raw response from the Etherscan proxy for the block.
//   - err: Any error that occurred during the initial request.
//
// Returns:
//   - A struct containing Timestamp and BaseFeePerGas.
//   - The Unix timestamp as an int64.
//   - An empty string (kept for signature compatibility).
//   - An empty string (kept for signature compatibility).
//   - An error if parsing fails.
func extractBlockDetails(proxyResp *ProxyResponse[json.RawMessage], err error) (struct {
	Timestamp     string   `json:"timestamp"`
	BaseFeePerGas string   `json:"baseFeePerGas"`
	Transactions  []string `json:"transactions"`
}, int64, string, string, error) {
	if len(proxyResp.Result) == 0 || string(proxyResp.Result) == "null" {
		return struct {
			Timestamp     string   `json:"timestamp"`
			BaseFeePerGas string   `json:"baseFeePerGas"`
			Transactions  []string `json:"transactions"`
		}{}, 0, "", "", errors.New("block not found")
	}

	var block struct {
		Timestamp     string   `json:"timestamp"`
		BaseFeePerGas string   `json:"baseFeePerGas"`
		Transactions  []string `json:"transactions"`
	}

	if err := json.Unmarshal(proxyResp.Result, &block); err != nil {
		var msg string
		if json.Unmarshal(proxyResp.Result, &msg) == nil {
			return struct {
				Timestamp     string   `json:"timestamp"`
				BaseFeePerGas string   `json:"baseFeePerGas"`
				Transactions  []string `json:"transactions"`
			}{}, 0, "", "", fmt.Errorf("Etherscan API error: %s", msg)
		}
		return struct {
			Timestamp     string   `json:"timestamp"`
			BaseFeePerGas string   `json:"baseFeePerGas"`
			Transactions  []string `json:"transactions"`
		}{}, 0, "", "", fmt.Errorf("unexpected response format for block: %w", err)
	}

	if block.Timestamp == "" {
		return struct {
			Timestamp     string   `json:"timestamp"`
			BaseFeePerGas string   `json:"baseFeePerGas"`
			Transactions  []string `json:"transactions"`
		}{}, 0, "", "", errors.New("timestamp not found in block")
	}

	lastTxHash := ""
	if len(block.Transactions) > 0 {
		lastTxHash = block.Transactions[len(block.Transactions)-1]
	}

	// Parse hex timestamp
	var unixTime int64
	_, err = fmt.Sscanf(block.Timestamp, "0x%x", &unixTime)
	if err != nil {
		return struct {
			Timestamp     string   `json:"timestamp"`
			BaseFeePerGas string   `json:"baseFeePerGas"`
			Transactions  []string `json:"transactions"`
		}{}, 0, "", "", fmt.Errorf("failed to parse timestamp: %w", err)
	}
	return block, unixTime, "", lastTxHash, nil
}
