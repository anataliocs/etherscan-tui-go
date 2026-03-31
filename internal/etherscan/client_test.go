// Package etherscan contains tests for the Etherscan client.
package etherscan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchTransaction_MockAPI(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		expectedErr  string
		expectedHash string
	}{
		{
			name:         "Success",
			responseBody: `{"jsonrpc":"2.0","id":1,"result":{"hash":"0x123","blockNumber":"0xb","type":"0x2"}}`,
			expectedHash: "0x123",
		},
		{
			name:         "Success With Timestamp",
			responseBody: `{"jsonrpc":"2.0","id":1,"result":{"hash":"0x456","blockNumber":"0x2"}}`,
			expectedHash: "0x456",
		},
		{
			name:         "Rate Limit Error (String Result)",
			responseBody: `{"jsonrpc":"2.0","id":1,"result":"Max rate limit reached"}`,
			expectedErr:  "Etherscan API error: Max rate limit reached",
		},
		{
			name:         "Explicit Error Object",
			responseBody: `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"Resource not found"}}`,
			expectedErr:  "Resource not found",
		},
		{
			name:         "Empty Result",
			responseBody: `{"jsonrpc":"2.0","id":1,"result":null}`,
			expectedErr:  "transaction not found or invalid response",
		},
		{
			name:         "Hash Not Found Error (String Result)",
			responseBody: `{"jsonrpc":"2.0","id":1,"result":"Error! Transaction hash not found"}`,
			expectedErr:  "Etherscan API error: Error! Transaction hash not found (Is the hash on the correct network?)",
		},
		{
			name:         "Success Repro Sepolia",
			responseBody: `{"jsonrpc":"2.0","id":1,"result":{"hash":"0xe16e8b72443aaee9c3d4ec42ecd973dc7faf583475f66d5a7ac9ebcce72b32c8","blockNumber":"0x63ef52","type":"0x2"}}`,
			expectedHash: "0xe16e8b72443aaee9c3d4ec42ecd973dc7faf583475f66d5a7ac9ebcce72b32c8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				action := r.URL.Query().Get("action")
				switch action {
				case "eth_getTransactionByHash":
					w.Write([]byte(tt.responseBody))
				case "eth_getBlockByNumber":
					w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x65d507c0", "transactions": ["0x123", "0x456"]}}`)) // 2024-02-20T20:12:48Z
				case "eth_getTransactionReceipt":
					w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"status":"0x1","gasUsed":"0x5208"}}`)) // 21000
				case "eth_blockNumber":
					if tt.name == "Success Repro Sepolia" {
						w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x63ef52"}`))
					} else {
						w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xb"}`)) // 11
					}
				default:
					w.Write([]byte(tt.responseBody))
				}
			})

			server := httptest.NewServer(mockHandler)
			defer server.Close()

			client := NewClient("test-api-key")
			client.baseURL = server.URL

			tx, err := client.FetchTransaction(context.Background(), "0xabc")

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tx.Hash != tt.expectedHash {
				t.Errorf("Expected hash '%s', got '%s'", tt.expectedHash, tx.Hash)
			}

			if tt.name == "Success" {
				if tx.BlockNumber != "11" {
					t.Errorf("Expected block number '11', got '%s'", tx.BlockNumber)
				}
				if tx.Type != "2 (EIP-1559)" {
					t.Errorf("Expected type '2 (EIP-1559)', got '%s'", tx.Type)
				}
			}

			if tt.name == "Success With Timestamp" {
				expectedTimestamp := "2024-02-20T20:12:48Z"
				if tx.Timestamp != expectedTimestamp {
					t.Errorf("Expected timestamp '%s', got '%s'", expectedTimestamp, tx.Timestamp)
				}
			}
		})
	}
}

func TestClient_ChainID(t *testing.T) {
	client := NewClient("test")
	if client.ChainID() != 1 {
		t.Errorf("Expected default chain ID 1, got %d", client.ChainID())
	}

	client.SetChainID(11155111)
	if client.ChainID() != 11155111 {
		t.Errorf("Expected chain ID 11155111, got %d", client.ChainID())
	}
}

func TestFetchTransactionReceipt(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		expectedStatus string
	}{
		{
			name:           "Success",
			responseBody:   `{"jsonrpc":"2.0","id":1,"result":{"status":"0x1","gasUsed":"0x5208"}}`,
			expectedStatus: "success",
		},
		{
			name:           "Failed",
			responseBody:   `{"jsonrpc":"2.0","id":1,"result":{"status":"0x0","gasUsed":"0x5208"}}`,
			expectedStatus: "failed",
		},
		{
			name:           "Pending",
			responseBody:   `{"jsonrpc":"2.0","id":1,"result":null}`,
			expectedStatus: "Pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("test")
			client.baseURL = server.URL

			status, _, _, err := client.FetchTransactionReceipt(context.Background(), "0xabc")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}
