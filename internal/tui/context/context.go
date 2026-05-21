// Package context defines the global program context shared across TUI components.
package context

import (
	"awesomeProject/internal/tui/theme"
)

// ProgramContext holds global state such as screen dimensions and the current theme.
type ProgramContext struct {
	ScreenWidth  int
	ScreenHeight int
	FooterWidth  int
	Theme        *theme.Theme
}
