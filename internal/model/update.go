package model

import (
	"awesomeProject/internal/tui/components/transaction"
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
)

// Update handles incoming bubbletea messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.ScreenWidth = msg.Width
		m.ctx.ScreenHeight = msg.Height
		m.header.UpdateProgramContext(m.ctx)
		m.input.UpdateProgramContext(m.ctx)
		m.transaction.UpdateProgramContext(m.ctx)
		m.footer.UpdateProgramContext(m.ctx)
		m.errorView.UpdateProgramContext(m.ctx)
		m.loader.UpdateProgramContext(m.ctx)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.state == inputState {
				return m, tea.Quit
			}
			m.state = inputState
			m.input.SetValue("")
			return m, m.input.Focus()
		case tea.KeyTab:
			if m.state == inputState {
				chainID := m.client.ChainID()
				if chainID == 1 {
					chainID = 11155111
				} else {
					chainID = 1
				}
				m.client.SetChainID(chainID)
				m.header.SetChainID(chainID)
				m.header.SetLatestBlock("", "") // Reset while fetching
				return m, fetchLatestBlockCmd(context.Background(), m.client)
			}
		case tea.KeyEnter, tea.KeyBackspace:
			if m.state == inputState && msg.Type == tea.KeyEnter {
				hash := strings.TrimSpace(m.input.Value())
				if hash == "" {
					return m, nil
				}
				m.state = loadingState
				m.loader.SetText(hash)
				return m, tea.Batch(fetchTransactionCmd(context.Background(), hash, m.client), m.loader.SetPercent(0), tickCmd())
			}
			if m.state == resultState || m.state == errorState {
				m.state = inputState
				m.input.SetValue("")
				return m, m.input.Focus()
			}
		case tea.KeyRunes:
			if (strings.Contains(string(msg.Runes), "R") || strings.Contains(string(msg.Runes), "r")) && m.state == resultState {
				hash := m.tx.Hash
				m.state = loadingState
				m.loader.SetText(hash)
				return m, tea.Batch(fetchTransactionCmd(context.Background(), hash, m.client), m.loader.SetPercent(0), tickCmd())
			}
		}
	case txMsg:
		m.tx = msg.tx
		m.state = resultState
		m.transaction = transaction.New(m.ctx, m.tx)
		m.footer.SetHelp("(r) refresh • (backspace/enter/esc) search again • (ctrl+c) quit")
		return m, m.loader.SetPercent(1.0)
	case latestBlockMsg:
		m.header.SetLatestBlock(msg.blockNumber, msg.lastTxHash)
		return m, nil
	case errMsg:
		m.err = msg
		m.errorView.SetError(msg)
		m.state = errorState
		m.footer.SetHelp("press backspace/enter/esc to try again • ctrl+c to quit")
		return m, nil
	case tickMsg:
		if m.state != loadingState {
			return m, nil
		}
		if m.loader.Percent() >= 0.9 {
			return m, nil
		}
		return m, tea.Batch(tickCmd(), m.loader.IncrPercent(0.1))
	}

	m.loader, cmd = m.loader.Update(msg)
	cmds = append(cmds, cmd)

	m.header, cmd = m.header.Update(msg)
	cmds = append(cmds, cmd)

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.transaction, cmd = m.transaction.Update(msg)
	cmds = append(cmds, cmd)

	m.footer, cmd = m.footer.Update(msg)
	cmds = append(cmds, cmd)

	m.errorView, cmd = m.errorView.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
