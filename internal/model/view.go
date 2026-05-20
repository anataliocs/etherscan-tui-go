package model

// View renders the current state of the Model.
func (m Model) View() string {
	var s string
	footerWidth := m.ctx.ScreenWidth

	switch m.state {
	case inputState:
		s = m.header.View() + "\n\n" + m.input.View()
	case loadingState:
		return "\n" + m.loader.View() + "\n"
	case resultState:
		s = m.transaction.View()
		if m.ctx.ScreenWidth >= 80 {
			footerWidth = int(float64(m.ctx.ScreenWidth) * 0.6)
		}
	case errorState:
		s = m.errorView.View()
	}

	m.ctx.FooterWidth = footerWidth
	return "\n" + s + "\n" + m.footer.View() + "\n"
}
