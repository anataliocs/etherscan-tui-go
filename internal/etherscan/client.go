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
		chainId: 1, // Default to Mainnet
	}
}

// SetChainID sets the Ethereum chain ID for the client.
// Parameters:
//   - id: The Ethereum chain ID (e.g., 1 for Mainnet, 11155111 for Sepolia).
func (c *Client) SetChainID(id int) {
	c.chainId = id
}

// ChainID returns the current Ethereum chain ID.
// Returns:
//   - The current Ethereum chain ID.
func (c *Client) ChainID() int {
	return c.chainId
}

// FetchTransaction fetches transaction details by its hash.
// Parameters:
//   - ctx: The context for the request.
//   - hash: The transaction hash to fetch.
//
// Returns:
//   - A pointer to the Transaction struct containing details.
//   - An error if the request fails or the transaction is not found.
func (c *Client) FetchTransaction(ctx context.Context, hash string) (*Transaction, error) {
	if c.apiKey == "" {
		return nil, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", c.baseURL, c.chainId, hash, c.apiKey)

	// small delay so the loading state is visible in the UI and to be polite with API
	transaction, err2, done := throttle(ctx)
	if done {
		return transaction, err2
	}

	proxyResp, err := doRequest[json.RawMessage](c, ctx, url)
	if err != nil {
		return nil, err
	}

	tx, t, err3 := buildTransaction(ctx, hash, proxyResp, err, c)
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
func throttle(ctx context.Context) (*Transaction, error, bool) {
	select {
	case <-time.After(500 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err(), true
	}
	return nil, nil, false
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

// FetchBlockDetails retrieves block timestamp, base fee and the list of transaction hashes for a given block number.
// Parameters:
//   - ctx: The context for the request.
//   - blockNumber: The block number (hex or tag) to fetch details for.
//
// Returns:
//   - The formatted timestamp string.
//   - The base fee per gas as a hex string.
//   - The list of transaction hashes in the block.
//   - An error if the request fails.
func (c *Client) FetchBlockDetails(ctx context.Context, blockNumber string) (string, string, []string, error) {
	if c.apiKey == "" {
		return "", "", nil, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getBlockByNumber&tag=%s&boolean=false&apikey=%s", c.baseURL, c.chainId, blockNumber, c.apiKey)

	proxyResp, err := doRequest[json.RawMessage](c, ctx, url)
	if err != nil {
		return "", "", nil, err
	}

	block, unixTime, _, _, err2 := extractBlockDetails(proxyResp, err)
	if err2 != nil {
		return "", "", nil, err2
	}

	return time.Unix(unixTime, 0).UTC().Format(time.RFC3339), block.BaseFeePerGas, block.Transactions, nil
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
	_, _, txHashes, err := c.FetchBlockDetails(ctx, fmt.Sprintf("0x%x", stringToBigInt(currentTx.BlockNumber)))
	if err == nil {
		for i, hash := range txHashes {
			if strings.EqualFold(hash, currentTx.Hash) {
				if i+1 < len(txHashes) {
					return txHashes[i+1], nil
				}
				break
			}
		}
	}

	// 2. If it's the last one or error fetching current block, try the next block
	nextBlockNum := new(big.Int).Add(stringToBigInt(currentTx.BlockNumber), big.NewInt(1))
	_, _, nextTxHashes, err := c.FetchBlockDetails(ctx, fmt.Sprintf("0x%x", nextBlockNum))
	if err != nil {
		return "", fmt.Errorf("could not fetch next block: %w", err)
	}

	if len(nextTxHashes) == 0 {
		return "", errors.New("no transactions found in the next block")
	}

	return nextTxHashes[0], nil
}

// IsContract checks if the given address is a smart contract.
// Parameters:
//   - ctx: The context for the request.
//   - address: The Ethereum address to check.
//
// Returns:
//   - A boolean indicating if the address is a contract.
//   - An error if the request fails.
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
func (c *Client) FetchTransactionReceipt(ctx context.Context, hash string) (string, string, string, error) {
	if c.apiKey == "" {
		return "", "", "", errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getTransactionReceipt&txhash=%s&apikey=%s", c.baseURL, c.chainId, hash, c.apiKey)

	proxyResp, err := doRequest[receiptResult](c, ctx, url)
	if err != nil {
		return "", "", "", err
	}

	status, s, s2, s3, err2, done := extractTransactionReceipt(proxyResp)
	if done {
		return s, s2, s3, err2
	}

	return status, proxyResp.Result.GasUsed, proxyResp.Result.EffectiveGasPrice, nil
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
