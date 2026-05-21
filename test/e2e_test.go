package test

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/model"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

type mockClient struct{}

func (m *mockClient) ChainID() int {
	return 1
}

func (m *mockClient) SetChainID(id int) {
}

func (m *mockClient) FetchTransaction(ctx context.Context, hash string) (*etherscan.Transaction, error) {
	return nil, nil
}

func (m *mockClient) FetchLatestBlockNumber(ctx context.Context) (string, error) {
	return "", nil
}

func (m *mockClient) FetchBlockDetails(ctx context.Context, blockNumber string) (string, string, []string, error) {
	return "", "", nil, nil
}

func (m *mockClient) FetchNextTransactionHash(ctx context.Context, currentTx *etherscan.Transaction) (string, error) {
	return "", nil
}

func (m *mockClient) FetchPreviousTransactionHash(ctx context.Context, currentTx *etherscan.Transaction) (string, error) {
	return "", nil
}

func (m *mockClient) IsContract(ctx context.Context, address string) (bool, error) {
	return false, nil
}

func (m *mockClient) FetchTransactionReceipt(ctx context.Context, hash string) (string, string, string, error) {
	return "", "", "", nil
}

func waitForText(t *testing.T, tm *teatest.TestModel, target string) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		out := stripANSI(string(bts))
		outFields := strings.Join(strings.Fields(out), " ")
		return strings.Contains(out, target) || strings.Contains(outFields, target)
	}, teatest.WithDuration(time.Second*10), teatest.WithCheckInterval(time.Millisecond*200))
}

func stripANSI(str string) string {
	const ansi = "[\u001B\u009B][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]"
	var re = regexp.MustCompile(ansi)
	s := re.ReplaceAllString(str, "")
	// Also remove other common TUI control characters if any
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func TestE2E(t *testing.T) {
	// 1. Setup Mock Server
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		action := r.URL.Query().Get("action")
		switch action {
		case "eth_getTransactionByHash":
			txhash := r.URL.Query().Get("txhash")
			if txhash == "0x123" {
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0x123","blockNumber":"0x100","type":"0x2","from":"0xaaa","to":"0xbbb","value":"0xde0b6b3a7640000","input":"0x"}}`))
			} else if txhash == "0x456" {
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0x456","blockNumber":"0x100","type":"0x2","from":"0xccc","to":"0xddd","value":"0x0","input":"0x"}}`))
			} else if txhash == "0x789" {
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"hash":"0x789","blockNumber":"0x101","type":"0x2","from":"0xeee","to":"0xfff","value":"0x0","input":"0x"}}`))
			} else {
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
			}
		case "eth_getBlockByNumber":
			tag := r.URL.Query().Get("tag")
			if tag == "0x100" {
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x65d507c0", "baseFeePerGas":"0x3b9aca00", "transactions": ["0x123", "0x456"]}}`))
			} else if tag == "0x101" {
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x65d507c0", "baseFeePerGas":"0x3b9aca00", "transactions": ["0x789"]}}`))
			} else {
				w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
			}
		case "eth_getTransactionReceipt":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"status":"0x1","gasUsed":"0x5208","effectiveGasPrice":"0x3b9aca00"}}`))
		case "eth_blockNumber":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x100"}`))
		case "eth_getCode":
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x"}`))
		default:
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
		}
	})

	server := httptest.NewServer(mockHandler)
	defer server.Close()

	// 2. Initialize Client and Model
	os.Setenv("ETHERSCAN_API_KEY", "test-api-key")
	defer os.Unsetenv("ETHERSCAN_API_KEY")

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
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

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

	// Wait for result
	t.Log("Waiting for transaction 0x123 details...")
	waitForText(t, tm, "Hash: 0x123")
	t.Log("Found 0x123.")

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
