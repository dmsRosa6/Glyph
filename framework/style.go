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

// ResolveStyle merges a shape's own style with its parent's.
//
// core.Transparent is the ONLY "inherit from parent" sentinel, for both
// Fg and Bg, applied symmetrically. Previously Bg additionally treated
// the zero-value core.Color{} as "unset" -- but core.Color{} has the same
// field values as core.Black, so an explicitly-set Black background was
// indistinguishable from "not set" and got silently replaced by the
// parent's background. That special case is gone: if you want to inherit,
// set Transparent; anything else (including Black) is taken literally.
// Fg no longer has the side effect of forcing Bg transparent too --
// each channel resolves independently.
func ResolveStyle(style, parent Style) *Style {
	resolved := Style{
		Fg: style.Fg,
		Bg: style.Bg,
	}

	if style.Fg == core.Transparent {
		resolved.Fg = parent.Fg
	}

	if style.Bg == core.Transparent {
		resolved.Bg = parent.Bg
	}

	return &resolved
}