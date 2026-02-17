package etherscan

import (
	"encoding/json"
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
			responseBody: `{"jsonrpc":"2.0","id":1,"result":{"hash":"0x123","blockNumber":"0x1"}}`,
			expectedHash: "0x123",
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
			// disable sleep for testing
			// Actually, we can't easily disable it without more refactoring or conditional compilation,
			// but 500ms per test case is acceptable for now.

			tx, err := client.FetchTransaction("0xabc")

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
		})
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
