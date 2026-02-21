package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00ADD8")).
			Width(15)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			MarginTop(1)

	activeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	inactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	pendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	droppedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#800080"))

	lightGrayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)
