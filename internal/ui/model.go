package ui

import (
	"fmt"
	"strings"

	"awesomeProject/internal/etherscan"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	inputState sessionState = iota
	loadingState
	resultState
	errorState
)

type Model struct {
	state     sessionState
	textInput textinput.Model
	tx        *etherscan.Transaction
	err       error
	client    *etherscan.Client
	chainID   int
}

type txMsg struct{ tx *etherscan.Transaction }
type errMsg error

func New(client *etherscan.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "0x..."
	ti.Focus()
	ti.CharLimit = 66
	ti.Width = 70

	return Model{
		state:     inputState,
		textInput: ti,
		client:    client,
		chainID:   client.ChainID(),
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
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
				return m, fetchTransactionCmd(hash, m.client)
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
		return m, nil
	case errMsg:
		m.err = msg
		m.state = errorState
		return m, nil
	}

	if m.state == inputState {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

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
		) + helpStyle.Render("\n\npress tab to switch network • enter to search • esc to quit")
	case loadingState:
		s = fmt.Sprintf("\n  Searching for %s...\n", m.textInput.Value())
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

func renderTransaction(tx *etherscan.Transaction) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Transaction Details") + "\n\n")

	items := []struct {
		label string
		value string
	}{
		{"Hash", tx.Hash},
		{"Block Number", tx.BlockNumber},
		{"From", tx.From},
		{"To", tx.To},
		{"Value", tx.Value},
		{"Gas", tx.Gas},
		{"Gas Price", tx.GasPrice},
		{"Nonce", tx.Nonce},
		{"Tx Index", tx.TransactionIndex},
	}

	for _, item := range items {
		if item.value == "" {
			item.value = "n/a"
		}
		b.WriteString(labelStyle.Render(item.label+":") + " " + valueStyle.Render(item.value) + "\n")
	}

	return b.String()
}

func fetchTransactionCmd(hash string, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		tx, err := client.FetchTransaction(hash)
		if err != nil {
			return errMsg(err)
		}
		return txMsg{tx: tx}
	}
}
