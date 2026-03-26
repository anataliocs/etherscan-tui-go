package transaction

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/context"
	"cmp"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	ctx *context.ProgramContext
	tx  *etherscan.Transaction
}

func New(ctx *context.ProgramContext, tx *etherscan.Transaction) Model {
	return Model{
		ctx: ctx,
		tx:  tx,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

func (m Model) View() string {
	if m.tx == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.ctx.Theme.Title.Render("Transaction Details") + "\n")
	b.WriteString(m.ctx.Theme.Purple.Render(strings.Repeat("─", 50)) + "\n\n")

	items := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"Status", m.formatStatus(m.tx.Status), m.getStatusStyle(m.tx.Status)},
		{"Hash", m.tx.Hash, m.ctx.Theme.Value},
		{"Type", m.tx.Type, m.ctx.Theme.Value},
		{"Timestamp", m.tx.Timestamp, m.ctx.Theme.Value},
		{"Block Number", m.tx.BlockNumber, m.ctx.Theme.Value},
		{"From", m.tx.From, m.ctx.Theme.Value},
		{"To", m.tx.To, m.ctx.Theme.Value},
		{"Value", m.tx.Value, m.ctx.Theme.Value},
		{"Gas Limit", m.tx.Gas, m.ctx.Theme.Value},
		{"Gas Usage", m.tx.GasUsed, m.ctx.Theme.Value},
		{"Gas Price", m.tx.GasPrice, m.ctx.Theme.Value},
		{"Transaction Fee", m.tx.TransactionFee, m.ctx.Theme.Value},
		{"Savings", m.tx.Savings, m.ctx.Theme.Savings},
		{"Burnt Fees", m.tx.BurntFees, m.ctx.Theme.Value},
		{"Gas Fees", m.formatGasFees(m.tx), m.ctx.Theme.Value},
		{"Nonce", m.tx.Nonce, m.ctx.Theme.Value},
		{"Tx Index", m.tx.TransactionIndex, m.ctx.Theme.Value},
	}

	for _, item := range items {
		if item.value == "" {
			item.value = "n/a"
		}

		var renderedValue string
		switch {
		case item.label == "Status":
			statusBox := item.style.Render(item.value)
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, m.ctx.Theme.Label.Render(item.label+":"), " ", statusBox) + "\n")
			continue
		case item.label == "Gas Price" && strings.Contains(item.value, "("):
			parts := strings.Split(item.value, " (")
			gwei := parts[0]
			eth := "(" + parts[1]
			renderedValue = item.style.Render(gwei) + " " + m.ctx.Theme.LightGray.Render(eth)
		case item.label == "Block Number" && m.tx.Confirmations != "":
			renderedValue = m.renderBlockNumber(m.tx, item.value, item.style)
		case item.label == "Timestamp" && item.value != "n/a":
			renderedValue = m.renderTimestamp(item.value, item.style)
		case item.label == "Gas Usage" && item.value != "n/a" && m.tx.Gas != "" && m.tx.Gas != "n/a":
			renderedValue = m.renderGasUsage(m.tx, item.value, item.style)
		case item.label == "To" && m.tx.ToAccountType != "":
			renderedValue = item.style.Render(item.value) + " " + m.ctx.Theme.DarkGray.Render(fmt.Sprintf("(%s)", m.tx.ToAccountType))
		default:
			renderedValue = item.style.Render(item.value)
		}

		b.WriteString(m.ctx.Theme.Label.Render(item.label+":") + " " + renderedValue + "\n")
	}

	return b.String()
}

func (m Model) formatGasFees(tx *etherscan.Transaction) string {
	if tx.MaxFeePerGas == "" && tx.MaxPriorityFeePerGas == "" && tx.BaseFeePerGas == "" {
		return "n/a"
	}

	base := cmp.Or(tx.BaseFeePerGas, "n/a")
	max := cmp.Or(tx.MaxFeePerGas, "n/a")
	priority := cmp.Or(tx.MaxPriorityFeePerGas, "n/a")

	return fmt.Sprintf("⛽ Base: %s Gwei | Max: %s Gwei | Max Priority: %s Gwei", base, max, priority)
}

func (m Model) formatStatus(status string) string {
	switch strings.ToLower(status) {
	case "success":
		return "✔ success"
	case "failed":
		return "✘ failed"
	case "pending":
		return "⧖ Pending"
	case "dropped":
		return "↓ dropped"
	case "replaced":
		return "↺ replaced"
	default:
		return status
	}
}

func (m Model) getStatusStyle(status string) lipgloss.Style {
	switch strings.ToLower(status) {
	case "success":
		return m.ctx.Theme.Success
	case "failed":
		return m.ctx.Theme.Failed
	case "pending":
		return m.ctx.Theme.Pending
	case "dropped", "replaced":
		return m.ctx.Theme.Dropped
	default:
		return m.ctx.Theme.Value
	}
}

func (m Model) renderGasUsage(tx *etherscan.Transaction, value string, style lipgloss.Style) string {
	var gasUsed, gasLimit float64
	if _, err := fmt.Sscan(value, &gasUsed); err == nil {
		if _, err := fmt.Sscan(tx.Gas, &gasLimit); err == nil && gasLimit > 0 {
			percentage := (gasUsed / gasLimit) * 100
			return style.Render(value) + " " + m.ctx.Theme.DarkGray.Render(fmt.Sprintf("(%.2f%%)", percentage))
		}
	}
	return style.Render(value)
}

func (m Model) renderBlockNumber(tx *etherscan.Transaction, value string, style lipgloss.Style) string {
	var confText string
	if _, err := fmt.Sscan(tx.Confirmations, new(int)); err == nil {
		confText = fmt.Sprintf(" (%s confirmations)", tx.Confirmations)
	} else {
		confText = fmt.Sprintf(" (%s)", tx.Confirmations)
	}
	return style.Render(value) + " " + m.ctx.Theme.DarkGray.Render(confText)
}

func (m Model) renderTimestamp(value string, style lipgloss.Style) string {
	t, err := time.Parse(time.RFC3339, value)
	if err == nil {
		duration := time.Since(t)
		h := int(duration.Hours())
		m_mins := int(duration.Minutes()) % 60
		s := int(duration.Seconds()) % 60
		var agoStr string
		if h > 0 {
			agoStr = fmt.Sprintf(" (%dh %dm %ds ago)", h, m_mins, s)
		} else if m_mins > 0 {
			agoStr = fmt.Sprintf(" (%dm %ds ago)", m_mins, s)
		} else {
			agoStr = fmt.Sprintf(" (%ds ago)", s)
		}
		return style.Render(value) + " " + m.ctx.Theme.DarkGray.Render(agoStr)
	}
	return style.Render(value)
}
