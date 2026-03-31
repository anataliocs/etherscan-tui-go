package etherscan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractTransactionReceipt(t *testing.T) {
	tests := []struct {
		name            string
		proxyResp       *ProxyResponse[receiptResult]
		expectedStatus  string
		expectedPending bool
	}{
		{
			name: "Success",
			proxyResp: &ProxyResponse[receiptResult]{
				Result: receiptResult{Status: "0x1", GasUsed: "0x5208"},
			},
			expectedStatus:  "success",
			expectedPending: false,
		},
		{
			name: "Failed",
			proxyResp: &ProxyResponse[receiptResult]{
				Result: receiptResult{Status: "0x0", GasUsed: "0x5208"},
			},
			expectedStatus:  "failed",
			expectedPending: false,
		},
		{
			name: "Pending",
			proxyResp: &ProxyResponse[receiptResult]{
				Result: receiptResult{Status: "", GasUsed: ""},
			},
			expectedStatus:  "",
			expectedPending: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, _, _, err, pending := extractTransactionReceipt(tt.proxyResp)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status != tt.expectedStatus {
				t.Errorf("status = %s; want %s", status, tt.expectedStatus)
			}
			if pending != tt.expectedPending {
				t.Errorf("pending = %v; want %v", pending, tt.expectedPending)
			}
		})
	}
}

func TestExtractBlockDetails(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		expectedErr   string
		expectedTime  int64
		expectedBaseF string
	}{
		{
			name:          "Success",
			json:          `{"timestamp":"0x65d507c0", "baseFeePerGas":"0x7", "transactions":["0x123", "0x456"]}`,
			expectedTime:  1708459968,
			expectedBaseF: "0x7",
		},
		{
			name:        "EmptyResult",
			json:        `null`,
			expectedErr: "block not found",
		},
		{
			name:        "MissingTimestamp",
			json:        `{"baseFeePerGas":"0x7"}`,
			expectedErr: "timestamp not found in block",
		},
		{
			name:        "InvalidTimestamp",
			json:        `{"timestamp":"invalid"}`,
			expectedErr: "failed to parse timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyResp := &ProxyResponse[json.RawMessage]{
				Result: json.RawMessage(tt.json),
			}
			block, unixTime, _, txHash, err := extractBlockDetails(proxyResp, nil)

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %s, got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("error %v does not contain %s", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if unixTime != tt.expectedTime {
				t.Errorf("unixTime = %d; want %d", unixTime, tt.expectedTime)
			}
			if block.BaseFeePerGas != tt.expectedBaseF {
				t.Errorf("BaseFeePerGas = %s; want %s", block.BaseFeePerGas, tt.expectedBaseF)
			}
			if tt.name == "Success" && txHash != "0x456" {
				t.Errorf("txHash = %s; want 0x456", txHash)
			}
		})
	}
}

func TestBuildTransaction(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		action := r.URL.Query().Get("action")
		switch action {
		case "eth_blockNumber":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xc"}`)) // 12
		case "eth_getTransactionReceipt":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"status":"0x1","gasUsed":"0x5208", "effectiveGasPrice":"0x3b9aca00"}}`))
		case "eth_getBlockByNumber":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x65d507c0", "baseFeePerGas":"0x7"}}`))
		case "eth_getCode":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1234"}`)) // is a contract
		default:
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
		}
	})

	server := httptest.NewServer(mockHandler)
	defer server.Close()

	client := NewClient("test")
	client.baseURL = server.URL

	proxyResp := &ProxyResponse[json.RawMessage]{
		Result: json.RawMessage(`{"hash":"0xabc","blockNumber":"0xa","value":"0xde0b6b3a7640000","gas":"0x5208","gasPrice":"0x3b9aca00","nonce":"0x1","transactionIndex":"0x0","type":"0x2","to":"0x123","maxFeePerGas":"0x4b9aca00"}`),
	}

	tx, _, err := buildTransaction(context.Background(), "0xabc", proxyResp, nil, client)
	if err != nil {
		t.Fatalf("buildTransaction failed: %v", err)
	}

	if tx.Hash != "0xabc" {
		t.Errorf("expected hash 0xabc, got %s", tx.Hash)
	}
	if tx.BlockNumber != "10" {
		t.Errorf("expected block number 10, got %s", tx.BlockNumber)
	}
	if tx.Confirmations != "3" { // 12 - 10 + 1 = 3
		t.Errorf("expected 3 confirmations, got %s", tx.Confirmations)
	}
	if tx.Status != "success" {
		t.Errorf("expected status success, got %s", tx.Status)
	}
	if tx.ToAccountType != "Smart Contract" {
		t.Errorf("expected Smart Contract, got %s", tx.ToAccountType)
	}
	if !strings.Contains(tx.Savings, "ETH") {
		t.Errorf("expected savings to contain ETH, got %s", tx.Savings)
	}
}
