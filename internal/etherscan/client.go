package etherscan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Transaction struct {
	Hash             string `json:"hash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Nonce            string `json:"nonce"`
	TransactionIndex string `json:"transactionIndex"`
	Input            string `json:"input"`
}

type Client struct {
	apiKey  string
	http    *http.Client
	baseURL string
	chainId int
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

func (c *Client) FetchTransaction(hash string) (*Transaction, error) {
	if c.apiKey == "" {
		return nil, errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", c.baseURL, c.chainId, hash, c.apiKey)

	// small delay so the loading state is visible in the UI and to be polite with API
	time.Sleep(500 * time.Millisecond)

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var proxyResp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &proxyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if proxyResp.Error != nil {
		return nil, errors.New(proxyResp.Error.Message)
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

	return &tx, nil
}
