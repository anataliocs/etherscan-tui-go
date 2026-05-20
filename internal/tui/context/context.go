package context

import (
	"awesomeProject/internal/tui/theme"
)

type ProgramContext struct {
	ScreenWidth  int
	ScreenHeight int
	FooterWidth  int
	Theme        *theme.Theme
}
