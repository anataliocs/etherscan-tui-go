package etherscan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoRequestWithRetry(t *testing.T) {
	attempts := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			// Simulate rate limit
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"Max calls per sec rate limit reached"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"OK"}`))
	}))
	defer server.Close()

	client := NewClient("test")
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	body, err := client.doRequestWithRetry(ctx, server.URL)
	if err != nil {
		t.Fatalf("doRequestWithRetry failed: %v", err)
	}

	if !strings.Contains(string(body), "OK") {
		t.Errorf("expected body to contain OK, got %s", string(body))
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
