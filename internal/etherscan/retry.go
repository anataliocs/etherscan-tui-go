// Package etherscan provides retry logic for API requests.
package etherscan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// doRequestWithRetry performs an HTTP GET request with exponential backoff retries.
// Parameters:
//   - ctx: The context for the request.
//   - url: The URL to fetch.
//
// Returns:
//   - The response body as a byte slice.
//   - An error if all retry attempts fail or the context is cancelled.
func (c *Client) doRequestWithRetry(ctx context.Context, url string) ([]byte, error) {
	maxRetries := 3
	var lastErr error

	for i := range maxRetries + 1 {
		if i > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(i-1)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Check for rate limit error in body
		bodyString := string(body)
		if strings.Contains(bodyString, "Max calls per sec rate limit reached") || strings.Contains(bodyString, "rate limit") {
			lastErr = fmt.Errorf("Etherscan API error: %s", strings.TrimSpace(bodyString))
			if strings.Contains(bodyString, "{") {
				// If it's JSON, try to extract message
				var proxyResp ProxyResponse[json.RawMessage]
				if json.Unmarshal(body, &proxyResp) == nil {
					if proxyResp.Error != nil {
						lastErr = fmt.Errorf("Etherscan API error: %s", proxyResp.Error.Message)
					} else {
						var msg string
						if json.Unmarshal(proxyResp.Result, &msg) == nil {
							lastErr = fmt.Errorf("Etherscan API error: %s", msg)
						}
					}
				}
			}
			continue
		}

		return body, nil
	}

	return nil, lastErr
}
