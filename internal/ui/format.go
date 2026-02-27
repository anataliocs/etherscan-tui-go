package ui

import (
	"awesomeProject/internal/etherscan"
	"cmp"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// formatGasFees returns a formatted string containing base, max, and priority gas fees.
// Parameters:
//   - tx: The transaction object containing the fee information.
//
// Returns:
//   - A string formatted as "Base: ... | Max: ... | Max Priority: ...".
func formatGasFees(tx *etherscan.Transaction) string {
	if tx.MaxFeePerGas == "" && tx.MaxPriorityFeePerGas == "" && tx.BaseFeePerGas == "" {
		return "n/a"
	}

	base := cmp.Or(tx.BaseFeePerGas, "n/a")
	max := cmp.Or(tx.MaxFeePerGas, "n/a")
	priority := cmp.Or(tx.MaxPriorityFeePerGas, "n/a")

	return fmt.Sprintf("Base: %s Gwei | Max: %s Gwei | Max Priority: %s Gwei", base, max, priority)
}

// formatStatus returns a human-readable string for the transaction status with symbols.
// Parameters:
//   - status: The transaction status as a string.
//
// Returns:
//   - A formatted status string (e.g., "✔ success", "✘ failed").
func formatStatus(status string) string {
	switch strings.ToLower(status) {
	case "success":
		return "✔ success"
	case "failed":
		return "✘ failed"
	case "pending":
		return "Pending"
	case "dropped":
		return "dropped"
	case "replaced":
		return "replaced"
	default:
		return status
	}
}

// getStatusStyle returns the lipgloss style for a given transaction status.
// Parameters:
//   - status: The transaction status string.
//
// Returns:
//   - A lipgloss.Style corresponding to the status.
func getStatusStyle(status string) lipgloss.Style {
	switch strings.ToLower(status) {
	case "success":
		return successStyle
	case "failed":
		return failedStyle
	case "pending":
		return pendingStyle
	case "dropped", "replaced":
		return droppedStyle
	default:
		return valueStyle
	}
}
