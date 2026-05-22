// Package header provides the header component for the Ethereum Transaction Explorer TUI.
// It displays the title, current network, and latest block information.
package header

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the header component state.
type Model struct {
	ctx             *context.ProgramContext
	chainID         int
	latestBlock     string
	latestTxHash    string
	isFetchingBlock bool
	spinner         spinner.Model
}

// New creates a new header component with the given context and chain ID.
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

// Update updates the header component state based on the received message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.isFetchingBlock {
		m.spinner, cmd = m.spinner.Update(msg)
	}
	return m, cmd
}

// Tick returns a command that performs a spinner tick.
func (m Model) Tick() tea.Cmd {
	return m.spinner.Tick
}

// UpdateProgramContext updates the header's reference to the global program context.
func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

// SetLatestBlock updates the header with the latest block and transaction hash.
func (m *Model) SetLatestBlock(block string, txHash string) {
	m.latestBlock = block
	m.latestTxHash = txHash
	m.isFetchingBlock = false
}

// SetChainID updates the chain ID and resets the fetching state.
func (m *Model) SetChainID(id int) {
	m.chainID = id
	m.isFetchingBlock = true
}

// LatestTxHash returns the latest transaction hash stored in the header.
func (m Model) LatestTxHash() string {
	return m.latestTxHash
}

// View renders the header component as a string.
func (m Model) View() string {
	var networkToggle string
	if m.chainID == 1 {
		networkToggle = m.ctx.Theme.Active.Render("Mainnet") + " | " + m.ctx.Theme.Inactive.Render("Sepolia")
	} else {
		networkToggle = m.ctx.Theme.Inactive.Render("Mainnet") + " | " + m.ctx.Theme.Active.Render("Sepolia")
	}

	latestBlockDisplay := "Total Transactions: "
	switch {
	case m.isFetchingBlock:
		latestBlockDisplay += m.spinner.View()
	case m.latestBlock != "":
		latestBlockDisplay += etherscan.FormatLatestBlock(m.latestBlock)
		if m.latestTxHash != "" {
			latestBlockDisplay += "\nLatest Transaction Hash: " + m.ctx.Theme.Inactive.Render(m.latestTxHash)
		}
	default:
		latestBlockDisplay += "n/a"
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.ctx.Theme.Title.Render("Ethereum Transaction Explorer"),
		latestBlockDisplay,
		"Network: "+networkToggle,
	)
}
