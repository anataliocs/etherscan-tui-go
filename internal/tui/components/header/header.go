package header

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	ctx             *context.ProgramContext
	chainID         int
	latestBlock     string
	latestTxHash    string
	isFetchingBlock bool
	spinner         spinner.Model
}

func New(ctx *context.ProgramContext, chainID int) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return Model{
		ctx:             ctx,
		chainID:         chainID,
		isFetchingBlock: true,
		spinner:         s,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.isFetchingBlock {
		m.spinner, cmd = m.spinner.Update(msg)
	}
	return m, cmd
}

func (m Model) Tick() tea.Cmd {
	return m.spinner.Tick
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
	m.isFetchingBlock = true
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
		latestBlockDisplay += m.spinner.View()
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
