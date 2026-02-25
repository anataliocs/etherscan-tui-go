package etherscan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
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
	Type             string `json:"type"`
	Confirmations    string `json:"confirmations,omitzero"`
	Status           string `json:"status"`             // "Pending", "success", "failed", "dropped", "replaced"
	Timestamp        string `json:"timestamp,omitzero"` // ISO 8601 format
	GasUsed          string `json:"gasUsed"`
	TransactionFee   string `json:"transactionFee"`
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

	latestBlock, err := c.FetchLatestBlockNumber()
	if err == nil {
		tx.Confirmations = calculateConfirmations(latestBlock, hexBlockNumber)
	} else {
		tx.Confirmations = "error"
	}

	status, gasUsed, _ := c.FetchTransactionReceipt(hash)
	tx.Status = status
	tx.GasUsed = hexToDecimal(gasUsed)
	tx.TransactionFee = formatTransactionFee(gasUsed, hexGasPrice)

	if hexBlockNumber != "" && hexBlockNumber != "0x0" {
		timestamp, err := c.FetchBlockTimestamp(hexBlockNumber)
		if err == nil {
			tx.Timestamp = timestamp
		}
	}

	return &tx, nil
}

func (c *Client) FetchLatestBlockNumber() (string, error) {
	if c.apiKey == "" {
		return "", errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_blockNumber&apikey=%s", c.baseURL, c.chainId, c.apiKey)

	resp, err := c.http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var proxyResp struct {
		Result string `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &proxyResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if proxyResp.Error != nil {
		return "", errors.New(proxyResp.Error.Message)
	}

	if proxyResp.Result == "" {
		return "", errors.New("invalid block number response")
	}

	return proxyResp.Result, nil
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

func (c *Client) FetchBlockTimestamp(blockNumber string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getBlockByNumber&tag=%s&boolean=false&apikey=%s", c.baseURL, c.chainId, blockNumber, c.apiKey)

	resp, err := c.http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var proxyResp struct {
		Result struct {
			Timestamp string `json:"timestamp"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &proxyResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if proxyResp.Error != nil {
		return "", errors.New(proxyResp.Error.Message)
	}

	if proxyResp.Result.Timestamp == "" {
		return "", errors.New("timestamp not found in block")
	}

	// Parse hex timestamp
	var unixTime int64
	_, err = fmt.Sscanf(proxyResp.Result.Timestamp, "0x%x", &unixTime)
	if err != nil {
		return "", fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return time.Unix(unixTime, 0).UTC().Format(time.RFC3339), nil
}

func (c *Client) FetchTransactionReceipt(hash string) (string, string, error) {
	if c.apiKey == "" {
		return "", "", errors.New("ETHERSCAN_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("%s?chainid=%d&module=proxy&action=eth_getTransactionReceipt&txhash=%s&apikey=%s", c.baseURL, c.chainId, hash, c.apiKey)

	resp, err := c.http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var proxyResp struct {
		Result struct {
			Status  string `json:"status"`
			GasUsed string `json:"gasUsed"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &proxyResp); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	if proxyResp.Error != nil {
		return "", "", errors.New(proxyResp.Error.Message)
	}

	if string(body) == `{"result":null}` || string(body) == `{"result": null}` {
		return "Pending", "", nil
	}

	status := "Pending"
	if proxyResp.Result.Status == "0x1" {
		status = "success"
	} else if proxyResp.Result.Status == "0x0" {
		status = "failed"
	}

	return status, proxyResp.Result.GasUsed, nil
}

func formatValue(hexStr string) string {
	eth, s, done := hexToFloat(hexStr, 1e18)
	if done {
		return s
	}

	return fmt.Sprintf("%s ETH", eth.Text('f', -1))
}

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

func formatGasPrice(hexStr string) string {
	gwei, s, done := hexToFloat(hexStr, 1e9)
	if done {
		return s
	}

	eth, _, _ := hexToFloat(hexStr, 1e18)

	return fmt.Sprintf("%s Gwei (%s ETH)", gwei.Text('f', -1), eth.Text('f', -1))
}

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
