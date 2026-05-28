package model

import (
	"awesomeProject/internal/etherscan"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_WindowSizeMsg(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	width, height := 100, 50
	msg := tea.WindowSizeMsg{Width: width, Height: height}

	m2, cmd := m.Update(msg)
	updatedModel := m2.(Model)

	if updatedModel.ctx.ScreenWidth != width {
		t.Errorf("expected ScreenWidth %d, got %d", width, updatedModel.ctx.ScreenWidth)
	}
	if updatedModel.ctx.ScreenHeight != height {
		t.Errorf("expected ScreenHeight %d, got %d", height, updatedModel.ctx.ScreenHeight)
	}
	if cmd != nil {
		t.Errorf("expected nil cmd for WindowSizeMsg, got %v", cmd)
	}
}

func TestUpdate_TickMsg(t *testing.T) {
	client := etherscan.NewClient("test-key")
	m := New(client)

	// tickMsg should only increment loader when in loadingState
	m.state = inputState
	m.loader.SetPercent(0.1)
	m2, _ := m.Update(tickMsg{})
	if m2.(Model).loader.Percent() != 0.1 {
		t.Errorf("expected loader percent to stay 0.1 in inputState, got %f", m2.(Model).loader.Percent())
	}

	m.state = loadingState
	m.loader.SetPercent(0.1)
	m3, cmd := m.Update(tickMsg{})
	if m3.(Model).loader.Percent() <= 0.1 {
		t.Errorf("expected loader percent to increment in loadingState, got %f", m3.(Model).loader.Percent())
	}
	if cmd == nil {
		t.Errorf("expected next tick cmd")
	}

	// Should stop incrementing at 0.9
	m.loader.SetPercent(0.9)
	m4, cmd2 := m.Update(tickMsg{})
	if m4.(Model).loader.Percent() != 0.9 {
		t.Errorf("expected loader percent to stay 0.9, got %f", m4.(Model).loader.Percent())
	}
	if cmd2 != nil {
		t.Errorf("expected nil cmd at 0.9 percent")
	}
}

func TestUpdate_ComponentDelegation(t *testing.T) {
	// This is tricky to test deeply without mocks, but we can check if messages
	// that should be handled by components result in state changes in those components.
	client := etherscan.NewClient("test-key")
	m := New(client)

	// Example: input component handles character messages
	m.state = inputState
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if m2.(Model).input.Value() != "a" {
		t.Errorf("expected input value 'a', got %q", m2.(Model).input.Value())
	}
}
