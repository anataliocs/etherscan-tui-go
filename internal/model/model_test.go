package model

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
