package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Box struct{
    bounds *geom.Bounds
	border *Border
	composite *Composite
    style *Style
    parentStyle *Style
    padding int
    layout *Layout
    layer int
}

type BoxConfig struct {
    BorderConfig BorderConfig

    Padding int

    Bg, Fg core.Color

    Anchor Anchor
    Layer int
}

//TODO Add colors to the composite
func NewBox(bounds *geom.Bounds, cfg BoxConfig) (*Box,error) {
    if cfg.Padding < 0 {
        panic("padding must be >= 0")
    }

    b := &Box{}
    var br *Border
    var err error
    var c *Composite

    b.bounds = bounds
    b.padding = cfg.Padding

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

    br, err = NewBorder(bounds,cfg.BorderConfig)    
    if err != nil {
        return nil ,err
    }

    b.border = br
    b.layout = &Layout{
					computedPos: bounds.Pos,
					anchor: &cfg.Anchor,
				}

    compositeBounds := geom.NewBounds(bounds.Pos.X + b.padding, bounds.Pos.Y + b.padding,
                                      bounds.W - 2*cfg.Padding, bounds.H - 2*cfg.Padding)

    c, err = NewComposite(compositeBounds, CompositeConfig{Layer: cfg.Layer})

    if err != nil {
        return nil ,err
    }

    b.composite = c
    b.style = s

    if err = b.SetLayer(cfg.Layer); err != nil {
        return nil ,err
    }

    return b, nil
}

func NewSimpleBox(
    x, y, w, h, thickness int,
    bg, fg, borderBg, borderFg core.Color, s *Style,
) (*Box, error) {
    bounds := geom.NewBounds(x,y,w,h)
    return NewBox(bounds, BoxConfig{
        Padding: thickness,
        BorderConfig: BorderConfig{
            Thickness: thickness,
            Style: UniformBorderStyle(' '),
            Fg: borderFg,
            Bg: borderBg,
        },
        Bg: bg,
        Fg: fg,
    })
}

func (b *Box) Draw(buf *core.Buffer, vec geom.Vector){
	v := geom.Vector{}
    v.AddVector(vec)
    v.AddVector(*geom.VectorFromPoint(*b.layout.computedPos))
    
    b.composite.Draw(buf, v)
	b.border.Draw(buf, vec)
}

func (r *Box) IsInBounds(parent geom.Bounds) bool{
	return r.composite.IsInBounds(parent) && r.border.IsInBounds(parent)	
}


func (r *Box) SetLayer(l int) error{
    if l < 0{
		return errors.New("Layers must be greater or equal to 0")
	} 
	
	r.layer = l
    r.border.SetLayer(l)
    r.composite.SetLayer(l)

	return nil
}

func (r *Box) GetLayer() int{
    return r.layer
}

func (b *Box) AddChild(child Drawable){
    b.composite.AddChild(child)
}

func (b *Box) RemoveChild(target Drawable) {
	b.composite.AddChild(target)
}

func (b *Box) Layout(parent geom.Bounds) {
    b.layout.computedPos.X = resolveAxis(b.layout.anchor.H, parent.Pos.X, parent.W, b.bounds.W, b.bounds.Pos.X)
    b.layout.computedPos.Y = resolveAxis(b.layout.anchor.H, parent.Pos.Y, parent.H, b.bounds.H, b.bounds.Pos.Y)
}

func (b *Box) SetParentStyle(s *Style){
    b.parentStyle = s
    b.border.SetParentStyle(s)
    b.composite.SetParentStyle(s)
}