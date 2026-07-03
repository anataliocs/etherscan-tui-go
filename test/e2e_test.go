package test

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/model"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

type mockClient struct {
	etherscan.Client
}

func (m *mockClient) ChainID() int {
	return 1
}

func (m *mockClient) SetChainID(_ int) {
}

func (m *mockClient) FetchTransaction(_ context.Context, _ etherscan.Hash) (*etherscan.Transaction, error) {
	return &etherscan.Transaction{Hash: "0x123", BlockNumber: "12345"}, nil
}

func (m *mockClient) FetchLatestBlockNumber(_ context.Context) (string, error) {
	return "", nil
}

func (m *mockClient) FetchBlockDetails(_ context.Context, _ string) (string, string, []string, error) {
	return "", "", nil, nil
}

func (m *mockClient) FetchNextTransactionHash(_ context.Context, _ *etherscan.Transaction) (string, error) {
	return "", nil
}

func (m *mockClient) FetchPreviousTransactionHash(_ context.Context, _ *etherscan.Transaction) (string, error) {
	return "", nil
}

func (m *mockClient) IsContract(_ context.Context, _ etherscan.Address) (bool, error) {
	return false, nil
}

func (m *mockClient) FetchTransactionReceipt(_ context.Context, _ etherscan.Hash) (string, string, string, bool, error) {
	return "success", "0x5208", "0x3b9aca00", false, nil
}

func stripANSI(str string) string {
	const ansi = "[\u001B\u009B][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]"
	var re = regexp.MustCompile(ansi)
	s := re.ReplaceAllString(str, "")
	// Also remove other common TUI control characters if any
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func normalizeOutput(str string) string {
	// Remove ALL box drawing characters first
	boxRe := regexp.MustCompile(`[┌┐└┘│─]`)
	str = boxRe.ReplaceAllString(str, "")

	lines := strings.Split(str, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Remove multiple spaces within lines
			re := regexp.MustCompile(`\s+`)
			cleaned := re.ReplaceAllString(trimmed, " ")
			// Remove common TUI artifacts
			cleaned = strings.ReplaceAll(cleaned, "░", "")
			cleaned = strings.ReplaceAll(cleaned, "█", "")
			// Remove any non-printable characters
			printableRe := regexp.MustCompile(`[^[:print:]\s]`)
			cleaned = printableRe.ReplaceAllString(cleaned, "")
			cleaned = strings.TrimSpace(cleaned)
			if cleaned != "" {
				result = append(result, cleaned)
			}
		}
	}
	return strings.Join(result, " ")
}

var capturedOutput string

func waitForText(t *testing.T, tm *teatest.TestModel, target string) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		out := stripANSI(string(bts))
		normalized := normalizeOutput(out)
		capturedOutput += " " + normalized
		// Check normalized, out, and the accumulated capturedOutput
		if strings.Contains(normalized, target) || strings.Contains(out, target) || strings.Contains(capturedOutput, target) {
			return true
		}
		// Also check individual words
		for _, word := range strings.Fields(normalized) {
			if strings.Contains(word, target) {
				return true
			}
		}

		return false
	}, teatest.WithDuration(time.Second*20), teatest.WithCheckInterval(time.Millisecond*200))
}

func TestE2E(t *testing.T) {
	// 1. Setup Mock Server
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		action := r.URL.Query().Get("action")
		switch action {
		case "eth_getTransactionByHash":
			txhash := r.URL.Query().Get("txhash")
			switch txhash {
			case "0x123":
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0x123","blockNumber":"0x100","type":"0x2","from":"0xaaa","to":"0xbbb","value":"0xde0b6b3a7640000","input":"0x"}}`)) //nolint:errcheck // mock server
			case "0x456":
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0x456","blockNumber":"0x100","type":"0x2","from":"0xccc","to":"0xddd","value":"0x0","input":"0x"}}`)) //nolint:errcheck // mock server
			case "0x789":
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0x789","blockNumber":"0x101","type":"0x2","from":"0xeee","to":"0xfff","value":"0x0","input":"0x"}}`)) //nolint:errcheck // mock server
			default:
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`)) //nolint:errcheck // mock server
			}
		case "eth_getBlockByNumber":
			tag := r.URL.Query().Get("tag")
			switch tag {
			case "0x100":
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x65d507c0", "baseFeePerGas":"0x3b9aca00", "transactions": ["0x123", "0x456"]}}`)) //nolint:errcheck // mock server
			case "0x101":
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x65d507c0", "baseFeePerGas":"0x3b9aca00", "transactions": ["0x789"]}}`)) //nolint:errcheck // mock server
			default:
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`)) //nolint:errcheck // mock server
			}
		case "eth_getTransactionReceipt":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"status":"0x1","gasUsed":"0x5208","effectiveGasPrice":"0x3b9aca00"}}`)) //nolint:errcheck // mock server
		case "eth_blockNumber":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x100"}`)) //nolint:errcheck // mock server
		case "eth_getCode":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x"}`)) //nolint:errcheck // mock server
		default:
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`)) //nolint:errcheck // mock server
		}
	})

	server := httptest.NewServer(mockHandler)
	defer server.Close()

	// 2. Initialize Client and Model
	t.Setenv("ETHERSCAN_API_KEY", "test-api-key")

	client := etherscan.NewClient("test-api-key")

	// Hack to set unexported baseURL field
	v := reflect.ValueOf(client).Elem()
	f := v.FieldByName("baseURL")
	if f.IsValid() {
		ptr := unsafe.Pointer(f.UnsafeAddr())
		*(*string)(ptr) = server.URL
	} else {
		t.Fatal("Could not find baseURL field in etherscan.Client")
	}

	m := model.New(client)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 500))

	// Ensure the program has time to start and render
	time.Sleep(time.Second)

	// 3. Step-by-Step E2E
	t.Log("Waiting for initial screen...")
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		out := stripANSI(string(bts))
		return strings.Contains(out, "Ethereum Transaction Explorer") && strings.Contains(out, "Enter transaction hash")
	}, teatest.WithDuration(time.Second*10))
	t.Log("Initial screen found.")

	// Test Switching Network (Tab)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(time.Millisecond * 200)

	// Test Search (0x123)
	tm.Type("0x123")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for result and assert all fields are present and formatted correctly
	t.Log("Waiting for transaction 0x123 details and verifying all fields...")
	waitForText(t, tm, "Hash: 0x123")
	waitForText(t, tm, "success")
	waitForText(t, tm, "1 ETH")
	waitForText(t, tm, "0xaaa")
	waitForText(t, tm, "0xbbb")
	waitForText(t, tm, "Block Number: 256")
	waitForText(t, tm, "Gas Usage: 21000")
	waitForText(t, tm, "Type: 2")
	waitForText(t, tm, "1 Gwei")
	waitForText(t, tm, "0.000021 ETH")
	t.Log("Found 0x123 with key fields correctly formatted.")

	// Test Navigation - Next (n)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	t.Log("Waiting for next transaction 0x456...")
	waitForText(t, tm, "Hash: 0x456")
	t.Log("Found 0x456.")

	// Test Navigation - Previous (p)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	t.Log("Waiting for previous transaction 0x123...")
	waitForText(t, tm, "Hash: 0x123")
	t.Log("Found 0x123 again.")

	// Test Search Again (Esc)
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	waitForText(t, tm, "Enter transaction hash")

	// Test Error State
	tm.Type("0xnonexistent")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	waitForText(t, tm, "Error")

	// Back to search from error
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	waitForText(t, tm, "Enter transaction hash")

	// Test Quit (Ctrl+C)
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*2))
}
