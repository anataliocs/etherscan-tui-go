package ui

import (
	"awesomeProject/internal/etherscan"
	"context"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	inputState sessionState = iota
	loadingState
	resultState
	errorState
)

type Model struct {
	state           sessionState
	textInput       textinput.Model
	progress        progress.Model
	tx              *etherscan.Transaction
	err             error
	client          *etherscan.Client
	chainID         int
	latestBlock     string
	isFetchingBlock bool
}

type txMsg struct{ tx *etherscan.Transaction }
type latestBlockMsg struct{ blockNumber string }
type errMsg error

// New creates a new Model with the given Etherscan client.
// Parameters:
//   - client: A pointer to an etherscan.Client used for API calls.
//
// Returns:
//   - A Model initialized with the provided client and default settings.
func New(client *etherscan.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "0x..."
	ti.Focus()
	ti.CharLimit = 66
	ti.Width = 70

	return Model{
		state:           inputState,
		textInput:       ti,
		progress:        progress.New(progress.WithDefaultGradient()),
		client:          client,
		chainID:         client.ChainID(),
		isFetchingBlock: true,
	}
}

// Init initializes the Model, starting the blinking cursor for the text input and fetching the latest block number.
// Returns:
//   - A tea.Cmd to be executed on initialization.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, fetchLatestBlockCmd(context.Background(), m.client))
}

// fetchTransactionCmd returns a tea.Cmd that fetches transaction details for a given hash.
// Parameters:
//   - ctx: The context for the API request.
//   - hash: The transaction hash to fetch.
//   - client: The Etherscan client to use for the request.
//
// Returns:
//   - A tea.Cmd that will produce either a txMsg on success or an errMsg on failure.
func fetchTransactionCmd(ctx context.Context, hash string, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		tx, err := client.FetchTransaction(ctx, hash)
		if err != nil {
			return errMsg(err)
		}
		return txMsg{tx: tx}
	}
}

// fetchLatestBlockCmd returns a tea.Cmd that fetches the latest block number.
// Parameters:
//   - ctx: The context for the API request.
//   - client: The Etherscan client to use for the request.
//
// Returns:
//   - A tea.Cmd that will produce a latestBlockMsg on success or an errMsg on failure.
func fetchLatestBlockCmd(ctx context.Context, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		blockNum, err := client.FetchLatestBlockNumber(ctx)
		if err != nil {
			return errMsg(err)
		}
		return latestBlockMsg{blockNumber: blockNum}
	}
}
