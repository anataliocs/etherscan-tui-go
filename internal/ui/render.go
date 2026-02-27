package ui

import (
	"awesomeProject/internal/etherscan"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderTransaction creates a formatted string representation of a transaction's details.
// Parameters:
//   - tx: The transaction object containing the data to render.
//
// Returns:
//   - A formatted string ready for display in the UI.
func renderTransaction(tx *etherscan.Transaction) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Transaction Details") + "\n\n")

	items := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"Hash", tx.Hash, valueStyle},
		{"Status", formatStatus(tx.Status), getStatusStyle(tx.Status)},
		{"Type", tx.Type, valueStyle},
		{"Timestamp", tx.Timestamp, valueStyle},
		{"Block Number", tx.BlockNumber, valueStyle},
		{"From", tx.From, valueStyle},
		{"To", tx.To, valueStyle},
		{"Value", tx.Value, valueStyle},
		{"Gas Limit", tx.Gas, valueStyle},
		{"Gas Usage", tx.GasUsed, valueStyle},
		{"Gas Price", tx.GasPrice, valueStyle},
		{"Transaction Fee", tx.TransactionFee, valueStyle},
		{"Gas Fees", formatGasFees(tx), valueStyle},
		{"Nonce", tx.Nonce, valueStyle},
		{"Tx Index", tx.TransactionIndex, valueStyle},
	}

	for _, item := range items {
		if item.value == "" {
			item.value = "n/a"
		}

		var renderedValue string
		switch {
		case item.label == "Gas Price" && strings.Contains(item.value, "("):
			parts := strings.Split(item.value, " (")
			gwei := parts[0]
			eth := "(" + parts[1]
			renderedValue = item.style.Render(gwei) + " " + lightGrayStyle.Render(eth)
		case item.label == "Block Number" && tx.Confirmations != "":
			renderedValue = renderBlockNumber(tx, renderedValue, item)
		case item.label == "Timestamp" && item.value != "n/a":
			renderedValue = renderTimestamp(item, renderedValue)
		case item.label == "Gas Usage" && item.value != "n/a" && tx.Gas != "" && tx.Gas != "n/a":
			renderedValue = renderGasUsage(tx, item, renderedValue)
		case item.label == "To" && tx.ToAccountType != "":
			renderedValue = item.style.Render(item.value) + " " + darkGrayStyle.Render(fmt.Sprintf("(%s)", tx.ToAccountType))
		default:
			renderedValue = item.style.Render(item.value)
		}

		b.WriteString(labelStyle.Render(item.label+":") + " " + renderedValue + "\n")
	}

	return b.String()
}

// renderGasUsage calculates and formats gas usage percentage relative to the limit.
// Parameters:
//   - tx: The transaction object containing gas limit and usage.
//   - item: The UI list item currently being processed.
//   - renderedValue: The pre-existing value string to modify.
//
// Returns:
//   - A string containing gas usage with its percentage.
func renderGasUsage(tx *etherscan.Transaction, item struct {
	label string
	value string
	style lipgloss.Style
}, renderedValue string) string {
	var gasUsed, gasLimit float64
	if _, err := fmt.Sscan(item.value, &gasUsed); err == nil {
		if _, err := fmt.Sscan(tx.Gas, &gasLimit); err == nil && gasLimit > 0 {
			percentage := (gasUsed / gasLimit) * 100
			renderedValue = item.style.Render(item.value) + " " + darkGrayStyle.Render(fmt.Sprintf("(%.2f%%)", percentage))
		} else {
			renderedValue = item.style.Render(item.value)
		}
	} else {
		renderedValue = item.style.Render(item.value)
	}
	return renderedValue
}

// renderBlockNumber formats the block number with confirmation count if available.
// Parameters:
//   - tx: The transaction object containing confirmation details.
//   - renderedValue: The block number as a string.
//   - item: The current UI item being rendered.
//
// Returns:
//   - A formatted block number string with confirmations.
func renderBlockNumber(tx *etherscan.Transaction, renderedValue string, item struct {
	label string
	value string
	style lipgloss.Style
}) string {
	var confText string
	if _, err := fmt.Sscan(tx.Confirmations, new(int)); err == nil {
		confText = fmt.Sprintf(" (%s confirmations)", tx.Confirmations)
	} else {
		confText = fmt.Sprintf(" (%s)", tx.Confirmations)
	}
	renderedValue = item.style.Render(item.value) + " " + darkGrayStyle.Render(confText)
	return renderedValue
}

// renderTimestamp formats the ISO 8601 timestamp with a relative time indicator.
// Parameters:
//   - item: The UI item containing the timestamp value.
//   - renderedValue: The pre-rendered timestamp string.
//
// Returns:
//   - A string with the original timestamp and the time elapsed since then.
func renderTimestamp(item struct {
	label string
	value string
	style lipgloss.Style
}, renderedValue string) string {
	t, err := time.Parse(time.RFC3339, item.value)
	if err == nil {
		duration := time.Since(t)
		h := int(duration.Hours())
		m := int(duration.Minutes()) % 60
		s := int(duration.Seconds()) % 60
		var agoStr string
		if h > 0 {
			agoStr = fmt.Sprintf(" (%dh %dm %ds ago)", h, m, s)
		} else if m > 0 {
			agoStr = fmt.Sprintf(" (%dm %ds ago)", m, s)
		} else {
			agoStr = fmt.Sprintf(" (%ds ago)", s)
		}
		renderedValue = item.style.Render(item.value) + " " + darkGrayStyle.Render(agoStr)
	} else {
		renderedValue = item.style.Render(item.value)
	}
	return renderedValue
}
