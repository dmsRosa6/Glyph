package core

// Cell is a small value type -- a rune plus two colors -- not a
// pointer, and there is deliberately no NewCell constructor anymore.
// Buffer stores these directly in a flat slice (see buffer.go) so
// Clear/Set can write values in place with zero heap allocation.
//
// Writing Cell{Ch: ch, Fg: fg, Bg: bg} with named fields at each call
// site is exactly as clear as a constructor call, and it makes a real
// bug that used to live here structurally impossible to reintroduce:
// the old NewCell(ch rune, bg, fg Color) took bg before fg, but
// Buffer.Set -- whose own parameters are also named (bg, fg), in that
// order -- called it as NewCell(ch, fg, bg), silently swapping every
// cell's foreground and background color. Since every widget's Draw
// goes through Buffer.Set, that swap wasn't a local glitch: it flipped
// Fg/Bg for every Rect fill, Border glyph, Button, Text draw, Spinner
// frame, and PaletteNode cell in the entire framework. Named-field
// literals with no intermediate positional constructor mean there's no
// argument order left to get backwards.
type Cell struct {
	Ch rune
	Fg Color
	Bg Color
}
