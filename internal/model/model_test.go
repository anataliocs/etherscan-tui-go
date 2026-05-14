package model

import (
	"awesomeProject/internal/etherscan"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	if m.state != inputState {
		t.Errorf("expected state inputState, got %v", m.state)
	}
	if m.client != client {
		t.Errorf("expected client to be set")
	}
	// We can't directly check component private fields, so we check through the client or component getters if we had them.
	// But let's check what's public.
	if m.client.ChainID() != 1 {
		t.Errorf("expected default chainID 1, got %d", m.client.ChainID())
	}
}

func TestUpdate_KeyEvents(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	// Test Ctrl+C quits
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Ctrl+C")
	}

	// Test Tab toggles chain ID
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedModel := m2.(Model)
	if updatedModel.client.ChainID() != 11155111 {
		t.Errorf("expected chainID 11155111 after tab, got %d", updatedModel.client.ChainID())
	}

	m3, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedModel2 := m3.(Model)
	if updatedModel2.client.ChainID() != 1 {
		t.Errorf("expected chainID 1 after second tab, got %d", updatedModel2.client.ChainID())
	}
}

func TestUpdate_Transitions(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	// Test txMsg transition
	tx := &etherscan.Transaction{Hash: "0xabc"}
	m2, _ := m.Update(txMsg{tx: tx})
	updatedModel := m2.(Model)
	if updatedModel.state != resultState {
		t.Errorf("expected state resultState after txMsg, got %v", updatedModel.state)
	}
	if updatedModel.tx != tx {
		t.Errorf("expected tx to be set")
	}

	// Test errMsg transition
	m3, _ := m.Update(errMsg(fmt.Errorf("test error")))
	updatedModel2 := m3.(Model)
	if updatedModel2.state != errorState {
		t.Errorf("expected state errorState after errMsg, got %v", updatedModel2.state)
	}

	// Test latestBlockMsg transition
	// m4, _ := m.Update(latestBlockMsg{blockNumber: "123"})
	// updatedModel3 := m4.(Model)
	// We can't easily check header's latest block without a getter or making it public.
}

func TestUpdate_EnterKey(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	// Set some input
	m.input.SetValue("0x123")

	// Test Enter starts loading
	m.state = inputState
	m.input.SetValue("0x123")
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel := m2.(Model)
	if updatedModel.state != loadingState {
		t.Errorf("expected state loadingState after Enter, got %v", updatedModel.state)
	}
	if cmd == nil {
		t.Errorf("expected non-nil cmd after Enter")
	}

	// Test Backspace from result state returns to input state
	updatedModel.state = resultState
	m3, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	updatedModel2 := m3.(Model)
	if updatedModel2.state != inputState {
		t.Errorf("expected state inputState after Backspace from resultState, got %v", updatedModel2.state)
	}
	if updatedModel2.input.Value() != "" {
		t.Errorf("expected textInput to be reset")
	}

	// Test Esc from result state returns to input state
	updatedModel.state = resultState
	m4, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updatedModel3 := m4.(Model)
	if updatedModel3.state != inputState {
		t.Errorf("expected state inputState after Esc from resultState, got %v", updatedModel3.state)
	}

	// Test Enter from result state returns to input state
	updatedModel.state = resultState
	m5, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel4 := m5.(Model)
	if updatedModel4.state != inputState {
		t.Errorf("expected state inputState after Enter from resultState, got %v", updatedModel4.state)
	}
}

func TestFooterHelpReset(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	initialHelp := "(tab) switch network • (l) latest hash • (enter) search • (ctrl+c) quit"
	if m.footer.Help() != initialHelp {
		t.Errorf("expected initial help %q, got %q", initialHelp, m.footer.Help())
	}

	// Transition to resultState
	tx := &etherscan.Transaction{Hash: "0xabc"}
	m2, _ := m.Update(txMsg{tx: tx})
	updatedModel := m2.(Model)
	resultHelp := "(r) refresh • (p) prev tx • (n) next tx • (backspace/enter/esc) search again • (ctrl+c) quit"
	if updatedModel.footer.Help() != resultHelp {
		t.Errorf("expected result help %q, got %q", resultHelp, updatedModel.footer.Help())
	}

	// Transition back to inputState via Esc
	m3, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updatedModel2 := m3.(Model)
	if updatedModel2.state != inputState {
		t.Errorf("expected state inputState, got %v", updatedModel2.state)
	}
	if updatedModel2.footer.Help() != initialHelp {
		t.Errorf("expected help to be reset to %q, got %q", initialHelp, updatedModel2.footer.Help())
	}

	// Transition to errorState
	m4, _ := m.Update(errMsg(fmt.Errorf("test error")))
	updatedModel3 := m4.(Model)
	errorHelp := "press backspace/enter/esc to try again • ctrl+c to quit"
	if updatedModel3.footer.Help() != errorHelp {
		t.Errorf("expected error help %q, got %q", errorHelp, updatedModel3.footer.Help())
	}

	// Transition back to inputState via Enter
	m5, _ := updatedModel3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel4 := m5.(Model)
	if updatedModel4.state != inputState {
		t.Errorf("expected state inputState, got %v", updatedModel4.state)
	}
	if updatedModel4.footer.Help() != initialHelp {
		t.Errorf("expected help to be reset to %q, got %q", initialHelp, updatedModel4.footer.Help())
	}
}

func TestUpdate_LatestHash(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	// Set latest hash in header
	m.header.SetLatestBlock("123", "0xlatest")

	// Test 'l' key
	m2, cmd := m.Update(tea.KeyMsg{Runes: []rune("l"), Type: tea.KeyRunes})
	updatedModel := m2.(Model)

	if updatedModel.state != loadingState {
		t.Errorf("expected state loadingState after 'l', got %v", updatedModel.state)
	}
	if updatedModel.input.Value() != "0xlatest" {
		t.Errorf("expected input value '0xlatest', got %q", updatedModel.input.Value())
	}
	if cmd == nil {
		t.Errorf("expected non-nil cmd after 'l'")
	}

	// Reset to input state
	m.state = inputState
	m.input.SetValue("")

	// Test 'L' key
	m3, cmd2 := m.Update(tea.KeyMsg{Runes: []rune("L"), Type: tea.KeyRunes})
	updatedModel2 := m3.(Model)

	if updatedModel2.state != loadingState {
		t.Errorf("expected state loadingState after 'L', got %v", updatedModel2.state)
	}
	if updatedModel2.input.Value() != "0xlatest" {
		t.Errorf("expected input value '0xlatest', got %q", updatedModel2.input.Value())
	}
	if cmd2 == nil {
		t.Errorf("expected non-nil cmd after 'L'")
	}
}

func TestUpdate_NextTransaction(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	// Set initial transaction
	tx := &etherscan.Transaction{Hash: "0x123", BlockNumber: "10"}
	m.tx = tx
	m.state = resultState

	// Test 'n' key
	m2, cmd := m.Update(tea.KeyMsg{Runes: []rune("n"), Type: tea.KeyRunes})
	updatedModel := m2.(Model)

	if updatedModel.state != loadingState {
		t.Errorf("expected state loadingState after 'n', got %v", updatedModel.state)
	}
	if updatedModel.loader.View() == "" || !strings.Contains(updatedModel.loader.View(), "next transaction") {
		t.Errorf("expected loader to mention next transaction")
	}
	if cmd == nil {
		t.Errorf("expected non-nil cmd after 'n'")
	}

	// Test 'N' key
	m.state = resultState
	m3, cmd2 := m.Update(tea.KeyMsg{Runes: []rune("N"), Type: tea.KeyRunes})
	updatedModel2 := m3.(Model)

	if updatedModel2.state != loadingState {
		t.Errorf("expected state loadingState after 'N', got %v", updatedModel2.state)
	}
	if cmd2 == nil {
		t.Errorf("expected non-nil cmd after 'N'")
	}
}

func TestUpdate_PreviousTransaction(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	// Set initial transaction
	tx := &etherscan.Transaction{Hash: "0x123", BlockNumber: "10"}
	m.tx = tx
	m.state = resultState

	// Test 'p' key
	m2, cmd := m.Update(tea.KeyMsg{Runes: []rune("p"), Type: tea.KeyRunes})
	updatedModel := m2.(Model)

	if updatedModel.state != loadingState {
		t.Errorf("expected state loadingState after 'p', got %v", updatedModel.state)
	}
	if updatedModel.loader.View() == "" || !strings.Contains(updatedModel.loader.View(), "previous transaction") {
		t.Errorf("expected loader to mention previous transaction")
	}
	if cmd == nil {
		t.Errorf("expected non-nil cmd after 'p'")
	}

	// Test 'P' key
	m.state = resultState
	m3, cmd2 := m.Update(tea.KeyMsg{Runes: []rune("P"), Type: tea.KeyRunes})
	updatedModel2 := m3.(Model)

	if updatedModel2.state != loadingState {
		t.Errorf("expected state loadingState after 'P', got %v", updatedModel2.state)
	}
	if cmd2 == nil {
		t.Errorf("expected non-nil cmd after 'P'")
	}
}

func TestLoadingViewNoFooter(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)
	m.state = loadingState
	m.loader.SetText("0x123")

	view := m.View()
	if !strings.Contains(view, "Searching for 0x123...") {
		t.Errorf("expected view to contain loader text, got %q", view)
	}

	initialHelp := "(tab) switch network • (l) latest hash • (enter) search • (ctrl+c) quit"
	if strings.Contains(view, initialHelp) {
		t.Errorf("expected loading view NOT to contain footer help text")
	}
}
