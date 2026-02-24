package ui

import (
	"fmt"
	"strings"
	"time"

	"awesomeProject/internal/etherscan"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	progress  progress.Model
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
		progress:  progress.New(progress.WithDefaultGradient()),
		client:    client,
		chainID:   client.ChainID(),
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

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
				return m, tea.Batch(fetchTransactionCmd(hash, m.client), tickCmd())
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
		{"Timestamp", tx.Timestamp, valueStyle},
		{"Block Number", tx.BlockNumber, valueStyle},
		{"From", tx.From, valueStyle},
		{"To", tx.To, valueStyle},
		{"Value", tx.Value, valueStyle},
		{"Gas", tx.Gas, valueStyle},
		{"Gas Price", tx.GasPrice, valueStyle},
		{"Transaction Fee", tx.TransactionFee, valueStyle},
		{"Nonce", tx.Nonce, valueStyle},
		{"Tx Index", tx.TransactionIndex, valueStyle},
	}

	for _, item := range items {
		if item.value == "" {
			item.value = "n/a"
		}

		var renderedValue string
		if item.label == "Gas Price" && strings.Contains(item.value, "(") {
			parts := strings.Split(item.value, " (")
			gwei := parts[0]
			eth := "(" + parts[1]
			renderedValue = item.style.Render(gwei) + " " + lightGrayStyle.Render(eth)
		} else {
			renderedValue = item.style.Render(item.value)
		}

		b.WriteString(labelStyle.Render(item.label+":") + " " + renderedValue + "\n")
	}

	return b.String()
}

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

func fetchTransactionCmd(hash string, client *etherscan.Client) tea.Cmd {
	return func() tea.Msg {
		tx, err := client.FetchTransaction(hash)
		if err != nil {
			return errMsg(err)
		}
		return txMsg{tx: tx}
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
