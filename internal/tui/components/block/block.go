// Package block provides a component for displaying detailed information about an Ethereum block.
package block

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the block details component state.
type Model struct {
	ctx   *context.ProgramContext
	block *etherscan.Block
}

// New creates a new block component with the given context and block data.
func New(ctx *context.ProgramContext, block *etherscan.Block) Model {
	return Model{
		ctx:   ctx,
		block: block,
	}
}

// Update updates the block component state.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// UpdateProgramContext updates the block component's reference to the global program context.
func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

// View renders the block details as a string.
func (m Model) View() string {
	if m.block == nil {
		return ""
	}

	width := m.calculateWidth()
	return m.renderDetails(width)
}

func (m Model) calculateWidth() int {
	if m.ctx.ScreenWidth == 0 {
		return 80 // fallback
	}
	return m.ctx.ScreenWidth
}

func (m Model) renderDetails(width int) string {
	var b strings.Builder
	b.WriteString(m.ctx.Theme.Title.Render("Block Details") + "\n")

	sepWidth := max(20, width-2)
	b.WriteString(m.ctx.Theme.Purple.Render(strings.Repeat("─", sepWidth)) + "\n\n")

	labelStyle := m.ctx.Theme.Label.Copy().Width(min(18, width-10))

	var items []struct {
		label string
		value string
	}

	items = append(items, struct {
		label string
		value string
	}{"Hash", m.block.Hash})
	items = append(items, struct {
		label string
		value string
	}{"Number", m.block.Number})
	items = append(items, struct {
		label string
		value string
	}{"Timestamp", m.block.Timestamp})
	items = append(items, struct {
		label string
		value string
	}{"Base Fee", m.block.BaseFeePerGas})
	items = append(items, struct {
		label string
		value string
	}{"Fee Recipient", m.block.FeeRecipient})

	items = append(items, struct {
		label string
		value string
	}{"Proposed On", fmt.Sprintf("Slot %s, Epoch %s", m.block.Slot, m.block.Epoch)})

	items = append(items, struct {
		label string
		value string
	}{"Block Reward", m.block.BlockReward})

	if m.block.Status != "" {
		items = append(items, struct {
			label string
			value string
		}{"Status", m.block.Status})
	}

	items = append(items, struct {
		label string
		value string
	}{"Transactions", fmt.Sprintf("%d", len(m.block.Transactions))})

	for _, item := range items {
		b.WriteString(labelStyle.Render(item.label+":") + " " + m.ctx.Theme.Value.Render(item.value) + "\n")
	}

	return b.String()
}
