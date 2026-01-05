package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Border struct{
    borderStyle BorderStyle
    bounds geom.Bounds
	thickness int
	fg, bg core.Color
}

type BorderConfig struct {
    Bounds    geom.Bounds
    Thickness int
    Style     BorderStyle
    Fg, Bg    core.Color
}

func DefaultBorderConfig() BorderConfig {
    return BorderConfig{
        Thickness: 1,
        Style: BorderStyle{
            TopLeft:     ' ',
            TopRight:    ' ',
            BottomLeft:  ' ',
            BottomRight: ' ',
            Horizontal:  ' ',
            Vertical:    ' ',
        },
        Fg: core.White,
        Bg: core.Transparent,
    }
}

func NewBorder(cfg BorderConfig) *Border {
    if cfg.Thickness < 1 {
        panic("border thickness must be >= 1")
    }

    return &Border{
        bounds:     cfg.Bounds,
        thickness:  cfg.Thickness,
        borderStyle: cfg.Style,
        fg:         cfg.Fg,
        bg:         cfg.Bg,
    }
}

func (r *Border) Draw(buf *core.Buffer) {
    for layer := 0; layer < r.thickness; layer++ {
        x0 := r.bounds.Pos.X + layer
        y0 := r.bounds.Pos.Y + layer
        x1 := r.bounds.Pos.X + r.bounds.W - 1 - layer
        y1 := r.bounds.Pos.Y + r.bounds.H - 1 - layer

        // corners
        buf.Set(x0, y0, r.borderStyle.TopLeft, r.bg, r.fg)
        buf.Set(x1, y0, r.borderStyle.TopRight, r.bg, r.fg)
        buf.Set(x0, y1, r.borderStyle.BottomLeft, r.bg, r.fg)
        buf.Set(x1, y1, r.borderStyle.BottomRight, r.bg, r.fg)

        // top & bottom edges
        for x := x0 + 1; x < x1; x++ {
            buf.Set(x, y0, r.borderStyle.Horizontal, r.bg, r.fg)
            buf.Set(x, y1, r.borderStyle.Horizontal, r.bg, r.fg)
        }

        // left & right edges
        for y := y0 + 1; y < y1; y++ {
            buf.Set(x0, y, r.borderStyle.Vertical, r.bg, r.fg)
            buf.Set(x1, y, r.borderStyle.Vertical, r.bg, r.fg)
        }
    }
}

func (r *Border) IsInBounds(parent geom.Bounds) bool{
	if r.bounds.Pos.X < parent.Pos.X {
		return false
	}

	if r.bounds.Pos.Y < parent.Pos.Y {
		return false
	}

	if r.bounds.Pos.Y + r.bounds.H > parent.Pos.Y + parent.H {
		return false
	}

	if r.bounds.Pos.X + r.bounds.W > parent.Pos.X + parent.W {
		return false
	}

	return true
}

func (r *Border) MoveTo(p geom.Point) {
    r.bounds.Pos = p
}

func (r *Border) Translate(v geom.Vector) {
    r.bounds.Pos = r.bounds.Pos.Add(v)
}


