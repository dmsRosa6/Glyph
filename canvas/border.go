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
    style *Style
    parentStyle *Style
    layer int
}

type BorderConfig struct {
    Thickness int
    BorderStyle     BorderStyle
    Style Style

    Layer int
}

func DefaultBorderConfig() BorderConfig {
    return BorderConfig{
        Thickness: 1,
        BorderStyle: EmptyBorder,
        Style: Style{Bg: core.Transparent, Fg: core.White},
    }
}

func NewBorder(bounds *geom.Bounds,cfg BorderConfig) (*Border, error) {
    if cfg.Thickness < 1 {
        panic("border thickness must be >= 1")
    }

    s := ResolveStyle(cfg.Style,*NewTransparentStyle())

    b := &Border{
        bounds:     bounds,
        thickness:  cfg.Thickness,
        borderStyle: cfg.BorderStyle,
        style: s,
    }

    if err := b.SetLayer(cfg.Layer); err != nil {
        return nil ,err
    }

    return b, nil
}

func (r *Border) Draw(buf *core.Buffer, vec geom.Vector) {
    var borderStyleBg core.Color
    var borderStyleFg core.Color
    
    s := ResolveStyle(*r.style,*r.parentStyle)

    borderStyleBg = s.Bg
    borderStyleFg = s.Fg

    for layer := 0; layer < r.thickness; layer++ {
        x0 := r.bounds.Pos.X + layer
        y0 := r.bounds.Pos.Y + layer
        x1 := r.bounds.Pos.X + r.bounds.W - 1 - layer
        y1 := r.bounds.Pos.Y + r.bounds.H - 1 - layer

        // corners
        buf.Set(vec.X + x0, vec.Y + y0, r.borderStyle.TopLeft, borderStyleBg, borderStyleFg)
        buf.Set(vec.X + x1, vec.Y + y0, r.borderStyle.TopRight, borderStyleBg, borderStyleFg)
        buf.Set(vec.X + x0, vec.Y + y1, r.borderStyle.BottomLeft, borderStyleBg, borderStyleFg)
        buf.Set(vec.X + x1, vec.Y + y1, r.borderStyle.BottomRight, borderStyleBg, borderStyleFg)

        // top & bottom edges
        for x := x0 + 1; x < x1; x++ {
            buf.Set(vec.X + x, vec.Y, r.borderStyle.Horizontal, borderStyleBg, borderStyleFg)
            buf.Set(vec.X + x, vec.Y + y1, r.borderStyle.Horizontal, borderStyleBg, borderStyleFg)
        }

        // left & right edges
        for y := y0 + 1; y < y1; y++ {
            buf.Set(vec.X + x0, vec.Y + y, r.borderStyle.Vertical, borderStyleBg, borderStyleFg)
            buf.Set(vec.X + x1, vec.Y + y, r.borderStyle.Vertical, borderStyleBg, borderStyleFg)
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

func (r *Border) SetLayer(l int) error{
    if l < 0{
		return errors.New("Layers must be greater or equal to 0")
	}

    r.layer = l
    return  nil
}

func (r *Border) GetLayer() int{
    return r.layer
}

func (r *Border) SetParentStyle(s *Style){
    r.parentStyle = s
}
