package etherscan

import "net/http"

type Transaction struct {
	Hash                 string `json:"hash"`
	BlockNumber          string `json:"blockNumber"`
	From                 string `json:"from"`
	To                   string `json:"to"`
	Value                string `json:"value"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	Nonce                string `json:"nonce"`
	TransactionIndex     string `json:"transactionIndex"`
	Input                string `json:"input"`
	Type                 string `json:"type"`
	Confirmations        string `json:"confirmations,omitzero"`
	Status               string `json:"status"`             // "Pending", "success", "failed", "dropped", "replaced"
	Timestamp            string `json:"timestamp,omitzero"` // ISO 8601 format
	GasUsed              string `json:"gasUsed"`
	TransactionFee       string `json:"transactionFee"`
	ToAccountType        string `json:"toAccountType,omitzero"` // "EOA" or "Smart Contract"
	MaxFeePerGas         string `json:"maxFeePerGas,omitzero"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitzero"`
	BaseFeePerGas        string `json:"baseFeePerGas,omitzero"`
	BurntFees            string `json:"burntFees,omitzero"`
}

type Client struct {
	apiKey  string
	http    *http.Client
	baseURL string
	chainId int
}
