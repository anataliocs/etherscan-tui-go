// Package etherscan provides a client for interacting with the Etherscan API.
package etherscan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

// ProxyResponse is a generic struct for handling Etherscan proxy responses.
type ProxyResponse[T any] struct {
	Result T `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a new Etherscan client with the provided API key.
// Parameters:
//   - apiKey: The Etherscan API key to use for requests.
//
// Returns:
//   - A pointer to the newly created Client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: "https://api.etherscan.io/v2/api",
		chainID: 1, // Default to Mainnet
	}
}

// SetChainID sets the Ethereum chain ID for the client.
// Parameters:
//   - id: The Ethereum chain ID (e.g., 1 for Mainnet, 11155111 for Sepolia).
func (c *Client) SetChainID(id int) {
	c.chainID = id
}

// ChainID returns the current Ethereum chain ID.
// Returns:
//   - The current Ethereum chain ID.
func (c *Client) ChainID() int {
	return c.chainID
}

// FetchTransaction fetches transaction details by its hash.
// Parameters:
//   - ctx: The context for the request.
//   - hash: The transaction hash to fetch.
//
// Returns:
//   - A pointer to the Transaction struct containing details.
//   - An error if the request fails or the transaction is not found.
func (c *Client) FetchTransaction(ctx context.Context, hash Hash) (*Transaction, error) {
	if c.apiKey == "" {
		return nil, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", c.baseURL, c.chainID, hash, c.apiKey)

	// small delay so the loading state is visible in the UI and to be polite with API
	transaction, done, err2 := throttle(ctx)
	if done {
		return transaction, err2
	}

	proxyResp, err := doRequest[json.RawMessage](ctx, c, url)
	if err != nil {
		return nil, err
	}

	tx, t, err3 := buildTransaction(ctx, hash, proxyResp, c)
	if err3 != nil {
		return t, err3
	}

	return &tx, nil
}

// throttle introduces a small delay to be polite with the Etherscan API.
// Parameters:
//   - ctx: The context for the request.
//
// Returns:
//   - A pointer to Transaction (always nil in this implementation).
//   - An error if the context is cancelled.
//   - A boolean indicating if the request should be considered done (e.g., on context cancellation).
func throttle(ctx context.Context) (*Transaction, bool, error) {
	select {
	case <-time.After(500 * time.Millisecond):
	case <-ctx.Done():
		return nil, true, ctx.Err()
	}
	return nil, false, nil
}

// FetchLatestBlockNumber retrieves the latest block number from Etherscan.
// Parameters:
//   - ctx: The context for the request.
//
// Returns:
//   - The latest block number as a hex string.
//   - An error if the request fails.
func (c *Client) FetchLatestBlockNumber(ctx context.Context) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_blockNumber&apikey=%s", c.baseURL, c.chainID, c.apiKey)

	proxyResp, err := doRequest[string](ctx, c, url)
	if err != nil {
		return "", err
	}

	if proxyResp.Result == "" {
		return "", errors.New("invalid block number response")
	}

	return proxyResp.Result, nil
}

// FetchBlockDetails retrieves block details.
func (c *Client) FetchBlockDetails(ctx context.Context, blockNumber string) (*BlockDetails, error) {
	if c.apiKey == "" {
		return nil, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	tag := blockNumber
	if bi := stringToBigInt(blockNumber); bi != nil {
		tag = fmt.Sprintf("0x%x", bi)
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getBlockByNumber&tag=%s&boolean=false&apikey=%s", c.baseURL, c.chainID, tag, c.apiKey)

	proxyResp, err := doRequest[json.RawMessage](ctx, c, url)
	if err != nil {
		return nil, err
	}

	block, unixTime, miner, _, err2 := extractBlockDetails(proxyResp)
	if err2 != nil {
		return nil, err2
	}

	return &BlockDetails{
		Timestamp:     time.Unix(unixTime, 0).UTC().Format(time.RFC3339),
		BaseFeePerGas: block.BaseFeePerGas,
		Transactions:  block.Transactions,
		Miner:         miner,
		Status:        "Unfinalized", // Defaulting for now
	}, nil
}

// FetchNextTransactionHash attempts to find the next transaction hash after the given one in the same block.
// If it's the last transaction in the block, it tries the first transaction of the next block.
// Parameters:
//   - ctx: The context for the request.
//   - currentTx: The current transaction object.
//
// Returns:
//   - The next transaction hash.
//   - An error if the next transaction cannot be found.
func (c *Client) FetchNextTransactionHash(ctx context.Context, currentTx *Transaction) (string, error) {
	if currentTx == nil || currentTx.BlockNumber == "" {
		return "", errors.New("invalid current transaction")
	}

	// 1. Try to find the next transaction in the current block
	details, err := c.FetchBlockDetails(ctx, fmt.Sprintf("0x%x", stringToBigInt(currentTx.BlockNumber)))
	if err == nil {
		for i, hash := range details.Transactions {
			if strings.EqualFold(hash, string(currentTx.Hash)) {
				if i+1 < len(details.Transactions) {
					return details.Transactions[i+1], nil
				}
				break
			}
		}
	}

	// 2. If it's the last one or error fetching current block, try the next block
	nextBlockNum := new(big.Int).Add(stringToBigInt(currentTx.BlockNumber), big.NewInt(1))
	nextDetails, err := c.FetchBlockDetails(ctx, fmt.Sprintf("0x%x", nextBlockNum))
	if err != nil {
		return "", fmt.Errorf("could not fetch next block: %w", err)
	}

	if len(nextDetails.Transactions) == 0 {
		return "", errors.New("no transactions found in the next block")
	}

	return nextDetails.Transactions[0], nil
}

// FetchPreviousTransactionHash attempts to find the previous transaction hash before the given one in the same block.
// If it's the first transaction in the block, it tries the last transaction of the previous block.
// Parameters:
//   - ctx: The context for the request.
//   - currentTx: The current transaction object.
//
// Returns:
//   - The previous transaction hash.
//   - An error if the previous transaction cannot be found.
func (c *Client) FetchPreviousTransactionHash(ctx context.Context, currentTx *Transaction) (string, error) {
	if currentTx == nil || currentTx.BlockNumber == "" {
		return "", errors.New("invalid current transaction")
	}

	// 1. Try to find the previous transaction in the current block
	details, err := c.FetchBlockDetails(ctx, fmt.Sprintf("0x%x", stringToBigInt(currentTx.BlockNumber)))
	if err == nil {
		for i, hash := range details.Transactions {
			if strings.EqualFold(hash, string(currentTx.Hash)) {
				if i > 0 {
					return details.Transactions[i-1], nil
				}
				break
			}
		}
	}

	// 2. If it's the first one or error fetching current block, try the previous block
	prevBlockNum := new(big.Int).Sub(stringToBigInt(currentTx.BlockNumber), big.NewInt(1))
	if prevBlockNum.Sign() < 0 {
		return "", errors.New("already at block 0")
	}

	prevDetails, err := c.FetchBlockDetails(ctx, fmt.Sprintf("0x%x", prevBlockNum))
	if err != nil {
		return "", fmt.Errorf("could not fetch previous block: %w", err)
	}

	if len(prevDetails.Transactions) == 0 {
		return "", errors.New("no transactions found in the previous block")
	}

	return prevDetails.Transactions[len(prevDetails.Transactions)-1], nil
}

// IsContract checks if the given address is a smart contract.
// Parameters:
//   - ctx: The context for the request.
//   - address: The Ethereum address to check.
//
// Returns:
//   - A boolean indicating if the address is a contract.
//   - An error if the request fails.
func (c *Client) IsContract(ctx context.Context, address Address) (bool, error) {
	if c.apiKey == "" {
		return false, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getCode&address=%s&tag=latest&apikey=%s", c.baseURL, c.chainID, address, c.apiKey)

	proxyResp, err := doRequest[string](ctx, c, url)
	if err != nil {
		return false, err
	}

	// eth_getCode returns "0x" if the address is an EOA
	return proxyResp.Result != "0x" && proxyResp.Result != "" && proxyResp.Result != "null", nil
}

// FetchTransactionReceipt retrieves the receipt for a transaction by its hash.
// Parameters:
//   - ctx: The context for the request.
//   - hash: The transaction hash to fetch the receipt for.
//
// Returns:
//   - The status of the transaction (e.g., "success", "failed").
//   - The gas used by the transaction (hex).
//   - The effective gas price (hex).
//   - An error if the request fails.
func (c *Client) FetchTransactionReceipt(ctx context.Context, hash Hash) (string, string, string, bool, error) {
	if c.apiKey == "" {
		return "", "", "", false, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getTransactionReceipt&txhash=%s&apikey=%s", c.baseURL, c.chainID, hash, c.apiKey)

	proxyResp, err := doRequest[receiptResultData](ctx, c, url)
	if err != nil {
		return "", "", "", false, err
	}

	status, s, s2, s3, done, err2 := extractTransactionReceipt(proxyResp)
	if done {
		return s, s2, s3, done, err2
	}

	return status, proxyResp.Result.GasUsed, proxyResp.Result.EffectiveGasPrice, false, nil
}

// doRequest is a helper function that performs a generic Etherscan API request.
// Parameters:
//   - c: The Etherscan client.
//   - ctx: The context for the request.
//   - url: The full URL for the request.
//
// Returns:
//   - A pointer to the generic ProxyResponse[T] struct.
//   - An error if the request or unmarshaling fails.
func doRequest[T any](ctx context.Context, c *Client, url string) (*ProxyResponse[T], error) {
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
