package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Text struct {
	pos   *geom.Point
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
	
	t := &Text{
		pos:   pos,
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
		
		if t.style.Fg.IsTransparent {
			continue
		}
	
		r := rune(t.value[i])
		
		buf.Set(vec.X + x+i, vec.Y + y, r, t.style.Bg, t.style.Fg)
	}
}

func (t *Text) IsInBounds(parent geom.Bounds) bool{

	if t.pos.X < 0 {
		return false
	}

	if t.pos.Y < 0 {
		return false
	}

	if t.pos.Y + 1 > parent.H {
		return false
	}

	if t.pos.X + len(t.value) > parent.W {
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
}