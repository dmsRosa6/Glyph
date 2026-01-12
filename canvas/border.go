package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Border struct{
    borderStyle BorderStyle
    bounds *geom.Bounds
	thickness int
	fg, bg core.Color

    layer int
}

type BorderConfig struct {
    Thickness int
    Style     BorderStyle
    Fg, Bg    core.Color

    Layer int
}

func DefaultBorderConfig() BorderConfig {
    return BorderConfig{
        Thickness: 1,
        Style: EmptyBorder,
        Fg: core.White,
        Bg: core.Transparent,
    }
}

func NewBorder(bounds *geom.Bounds, cfg BorderConfig) (*Border, error) {
    if cfg.Thickness < 1 {
        panic("border thickness must be >= 1")
    }

    b := &Border{
        bounds:     bounds,
        thickness:  cfg.Thickness,
        borderStyle: cfg.Style,
        fg:         cfg.Fg,
        bg:         cfg.Bg,
    }

    if err := b.SetLayer(cfg.Layer); err != nil {
        return nil ,err
    }

    return b, nil
}

func (r *Border) Draw(buf *core.Buffer, vec geom.Vector) {
    for layer := 0; layer < r.thickness; layer++ {
        x0 := r.bounds.Pos.X + layer
        y0 := r.bounds.Pos.Y + layer
        x1 := r.bounds.Pos.X + r.bounds.W - 1 - layer
        y1 := r.bounds.Pos.Y + r.bounds.H - 1 - layer

        // corners
        buf.Set(vec.X + x0, vec.Y + y0, r.borderStyle.TopLeft, r.bg, r.fg)
        buf.Set(vec.X + x1, vec.Y + y0, r.borderStyle.TopRight, r.bg, r.fg)
        buf.Set(vec.X + x0, vec.Y + y1, r.borderStyle.BottomLeft, r.bg, r.fg)
        buf.Set(vec.X + x1, vec.Y + y1, r.borderStyle.BottomRight, r.bg, r.fg)

        // top & bottom edges
        for x := x0 + 1; x < x1; x++ {
            buf.Set(vec.X + x, vec.Y + y0, r.borderStyle.Horizontal, r.bg, r.fg)
            buf.Set(vec.X + x, vec.Y + y1, r.borderStyle.Horizontal, r.bg, r.fg)
        }

        // left & right edges
        for y := y0 + 1; y < y1; y++ {
            buf.Set(vec.X + x0, vec.Y + y, r.borderStyle.Vertical, r.bg, r.fg)
            buf.Set(vec.X + x1, vec.Y + y, r.borderStyle.Vertical, r.bg, r.fg)
        }
    }
}

func (r *Border) IsInBounds(parent geom.Bounds) bool{
	if r.bounds.Pos.X < 0 {
		return false
	}

	if r.bounds.Pos.Y < 0 {
		return false
	}

	if r.bounds.Pos.Y + r.bounds.H > parent.H {
		return false
	}

	if r.bounds.Pos.X + r.bounds.W > parent.W {
		return false
	}

	return true
}

func (r *Border) MoveTo(p *geom.Point) {
    r.bounds.Pos = p
}

func (r *Border) Translate(v geom.Vector) {
    r.bounds.Pos.AddVector(v)
}

func (r *Border) SetLayer(l int) error{
    if l < 0{
		return errors.New("Layers must be greater or equal to 0")
	} 


    r.layer = l
    return  nil
}

func (r *Border) getLayer() int{
    return r.layer
}
