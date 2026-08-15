package term

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
)

func CellToANSI(c core.Cell, defaultFg, defaultBg core.Color) string {
	bg := c.Bg
	if bg.IsTransparent {
		bg = defaultBg
	}
	fg := c.Fg
	if fg.IsTransparent {
		fg = defaultFg
	}
	seq := bgToANSI(bg) + fgToANSI(fg)
	if c.Ch == 0 {
		seq += " "
	}
	return seq + string(c.Ch)
}

func fgToANSI(color core.Color) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", color.R, color.G, color.B)
}

func bgToANSI(color core.Color) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", color.R, color.G, color.B)
}
