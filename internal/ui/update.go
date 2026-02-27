package ui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"
)

// Update handles incoming bubbletea messages to transition the model state.
// Parameters:
//   - msg: The incoming tea.Msg to be processed.
//
// Returns:
//   - The updated tea.Model and a tea.Cmd to execute.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - 10
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			if m.state == inputState {
				if m.chainID == 1 {
					m.chainID = 11155111
				} else {
					m.chainID = 1
				}
				m.client.SetChainID(m.chainID)
			}
		case tea.KeyEnter:
			if m.state == inputState {
				hash := strings.TrimSpace(m.textInput.Value())
				if hash == "" {
					return m, nil
				}
				m.state = loadingState
				m.progress.SetPercent(0)
				// Use m.textInput as a unique ID for the context if needed, but here simple background is fine for now
				// though better to have it cancellable.
				return m, tea.Batch(fetchTransactionCmd(context.Background(), hash, m.client), tickCmd())
			}
			if m.state == resultState || m.state == errorState {
				m.state = inputState
				m.textInput.Reset()
				m.textInput.Focus()
				return m, nil
			}
		}
	case txMsg:
		m.tx = msg.tx
		m.state = resultState
		m.progress.SetPercent(1.0)
		return m, nil
	case errMsg:
		m.err = msg
		m.state = errorState
		return m, nil
	case tickMsg:
		if m.state != loadingState {
			return m, nil
		}
		if m.progress.Percent() >= 0.9 {
			return m, nil
		}
		cmd := m.progress.IncrPercent(0.1)
		return m, tea.Batch(tickCmd(), cmd)
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	if m.state == inputState {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

type tickMsg time.Time

// tickCmd returns a tea.Cmd that sends a tickMsg after 100 milliseconds.
// Returns:
//   - A tea.Cmd to trigger a tick event.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
