package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Rect struct {
	bounds *geom.Bounds
	ch     rune
	fg core.Color
	bg core.Color

	layer int
}

type RectConfig struct {
    Ch     rune
    Fg, Bg core.Color

	Layer int
}

func NewRect(bounds *geom.Bounds, cfg RectConfig) (*Rect, error) {
	
	//0 does not occupy any space so its fucks the layot

	var actualRune rune

	if cfg.Ch == 0 {
		actualRune = 32
	}else{
		actualRune = cfg.Ch
	}

	r :=  &Rect{
        bounds: bounds,
        ch:     actualRune,
        fg:     cfg.Fg,
        bg:     cfg.Bg,
    }

	if err := r.SetLayer(cfg.Layer); err != nil {
		return nil, err
	}


	return r, nil
}

func (r *Rect) Draw(buf *core.Buffer, vec geom.Vector) {
    for y := r.bounds.Pos.Y; y < r.bounds.Pos.Y+r.bounds.H; y++ {
            for x := r.bounds.Pos.X; x < r.bounds.Pos.X+r.bounds.W; x++ {
                buf.Set(vec.X + x, vec.Y + y, r.ch, r.bg, r.fg)
            }
    }
}

func (r *Rect) IsInBounds(parent geom.Bounds) bool{
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

func (r *Rect) MoveTo(p *geom.Point) {
    r.bounds.Pos = p
}

func (r *Rect) Translate(v geom.Vector) {
    r.bounds.Pos.AddVector(v)
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