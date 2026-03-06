package ui

import (
	"awesomeProject/internal/etherscan"
	"fmt"
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
	if m.chainID != 1 {
		t.Errorf("expected default chainID 1, got %d", m.chainID)
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
	// We can't easily check if it's tea.Quit without more complex machinery,
	// but we can check if it's not nil.

	// Test Tab toggles chain ID
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedModel := m2.(Model)
	if updatedModel.chainID != 11155111 {
		t.Errorf("expected chainID 11155111 after tab, got %d", updatedModel.chainID)
	}

	m3, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedModel2 := m3.(Model)
	if updatedModel2.chainID != 1 {
		t.Errorf("expected chainID 1 after second tab, got %d", updatedModel2.chainID)
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
	m4, _ := m.Update(latestBlockMsg{blockNumber: "123"})
	updatedModel3 := m4.(Model)
	if updatedModel3.latestBlock != "123" {
		t.Errorf("expected latestBlock 123, got %s", updatedModel3.latestBlock)
	}
	if updatedModel3.isFetchingBlock {
		t.Errorf("expected isFetchingBlock to be false")
	}
}

func TestUpdate_EnterKey(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	// Set some input
	m.textInput.SetValue("0x123")

	// Test Enter starts loading
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel := m2.(Model)
	if updatedModel.state != loadingState {
		t.Errorf("expected state loadingState after Enter, got %v", updatedModel.state)
	}
	if cmd == nil {
		t.Errorf("expected non-nil cmd after Enter")
	}

	// Test Enter from result state returns to input state
	updatedModel.state = resultState
	m3, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel2 := m3.(Model)
	if updatedModel2.state != inputState {
		t.Errorf("expected state inputState after Enter from resultState, got %v", updatedModel2.state)
	}
	if updatedModel2.textInput.Value() != "" {
		t.Errorf("expected textInput to be reset")
	}
}
