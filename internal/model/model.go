package model

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/components/errorview"
	"awesomeProject/internal/tui/components/footer"
	"awesomeProject/internal/tui/components/header"
	"awesomeProject/internal/tui/components/input"
	"awesomeProject/internal/tui/components/loader"
	"awesomeProject/internal/tui/components/transaction"
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	goctx "context"

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
	state       sessionState
	ctx         *context.ProgramContext
	header      header.Model
	input       input.Model
	transaction transaction.Model
	footer      footer.Model
	errorView   errorview.Model
	loader      loader.Model
	client      *etherscan.Client
	tx          *etherscan.Transaction
	err         error
}

type txMsg struct{ tx *etherscan.Transaction }
type latestBlockMsg struct {
	blockNumber string
	lastTxHash  string
}
type errMsg error

// New creates a new Model with the given Etherscan client.
func New(client *etherscan.Client) Model {
	pCtx := &context.ProgramContext{
		Theme: theme.DefaultTheme(),
	}

	return Model{
		state:       inputState,
		ctx:         pCtx,
		header:      header.New(pCtx, client.ChainID()),
		input:       input.New(pCtx),
		transaction: transaction.New(pCtx, nil),
		footer:      footer.New(pCtx, "(tab) switch network • (l) latest hash • (enter) search • (ctrl+c) quit"),
		errorView:   errorview.New(pCtx, nil),
		loader:      loader.New(pCtx),
		client:      client,
	}
}

// Init initializes the Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.input.Focus(),
		fetchLatestBlockCmd(goctx.Background(), m.client),
		m.header.Tick(),
	)
}

func fetchTransactionCmd(ctx goctx.Context, hash string, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		tx, err := client.FetchTransaction(ctx, hash)
		if err != nil {
			return errMsg(err)
		}
		return txMsg{tx: tx}
	}
}

func fetchNextTransactionCmd(ctx goctx.Context, currentTx *etherscan.Transaction, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		hash, err := client.FetchNextTransactionHash(ctx, currentTx)
		if err != nil {
			return errMsg(err)
		}
		tx, err := client.FetchTransaction(ctx, hash)
		if err != nil {
			return errMsg(err)
		}
		return txMsg{tx: tx}
	}
}

func fetchLatestBlockCmd(ctx goctx.Context, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		blockNum, err := client.FetchLatestBlockNumber(ctx)
		if err != nil {
			return errMsg(err)
		}
		_, _, txHashes, err := client.FetchBlockDetails(ctx, blockNum)
		if err != nil {
			return latestBlockMsg{blockNumber: blockNum}
		}
		var txHash string
		if len(txHashes) > 0 {
			txHash = txHashes[len(txHashes)-1]
		}
		return latestBlockMsg{blockNumber: blockNum, lastTxHash: txHash}
	}
}
