package model

// View renders the current state of the Model.
func (m Model) View() string {
	var s string
	switch m.state {
	case inputState:
		s = m.header.View() + "\n\n" + m.input.View()
	case loadingState:
		return "\n" + m.loader.View() + "\n"
	case resultState:
		s = m.transaction.View()
	case errorState:
		s = m.errorView.View()
	}
	return "\n" + s + m.footer.View() + "\n"
}
