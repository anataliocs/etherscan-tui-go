package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Title     lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Error     lipgloss.Style
	Active    lipgloss.Style
	Inactive  lipgloss.Style
	Help      lipgloss.Style
	Pending   lipgloss.Style
	Success   lipgloss.Style
	Failed    lipgloss.Style
	Dropped   lipgloss.Style
	LightGray lipgloss.Style
	DarkGray  lipgloss.Style
	Savings   lipgloss.Style
}

func DefaultTheme() *Theme {
	return &Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"}).
			MarginBottom(1),

		Label: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#00ADD8", Dark: "#00ADD8"}).
			Width(18),

		Value: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#FAFAFA"}),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}).
			MarginTop(1),

		Active: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"}),

		Inactive: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#626262", Dark: "#626262"}),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#626262", Dark: "#626262"}).
			MarginTop(1),

		Pending: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#D4AF37", Dark: "#FFFF00"}).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#D4AF37", Dark: "#FFFF00"}).
			Padding(0, 1),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#008000", Dark: "#00FF00"}).
			Bold(true).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#008000", Dark: "#00FF00"}).
			Padding(0, 1),

		Failed: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}).
			Padding(0, 1),

		Dropped: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#800080", Dark: "#800080"}).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#800080", Dark: "#800080"}).
			Padding(0, 1),

		LightGray: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#888888"}),

		DarkGray: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#555555", Dark: "#555555"}),

		Savings: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#008000", Dark: "#00FF00"}).
			Italic(true),
	}
}
