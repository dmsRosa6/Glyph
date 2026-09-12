package framework

import "github.com/dmsRosa6/glyph/core"

type Style struct {
	Bg core.Color
	Fg core.Color
}

func NewTransparentStyle() *Style {
	return &Style{
		Bg: core.Transparent,
		Fg: core.Transparent,
	}
}

func ResolveStyle(style, parent Style) *Style {
	resolved := Style{
		Bg: style.Bg,
		Fg: style.Fg,
	}

	if style.Fg == core.Transparent {
		resolved.Fg = parent.Fg
	}

	if style.Bg == core.Transparent {
		resolved.Bg = parent.Bg
	}

	return &resolved
}

func StyleBg(bg core.Color) Style {
	return Style{Bg: bg, Fg: core.Transparent}
}

func StyleFg(fg core.Color) Style {
	return Style{Bg: core.Transparent, Fg: fg}
}
