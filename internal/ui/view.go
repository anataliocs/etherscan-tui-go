package ui

import "fmt"

// View renders the current state of the Model as a string.
// Returns:
//   - A string representation of the UI.
func (m Model) View() string {
	var s string
	switch m.state {
	case inputState:
		var networkToggle string
		if m.chainID == 1 {
			networkToggle = activeStyle.Render("Mainnet") + " | " + inactiveStyle.Render("Sepolia")
		} else {
			networkToggle = inactiveStyle.Render("Mainnet") + " | " + activeStyle.Render("Sepolia")
		}

		s = fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s",
			titleStyle.Render("Ethereum Transaction Explorer"),
			"Network: "+networkToggle,
			"Enter transaction hash:",
			m.textInput.View(),
		) + helpStyle.Render("\n\n(tab) switch network • (enter) search • (esc) quit")
	case loadingState:
		s = fmt.Sprintf(
			"\n  Searching for %s...\n\n  %s",
			m.textInput.Value(),
			m.progress.View(),
		)
	case resultState:
		s = renderTransaction(m.tx)
		s += helpStyle.Render("\n\npress enter to search again • esc to quit")
	case errorState:
		s = fmt.Sprintf(
			"%s\n\n%s",
			titleStyle.Render("Error"),
			errorStyle.Render(m.err.Error()),
		) + helpStyle.Render("\n\npress enter to try again • esc to quit")
	}
	return "\n" + s + "\n"
}
