// Package model implements the main Bubble Tea application model and message handling.
package model

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/components/block"
	"awesomeProject/internal/tui/components/errorview"
	"awesomeProject/internal/tui/components/footer"
	"awesomeProject/internal/tui/components/header"
	"awesomeProject/internal/tui/components/input"
	"awesomeProject/internal/tui/components/loader"
	"awesomeProject/internal/tui/components/transaction"
	"awesomeProject/internal/tui/context"
	"awesomeProject/internal/tui/theme"
	goctx "context"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	inputState sessionState = iota
	loadingState
	resultState
	errorState
)

// Model is the main application model.
type Model struct {
	state       sessionState
	ctx         *context.ProgramContext
	header      header.Model
	input       input.Model
	transaction transaction.Model
	block       block.Model
	footer      footer.Model
	errorView   errorview.Model
	loader      loader.Model
	client      *etherscan.Client
	tx          *etherscan.Transaction
	blockData   *etherscan.Block
	err         error
}

type txMsg struct{ tx *etherscan.Transaction }
type blockMsg struct{ block *etherscan.Block }
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
		block:       block.New(pCtx, nil),
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

func detectAndFetchCmd(ctx goctx.Context, hash string, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		hash = strings.TrimSpace(hash)
		if isBlockNumber(hash) {
			details, err := client.FetchBlockDetails(ctx, hash)
			if err == nil {
				return blockMsg{block: &etherscan.Block{
					Hash:          "",
					Number:        hash,
					Timestamp:     details.Timestamp,
					BaseFeePerGas: details.BaseFeePerGas,
					Transactions:  details.Transactions,
					FeeRecipient:  details.Miner,
				}}
			}
		}

		// try fetch transaction
		tx, err := client.FetchTransaction(ctx, etherscan.Hash(hash))
		if err != nil {
			return errMsg(err)
		}
		return txMsg{tx: tx}
	}
}

func isBlockNumber(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	if strings.HasPrefix(s, "0x") {
		// A transaction hash is exactly 66 characters long (0x + 64 hex characters).
		// A block number in hex will be much shorter.
		return len(s) <= 20
	}
	return false
}

func fetchTransactionCmd(ctx goctx.Context, hash etherscan.Hash, client *etherscan.Client) tea.Cmd {
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
		tx, err := client.FetchTransaction(ctx, etherscan.Hash(hash))
		if err != nil {
			return errMsg(err)
		}
		return txMsg{tx: tx}
	}
}

func fetchPreviousTransactionCmd(ctx goctx.Context, currentTx *etherscan.Transaction, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		hash, err := client.FetchPreviousTransactionHash(ctx, currentTx)
		if err != nil {
			return errMsg(err)
		}
		tx, err := client.FetchTransaction(ctx, etherscan.Hash(hash))
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
		details, err := client.FetchBlockDetails(ctx, blockNum)
		if err != nil {
			return latestBlockMsg{blockNumber: blockNum}
		}
		var txHash string
		if len(details.Transactions) > 0 {
			txHash = details.Transactions[len(details.Transactions)-1]
		}
		return latestBlockMsg{blockNumber: blockNum, lastTxHash: txHash}
	}
}
