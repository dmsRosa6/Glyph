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
    Fg, Bg core.Color

	Anchor Anchor
	Layer int
}

func NewRect(bounds *geom.Bounds, cfg RectConfig) (*Rect, error) {
	
	//0 does not occupy any space so its fucks the layout

	var actualRune rune

	if cfg.Ch == 0 {
		actualRune = 32
	}else{
		actualRune = cfg.Ch
	}
    var bg core.Color
    var fg core.Color

    if cfg.Bg == (core.Color{}){
        bg = core.Transparent
    }else{
        bg = core.Transparent
    }

    if cfg.Fg == (core.Color{}){
        fg = core.Transparent
    }else{
        fg = cfg.Fg
    }

    s := &Style{
        Bg: bg,
        Fg: fg,
    }


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
    for y := r.bounds.Pos.Y; y < r.bounds.Pos.Y+r.bounds.H; y++ {
            for x := r.bounds.Pos.X; x < r.bounds.Pos.X+r.bounds.W; x++ {
                buf.Set(vec.X + x, vec.Y + y, r.ch, r.style.Bg, r.style.Fg)
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
    r.layout.computedPos.X = resolveAxis(r.layout.anchor.H, parent.Pos.X, parent.W, r.bounds.W, r.bounds.Pos.X)
    r.layout.computedPos.Y = resolveAxis(r.layout.anchor.H, parent.Pos.Y, parent.H, r.bounds.H, r.bounds.Pos.Y)
}

func (r *Rect) SetParentStyle(s *Style){
    r.parentStyle = s
}
