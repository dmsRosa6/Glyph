package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Rect struct {
	bounds *geom.Bounds
	ch     rune
	style *Style
	parentStyle *Style
	layout *Layout
	layer int
}

type RectConfig struct {
    Ch     rune
	Style Style

	Anchor Anchor
	Layer int
}

func NewRect(bounds *geom.Bounds, cfg RectConfig) (*Rect, error) {
	var actualRune rune

	if cfg.Ch == 0 {
		actualRune = 32
	}else{
		actualRune = cfg.Ch
	}

    s := ResolveStyle(cfg.Style, *NewTransparentStyle())

	r :=  &Rect{
        bounds: bounds,
        ch:     actualRune,
		style: s,
		layout : &Layout{
					computedPos: bounds.Pos,
					anchor: &cfg.Anchor,
				},

    }

	if err := r.SetLayer(cfg.Layer); err != nil {
		return nil, err
	}


	return r, nil
}

func (r *Rect) Draw(buf *core.Buffer, vec geom.Vector) {

	rectX := r.layout.computedPos.X
	rectY := r.layout.computedPos.Y

	s := ResolveStyle(*r.style, *r.parentStyle)

	fg := s.Fg
	bg := s.Bg

    for y := rectY; y < rectY+r.bounds.H; y++ {
            for x := rectX; x < rectX+r.bounds.W; x++ {
                buf.Set(vec.X + x, vec.Y + y, r.ch, bg, fg)
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

func (r *Rect) Layout(parent geom.Bounds) {
    r.layout.computedPos.X = resolveAxis(r.layout.anchor.H, parent.W, r.bounds.W, r.bounds.Pos.X)
    r.layout.computedPos.Y = resolveAxis(r.layout.anchor.V, parent.H, r.bounds.H, r.bounds.Pos.Y)
}

func (r *Rect) SetParentStyle(s *Style){
    r.parentStyle = s
}
