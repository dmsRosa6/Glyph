package framework

import (
	"testing"

	"github.com/dmsRosa6/glyph/core"
)

func TestResolveStyle(t *testing.T) {
	cases := []struct {
		name   string
		style  Style
		parent Style
		want   Style
	}{
		{
			name:   "concrete colors stay as-is, parent ignored",
			style:  Style{Bg: core.Red, Fg: core.White},
			parent: Style{Bg: core.Blue, Fg: core.Black},
			want:   Style{Bg: core.Red, Fg: core.White},
		},
		{
			name:   "transparent Bg inherits the parent's Bg",
			style:  Style{Bg: core.Transparent, Fg: core.White},
			parent: Style{Bg: core.Blue, Fg: core.Black},
			want:   Style{Bg: core.Blue, Fg: core.White},
		},
		{
			name:   "transparent Fg inherits the parent's Fg",
			style:  Style{Bg: core.Red, Fg: core.Transparent},
			parent: Style{Bg: core.Blue, Fg: core.Black},
			want:   Style{Bg: core.Red, Fg: core.Black},
		},
		{
			name:   "both transparent inherits both, independently",
			style:  Style{Bg: core.Transparent, Fg: core.Transparent},
			parent: Style{Bg: core.Blue, Fg: core.Black},
			want:   Style{Bg: core.Blue, Fg: core.Black},
		},
		{
			name:   "parent itself transparent propagates transparent through",
			style:  Style{Bg: core.Transparent, Fg: core.Transparent},
			parent: Style{Bg: core.Transparent, Fg: core.Transparent},
			want:   Style{Bg: core.Transparent, Fg: core.Transparent},
		},
		{
			// The documented gotcha: a bare Style{} is NOT the same as
			// "inherit everything" -- its zero-value Bg/Fg equal
			// core.Black (R:0,G:0,B:0,IsTransparent:false), not
			// core.Transparent, so this resolves to opaque black on
			// both channels regardless of what the parent is.
			name:   "bare Style{} resolves to Black/Black, not inherited",
			style:  Style{},
			parent: Style{Bg: core.Red, Fg: core.White},
			want:   Style{Bg: core.Black, Fg: core.Black},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveStyle(tc.style, tc.parent)
			if got.Bg != tc.want.Bg {
				t.Errorf("Bg = %+v, want %+v", got.Bg, tc.want.Bg)
			}
			if got.Fg != tc.want.Fg {
				t.Errorf("Fg = %+v, want %+v", got.Fg, tc.want.Fg)
			}
		})
	}
}

func TestStyleBg(t *testing.T) {
	got := StyleBg(core.Red)
	want := Style{Bg: core.Red, Fg: core.Transparent}
	if got != want {
		t.Errorf("StyleBg(core.Red) = %+v, want %+v", got, want)
	}
}

func TestStyleFg(t *testing.T) {
	got := StyleFg(core.Red)
	want := Style{Bg: core.Transparent, Fg: core.Red}
	if got != want {
		t.Errorf("StyleFg(core.Red) = %+v, want %+v", got, want)
	}
}

// TestNewTransparentStyle confirms the helper actually produces the
// "inherit everything" sentinel, not the Style{} zero value that looks
// similar but isn't -- see the bare-Style{} case in TestResolveStyle
// for why that distinction matters.
func TestNewTransparentStyle(t *testing.T) {
	got := NewTransparentStyle()
	if got.Bg != core.Transparent || got.Fg != core.Transparent {
		t.Errorf("NewTransparentStyle() = %+v, want both fields core.Transparent", *got)
	}
}
