package term

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
)

// ResolveCellColors resolves a Cell's Fg/Bg against buffer defaults --
// core.Transparent is the "inherit the buffer's own default" sentinel,
// the same convention framework.ResolveStyle uses for widget styles.
// Factored out so CellToANSI (one cell) and render.Renderer.Flush's
// run-batching (one resolution per contiguous run of identically-styled
// cells, not per cell) share exactly one implementation.
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

// StyleANSI returns just the fg/bg-setting escape sequence, no glyph --
// the piece CellToANSI builds internally, exposed separately so a
// caller doing its own run-batching (see render.Renderer.Flush) can
// emit ONE style escape for a whole run of identically-styled cells
// instead of recomputing and re-emitting it once per cell.
func StyleANSI(fg, bg core.Color) string {
	return bgToANSI(bg) + fgToANSI(fg)
}

// CellToANSI renders one cell's full escape sequence + glyph. ch == 0
// (a Cell's zero value, never actually produced by Buffer.Set/Clear
// today, but not assumed away either) substitutes a space rather than
// printing a literal NUL byte -- this used to append " " to the escape
// sequence and THEN still append string(c.Ch) unconditionally, which
// for Ch == 0 meant emitting a stray space followed by an actual NUL
// character rather than substituting one for the other. Never actually
// triggered in practice (every real Cell so far has always had a real
// Ch), but wrong regardless, and now fixed alongside the code around it
// rather than left as a latent trap for whatever writes a Cell next.
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
