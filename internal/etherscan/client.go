package etherscan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ProxyResponse[T any] struct {
	Result T `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: "https://api.etherscan.io/v2/api",
		chainId: 1, // Default to Mainnet
	}
}

func (c *Client) SetChainID(id int) {
	c.chainId = id
}

func (c *Client) ChainID() int {
	return c.chainId
}

func (c *Client) FetchTransaction(ctx context.Context, hash string) (*Transaction, error) {
	if c.apiKey == "" {
		return nil, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", c.baseURL, c.chainId, hash, c.apiKey)

	// small delay so the loading state is visible in the UI and to be polite with API
	select {
	case <-time.After(500 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	proxyResp, err := doRequest[json.RawMessage](c, ctx, url)
	if err != nil {
		return nil, err
	}

	if len(proxyResp.Result) == 0 || string(proxyResp.Result) == "null" {
		return nil, errors.New("transaction not found or invalid response")
	}

	// Try to unmarshal Result as a Transaction object
	var tx Transaction
	if err := json.Unmarshal(proxyResp.Result, &tx); err != nil {
		// If it's not a Transaction object, check if it's a string (e.g., an error message)
		var msg string
		if json.Unmarshal(proxyResp.Result, &msg) == nil {
			// If the message contains "Error!" it's likely a transaction not found on this network
			if strings.Contains(msg, "Error!") {
				return nil, fmt.Errorf("Etherscan API error: %s (Is the hash on the correct network?)", msg)
			}
			return nil, fmt.Errorf("Etherscan API error: %s", msg)
		}
		return nil, fmt.Errorf("unexpected response format for result: %w", err)
	}

	// Keep hex block number for timestamp fetching
	hexBlockNumber := tx.BlockNumber

	// Keep hex fields for fee calculation
	hexGasPrice := tx.GasPrice

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

	status, gasUsed, _, _ := c.FetchTransactionReceipt(ctx, hash)
	tx.Status = status
	tx.GasUsed = hexToDecimal(gasUsed)
	tx.TransactionFee = formatTransactionFee(gasUsed, hexGasPrice)

	if hexBlockNumber != "" && hexBlockNumber != "0x0" {
		timestamp, baseFee, err := c.FetchBlockDetails(ctx, hexBlockNumber)
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

	return &tx, nil
}

func (c *Client) doRequestWithRetry(ctx context.Context, url string) ([]byte, error) {
	maxRetries := 3
	var lastErr error

	for i := range maxRetries + 1 {
		if i > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(i-1)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Check for rate limit error in body
		bodyString := string(body)
		if strings.Contains(bodyString, "Max calls per sec rate limit reached") || strings.Contains(bodyString, "rate limit") {
			lastErr = fmt.Errorf("Etherscan API error: %s", strings.TrimSpace(bodyString))
			if strings.Contains(bodyString, "{") {
				// If it's JSON, try to extract message
				var proxyResp ProxyResponse[json.RawMessage]
				if json.Unmarshal(body, &proxyResp) == nil {
					if proxyResp.Error != nil {
						lastErr = fmt.Errorf("Etherscan API error: %s", proxyResp.Error.Message)
					} else {
						var msg string
						if json.Unmarshal(proxyResp.Result, &msg) == nil {
							lastErr = fmt.Errorf("Etherscan API error: %s", msg)
						}
					}
				}
			}
			continue
		}

		return body, nil
	}

	return nil, lastErr
}

func (c *Client) FetchLatestBlockNumber(ctx context.Context) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_blockNumber&apikey=%s", c.baseURL, c.chainId, c.apiKey)

	proxyResp, err := doRequest[string](c, ctx, url)
	if err != nil {
		return "", err
	}

	if proxyResp.Result == "" {
		return "", errors.New("invalid block number response")
	}

	return proxyResp.Result, nil
}

func (c *Client) FetchBlockDetails(ctx context.Context, blockNumber string) (string, string, error) {
	if c.apiKey == "" {
		return "", "", errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getBlockByNumber&tag=%s&boolean=false&apikey=%s", c.baseURL, c.chainId, blockNumber, c.apiKey)

	proxyResp, err := doRequest[json.RawMessage](c, ctx, url)
	if err != nil {
		return "", "", err
	}

	if len(proxyResp.Result) == 0 || string(proxyResp.Result) == "null" {
		return "", "", errors.New("block not found")
	}

	var block struct {
		Timestamp     string `json:"timestamp"`
		BaseFeePerGas string `json:"baseFeePerGas"`
	}

	if err := json.Unmarshal(proxyResp.Result, &block); err != nil {
		var msg string
		if json.Unmarshal(proxyResp.Result, &msg) == nil {
			return "", "", fmt.Errorf("Etherscan API error: %s", msg)
		}
		return "", "", fmt.Errorf("unexpected response format for block: %w", err)
	}

	if block.Timestamp == "" {
		return "", "", errors.New("timestamp not found in block")
	}

	// Parse hex timestamp
	var unixTime int64
	_, err = fmt.Sscanf(block.Timestamp, "0x%x", &unixTime)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return time.Unix(unixTime, 0).UTC().Format(time.RFC3339), block.BaseFeePerGas, nil
}

func (c *Client) IsContract(ctx context.Context, address string) (bool, error) {
	if c.apiKey == "" {
		return false, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getCode&address=%s&tag=latest&apikey=%s", c.baseURL, c.chainId, address, c.apiKey)

	proxyResp, err := doRequest[string](c, ctx, url)
	if err != nil {
		return false, err
	}

	// eth_getCode returns "0x" if the address is an EOA
	return proxyResp.Result != "0x" && proxyResp.Result != "" && proxyResp.Result != "null", nil
}

func (c *Client) FetchTransactionReceipt(ctx context.Context, hash string) (string, string, string, error) {
	if c.apiKey == "" {
		return "", "", "", errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getTransactionReceipt&txhash=%s&apikey=%s", c.baseURL, c.chainId, hash, c.apiKey)

	type receiptResult struct {
		Status            string `json:"status"`
		GasUsed           string `json:"gasUsed"`
		EffectiveGasPrice string `json:"effectiveGasPrice"`
	}

	proxyResp, err := doRequest[receiptResult](c, ctx, url)
	if err != nil {
		return "", "", "", err
	}

	if proxyResp.Result.Status == "" && proxyResp.Result.GasUsed == "" {
		return "Pending", "", "", nil
	}

	status := "Pending"
	if proxyResp.Result.Status == "0x1" {
		status = "success"
	} else if proxyResp.Result.Status == "0x0" {
		status = "failed"
	}

	return status, proxyResp.Result.GasUsed, proxyResp.Result.EffectiveGasPrice, nil
}

func doRequest[T any](c *Client, ctx context.Context, url string) (*ProxyResponse[T], error) {
	body, err := c.doRequestWithRetry(ctx, url)
	if err != nil {
		return nil, err
	}

	var proxyResp ProxyResponse[T]
	if err := json.Unmarshal(body, &proxyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if proxyResp.Error != nil {
		return nil, errors.New(proxyResp.Error.Message)
	}

	return &proxyResp, nil
}
