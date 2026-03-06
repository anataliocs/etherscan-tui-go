package context

import (
	"awesomeProject/internal/tui/theme"
)

type ProgramContext struct {
	ScreenWidth  int
	ScreenHeight int
	Theme        *theme.Theme
}
