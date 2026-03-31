package header

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx             *context.ProgramContext
	chainID         int
	latestBlock     string
	latestTxHash    string
	isFetchingBlock bool
}

func New(ctx *context.ProgramContext, chainID int) Model {
	return Model{
		ctx:             ctx,
		chainID:         chainID,
		isFetchingBlock: true,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

func (m *Model) SetLatestBlock(block string, txHash string) {
	m.latestBlock = block
	m.latestTxHash = txHash
	m.isFetchingBlock = false
}

func (m *Model) SetChainID(id int) {
	m.chainID = id
}

func (m Model) View() string {
	var networkToggle string
	if m.chainID == 1 {
		networkToggle = m.ctx.Theme.Active.Render("Mainnet") + " | " + m.ctx.Theme.Inactive.Render("Sepolia")
	} else {
		networkToggle = m.ctx.Theme.Inactive.Render("Mainnet") + " | " + m.ctx.Theme.Active.Render("Sepolia")
	}

	latestBlockDisplay := "Total Transactions: "
	if m.isFetchingBlock {
		latestBlockDisplay += "Loading..."
	} else if m.latestBlock != "" {
		latestBlockDisplay += etherscan.FormatLatestBlock(m.latestBlock)
		if m.latestTxHash != "" {
			latestBlockDisplay += "\nLatest Transaction Hash: " + m.ctx.Theme.Inactive.Render(m.latestTxHash)
		}
	} else {
		latestBlockDisplay += "n/a"
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.ctx.Theme.Title.Render("Ethereum Transaction Explorer"),
		latestBlockDisplay,
		"Network: "+networkToggle,
	)
}
