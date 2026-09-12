package term

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
)

func ResolveCellColors(c core.Cell, defaultBg, defaultFg core.Color) (fg, bg core.Color) {
	bg = c.Bg
	if bg.IsTransparent {
		bg = defaultBg
	}
	fg = c.Fg
	if fg.IsTransparent {
		fg = defaultFg
	}
	return fg, bg
}

func StyleANSI(fg, bg core.Color) string {
	return bgToANSI(bg) + fgToANSI(fg)
}

func CellToANSI(c core.Cell, defaultBg, defaultFg core.Color) string {
	fg, bg := ResolveCellColors(c, defaultBg, defaultFg)
	ch := c.Ch
	if ch == 0 {
		ch = ' '
	}
	return bgToANSI(bg) + fgToANSI(fg) + string(ch)
}

func fgToANSI(color core.Color) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", color.R, color.G, color.B)
}

func bgToANSI(color core.Color) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", color.R, color.G, color.B)
}
