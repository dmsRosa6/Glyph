package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Rect struct {
	Bounds geom.Bounds
	Ch     rune
	Fg, Bg core.Color

	layer int
}

type RectConfig struct {
    Ch     rune
    Fg, Bg core.Color

	Layer int
}

func NewRect(bounds geom.Bounds, cfg RectConfig) (*Rect, error) {
	r :=  &Rect{
        Bounds: bounds,
        Ch:     cfg.Ch,
        Fg:     cfg.Fg,
        Bg:     cfg.Bg,
    }

	if err := r.SetLayer(cfg.Layer); err != nil {
		return nil, err
	}


	return r, nil
}

func (r *Rect) Draw(buf *core.Buffer) {
    for y := r.Bounds.Pos.Y; y < r.Bounds.Pos.Y+r.Bounds.H; y++ {
            for x := r.Bounds.Pos.X; x < r.Bounds.Pos.X+r.Bounds.W; x++ {

                buf.Set(x, y, r.Ch, r.Bg, r.Fg)
            }
    }
}

func (r *Rect) IsInBounds(parent geom.Bounds) bool{
	if r.Bounds.Pos.X < parent.Pos.X {
		return false
	}

	if r.Bounds.Pos.Y < parent.Pos.Y {
		return false
	}

	if r.Bounds.Pos.Y + r.Bounds.H > parent.Pos.Y + parent.H {
		return false
	}

	if r.Bounds.Pos.X + r.Bounds.W > parent.Pos.X + parent.W {
		return false
	}

	return true
}


func (r *Rect) MoveTo(p geom.Point) {
    r.Bounds.Pos = p
}

func (r *Rect) Translate(v geom.Vector) {
    r.Bounds.Pos = r.Bounds.Pos.Add(v)
}

func (r *Rect) SetLayer(l int) error{
	if l < 0{
		return errors.New("Layers must be greater or equal to 0")
	} 

	r.layer = l

	return nil
}

func (r *Rect) GetLayer() int{
    return r.layer
}