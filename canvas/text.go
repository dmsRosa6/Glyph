package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Text struct {
	bounds geom.Bounds
	value string
	style *Style
	parentStyle *Style
	layout *Layout
	layer int
}

type TextConfig struct {
    Value string
	Bg, Fg core.Color
	Anchor Anchor
	Layer int
}


func NewText(pos *geom.Point, cfg TextConfig) (*Text, error) {
	var bg core.Color
    var fg core.Color

	bg = core.Transparent

	fg = cfg.Fg

    s := &Style{
        Bg: bg,
        Fg: fg,
    }
	
	t := &Text{
		bounds:  *geom.NewBounds(pos.X, pos.Y, len(cfg.Value), 1),
		value: cfg.Value,
		style: s,
		layout: &Layout{anchor: &cfg.Anchor, computedPos: pos},
	}

	if err := t.SetLayer(cfg.Layer); err != nil{
		return nil, err
	}

	return t, nil
}

func (t *Text) Draw(buf *core.Buffer, vec geom.Vector) {
	x := t.layout.computedPos.X
	y := t.layout.computedPos.Y

	for i := 0; i < len(t.value); i++{
	
		r := rune(t.value[i])
		
		buf.Set(vec.X + x+i, vec.Y + y, r, t.style.Bg, t.style.Fg)
	}
}

func (t *Text) IsInBounds(parent geom.Bounds) bool{

	if t.bounds.Pos.X < 0 {
		return false
	}

	if t.bounds.Pos.Y < 0 {
		return false
	}

	if t.bounds.Pos.Y + t.bounds.H > parent.H {
		return false
	}

	if t.bounds.Pos.X + t.bounds.W > parent.W {
		return false
	}

	return true
}

func (r *Text) SetLayer(l int) error{
	if l < 0{
		return errors.New("Layers must be greater or equal to 0")
	} 

	r.layer = l
	return nil
}

func (r *Text) GetLayer() int{
    return r.layer
}

func (t *Text) SetParentStyle(s *Style){
    t.parentStyle = s
	t.style.Bg = s.Bg
}

func (t *Text) Layout(parent geom.Bounds){
	t.layout.computedPos.X = resolveAxis(t.layout.anchor.H, parent.Pos.X, parent.W, t.bounds.W, t.bounds.Pos.X)
    t.layout.computedPos.Y = resolveAxis(t.layout.anchor.V, parent.Pos.Y, parent.H, t.bounds.H, t.bounds.Pos.Y)

}