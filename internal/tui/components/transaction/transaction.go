package transaction

import (
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/tui/context"
	"cmp"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	ctx      *context.ProgramContext
	tx       *etherscan.Transaction
	viewport viewport.Model
}

func New(ctx *context.ProgramContext, tx *etherscan.Transaction) Model {
	m := Model{
		ctx: ctx,
		tx:  tx,
	}

	if tx != nil && tx.Input != "" && tx.Input != "0x" {
		m.viewport = viewport.New(0, 0)
		m.viewport.SetContent(m.renderInputHex(tx.Input))
	}

	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

func (m Model) View() string {
	if m.tx == nil {
		return ""
	}

	detailsWidth, inputWidth := m.calculateWidths()

	if inputWidth == 0 {
		// Vertical layout for small screens
		details := m.renderDetails(detailsWidth)
		input := m.renderInputData(detailsWidth)
		if input == "" {
			return details
		}
		return details + "\n\n" + input
	}

	details := m.renderDetails(detailsWidth)
	input := m.renderInputData(inputWidth)

	if input == "" {
		return details
	}

	detailsStyle := lipgloss.NewStyle().Width(detailsWidth).PaddingRight(2)
	inputStyle := lipgloss.NewStyle().Width(inputWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		detailsStyle.Render(details),
		inputStyle.Render(input),
	)
}

func (m Model) calculateWidths() (int, int) {
	if m.ctx.ScreenWidth > 0 && m.ctx.ScreenWidth < 80 {
		return m.ctx.ScreenWidth, 0 // Vertical layout signal: return full width and 0 for input
	}

	// Use exactly 50% of width for each view
	halfWidth := m.ctx.ScreenWidth / 2
	if halfWidth == 0 {
		halfWidth = 50 // fallback
	}

	return halfWidth, halfWidth - 2
}

func (m Model) renderDetails(width int) string {
	var b strings.Builder
	b.WriteString(m.ctx.Theme.Title.Render("Transaction Details") + "\n")

	sepWidth := max(20, width-2)
	b.WriteString(m.ctx.Theme.Purple.Render(strings.Repeat("─", sepWidth)) + "\n\n")

	labelStyle := m.ctx.Theme.Label.Copy().Width(min(18, width-10))

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
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(item.label+":"), " ", statusBox) + "\n")
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

		b.WriteString(labelStyle.Render(item.label+":") + " " + renderedValue + "\n")
	}

	return b.String()
}

func (m Model) renderInputData(width int) string {
	if m.tx.Input == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.ctx.Theme.Title.Render("Input Data (Raw Hex)") + "\n")

	sepWidth := max(20, width)
	b.WriteString(m.ctx.Theme.Purple.Render(strings.Repeat("─", sepWidth)) + "\n\n")

	if m.tx.Input == "0x" {
		b.WriteString(m.ctx.Theme.Value.Render("0x") + "\n")
		return b.String()
	}

	// For non-empty input, use the viewport
	// Calculate height based on screen height or some reasonable limit
	height := 10 // default
	if m.ctx.ScreenHeight > 20 {
		height = m.ctx.ScreenHeight - 15 // Leave space for header/footer and details
	}
	if height < 5 {
		height = 5
	}

	m.viewport.Width = width
	m.viewport.Height = height

	// Indicators for scrolling
	var indicators string
	if m.viewport.AtTop() && m.viewport.AtBottom() {
		// All content fits, no indicators needed
	} else {
		if !m.viewport.AtTop() {
			indicators += " ↑"
		}
		if !m.viewport.AtBottom() {
			indicators += " ↓"
		}
		b.WriteString(m.ctx.Theme.DarkGray.Render("Scrollable:"+indicators) + "\n")
	}

	b.WriteString(m.viewport.View())

	return b.String()
}

func (m Model) renderInputHex(hexInput string) string {
	var b strings.Builder
	// Remove 0x prefix for formatting
	input := strings.TrimPrefix(hexInput, "0x")

	// Format as a grid: 16 bytes (32 chars) per row
	// Example: 0000: 60 80 60 40 52 34 80 15 61 00 10 57 60 00 80 fd
	for i := 0; i < len(input); i += 32 {
		end := min(i+32, len(input))
		row := input[i:end]

		// Offset
		b.WriteString(m.ctx.Theme.DarkGray.Render(fmt.Sprintf("%04x: ", i/2)))

		// Hex bytes
		for j := 0; j < len(row); j += 2 {
			byteEnd := min(j+2, len(row))
			b.WriteString(m.ctx.Theme.Value.Render(row[j:byteEnd]) + " ")
		}

		// Pad short rows
		if len(row) < 32 {
			padding := (32 - len(row)) / 2
			b.WriteString(strings.Repeat("   ", padding))
		}

		b.WriteString("\n")
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
