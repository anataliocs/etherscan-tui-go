package etherscan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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
					w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x65d507c0"}}`)) // 2024-02-20T20:12:48Z
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
				if tx.Confirmations != "10" {
					t.Errorf("Expected 10 confirmations, got %s", tx.Confirmations)
				}
			}

			if tt.name == "Success" || tt.name == "Success Repro Sepolia" {
				if tx.Confirmations != "1" {
					t.Errorf("Expected 1 confirmation, got %s", tx.Confirmations)
				}
			}

			if tt.name == "Success Repro Sepolia" {
				if tx.Type != "2 (EIP-1559)" {
					t.Errorf("Expected type '2 (EIP-1559)', got '%s'", tx.Type)
				}
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		hex      string
		expected string
	}{
		{"0xde0b6b3a7640000", "1 ETH"},   // 10^18
		{"0x1bc16d674ec80000", "2 ETH"},  // 2 * 10^18
		{"0x6f05b59d3b20000", "0.5 ETH"}, // 0.5 * 10^18
		{"0x0", "0 ETH"},
		{"", ""},
		{"123", "123"},
	}

	for _, tt := range tests {
		got := formatValue(tt.hex)
		if got != tt.expected {
			t.Errorf("formatValue(%s) = %s; want %s", tt.hex, got, tt.expected)
		}
	}
}

func TestFormatGasPrice(t *testing.T) {
	tests := []struct {
		hex      string
		expected string
	}{
		{"0x3b9aca00", "1 Gwei (0.000000001 ETH)"},
		{"0x77359400", "2 Gwei (0.000000002 ETH)"},
		{"0x1dcd6500", "0.5 Gwei (0.0000000005 ETH)"},
		{"0x0", "0 Gwei (0 ETH)"},
		{"", ""},
		{"123", "123"},
	}

	for _, tt := range tests {
		got := formatGasPrice(tt.hex)
		if got != tt.expected {
			t.Errorf("formatGasPrice(%s) = %s; want %s", tt.hex, got, tt.expected)
		}
	}
}

func TestFormatTransactionFee(t *testing.T) {
	tests := []struct {
		gasUsed  string
		gasPrice string
		expected string
	}{
		{"0x5208", "0x3b9aca00", "0.000021 ETH"}, // 21000 * 1 Gwei
		{"0x5208", "0x77359400", "0.000042 ETH"}, // 21000 * 2 Gwei
		{"0x0", "0x3b9aca00", "0 ETH"},
		{"0x5208", "0x0", "0 ETH"},
		{"", "0x3b9aca00", ""},
		{"0x5208", "", ""},
	}

	for _, tt := range tests {
		got := formatTransactionFee(tt.gasUsed, tt.gasPrice)
		if got != tt.expected {
			t.Errorf("formatTransactionFee(%s, %s) = %s; want %s", tt.gasUsed, tt.gasPrice, got, tt.expected)
		}
	}
}

func TestFormatTransactionType(t *testing.T) {
	tests := []struct {
		hex      string
		expected string
	}{
		{"0x0", "0 (Legacy)"},
		{"0x1", "1 (Access List)"},
		{"0x2", "2 (EIP-1559)"},
		{"0x02", "2 (EIP-1559)"},
		{"0x3", "3 (EIP-4844)"},
		{"0xa", "10"},
		{"", "0 (Legacy)"},
		{"0x", "0 (Legacy)"},
	}

	for _, tt := range tests {
		got := formatTransactionType(tt.hex)
		if got != tt.expected {
			t.Errorf("formatTransactionType(%s) = %s; want %s", tt.hex, got, tt.expected)
		}
	}
}

func TestFetchTransaction_ResultAsString(t *testing.T) {
	// This JSON simulates the case where 'result' is a string instead of an object.
	// We want to make sure we handle it gracefully and don't fail with "json: cannot unmarshal string..."
	jsonData := `{"jsonrpc":"2.0","id":1,"result":"Max rate limit reached"}`

	var proxyResp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	err := json.Unmarshal([]byte(jsonData), &proxyResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal outer response: %v", err)
	}

	var tx Transaction
	err = json.Unmarshal(proxyResp.Result, &tx)
	if err == nil {
		t.Fatal("Expected error when unmarshaling string into Transaction struct, but got nil")
	}

	// Now check how our logic would handle it
	var msg string
	if err := json.Unmarshal(proxyResp.Result, &msg); err != nil {
		t.Fatalf("Expected to be able to unmarshal result as string, but got: %v", err)
	}

	if msg != "Max rate limit reached" {
		t.Errorf("Expected 'Max rate limit reached', got '%s'", msg)
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

func TestFetchTransaction_Success(t *testing.T) {
	jsonData := `{"jsonrpc":"2.0","id":1,"result":{"hash":"0x123","blockNumber":"0x1"}}`

	var proxyResp struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal([]byte(jsonData), &proxyResp); err != nil {
		t.Fatalf("Failed to unmarshal outer: %v", err)
	}

	var tx Transaction
	if err := json.Unmarshal(proxyResp.Result, &tx); err != nil {
		t.Fatalf("Failed to unmarshal transaction: %v", err)
	}

	if tx.Hash != "0x123" {
		t.Errorf("Expected hash 0x123, got %s", tx.Hash)
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

			client := NewClient("test-api-key")
			client.baseURL = server.URL

			status, gasUsed, err := client.FetchTransactionReceipt(context.Background(), "0xabc")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if status != tt.expectedStatus {
				t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, status)
			}
			// In Success and Failed cases, we expect a value for gasUsed if provided in mock
			if tt.name == "Success" || tt.name == "Failed" {
				if gasUsed != "0x5208" {
					t.Errorf("Expected gasUsed 0x5208, got %s", gasUsed)
				}
			}
		})
	}
}

func TestFetchTransaction_RetryOnRateLimit(t *testing.T) {
	var callCount atomic.Int32
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount.Add(1)

		action := r.URL.Query().Get("action")
		switch action {
		case "eth_getTransactionByHash":
			// We only want to test retry for THIS call specifically in this test
			if callCount.Load() == 1 {
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"Max calls per sec rate limit reached"}`))
				return
			}
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0xabc","blockNumber":"0x1","type":"0x2"}}`))
		case "eth_blockNumber":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`))
		case "eth_getTransactionReceipt":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"status":"0x1","gasUsed":"0x5208"}}`))
		case "eth_getBlockByNumber":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x65d507c0"}}`))
		}
	})

	server := httptest.NewServer(mockHandler)
	defer server.Close()

	client := NewClient("test-api-key")
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := client.FetchTransaction(ctx, "0xabc")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 1 (failed getTransactionByHash) + 1 (success getTransactionByHash)
	// + 1 (blockNumber) + 1 (receipt) + 1 (block timestamp) = 5 calls
	if callCount.Load() != 5 {
		t.Errorf("Expected 5 calls to the API, got %d", callCount.Load())
	}

	if tx.Hash != "0xabc" {
		t.Errorf("Expected hash '0xabc', got '%s'", tx.Hash)
	}
}
