package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Text struct {
	Pos   *geom.Point
	Value string
	Bg, Fg core.Color
	
	layer int
}

func NewText(x, y, layer int, value string, bg, fg core.Color) (*Text, error) {
	t := &Text{
		Pos:   &geom.Point{X: x, Y: y},
		Value: value,
		Bg: bg,
		Fg: fg,
	}

	if err := t.SetLayer(layer); err != nil{
		return nil, err
	}

	return t, nil
}

func (t *Text) Draw(buf *core.Buffer, vec geom.Vector) {
	x := t.Pos.X
	y := t.Pos.Y

	for i := 0; i < len(t.Value); i++{
		
		if t.Fg.IsTransparent {
			continue
		}
	
		r := rune(t.Value[i])
		
		buf.Set(vec.X + x+i, vec.Y + y, r, t.Bg, t.Fg)
	}
}

func (t *Text) IsInBounds(parent geom.Bounds) bool{

	if t.Pos.X < 0 {
		return false
	}

	if t.Pos.Y < 0 {
		return false
	}

	if t.Pos.Y + 1 > parent.H {
		return false
	}

	if t.Pos.X + len(t.Value) > parent.W {
		return false
	}

	return true
}

func (t *Text) MoveTo(p *geom.Point) {
	t.Pos = p
}

func (t *Text) Translate(v geom.Vector) {
	t.Pos.AddVector(v)
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