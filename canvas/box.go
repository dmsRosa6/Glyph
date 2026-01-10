package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Box struct{
	border *Border
	composite *Composite
    layer int
}

type BoxConfig struct {
    BorderConfig BorderConfig

    Padding int

    Bg, Fg core.Color

    Layer int
}

//TODO Add colors to the composite
func NewBox(bounds geom.Bounds, cfg BoxConfig) (*Box,error) {
    if cfg.Padding < 0 {
        panic("padding must be >= 0")
    }

    b := &Box{}
    var br *Border
    var err error
    var c *Composite

    br, err = NewBorder(bounds,cfg.BorderConfig)    
    if err != nil {
        return nil ,err
    }

    b.border = br

    compositeBounds := geom.Bounds{
            Pos: geom.Point{
                X: bounds.Pos.X + cfg.Padding,
                Y: bounds.Pos.Y + cfg.Padding,
            },
            W: bounds.W - 2*cfg.Padding,
            H: bounds.H - 2*cfg.Padding,
        }

    c, err = NewComposite(compositeBounds, CompositeConfig{
        Layer: cfg.Layer,
    })
    if err != nil {
        return nil ,err
    }


    b.composite = c

    if err = b.SetLayer(cfg.Layer); err != nil {
        return nil ,err
    }

    return b, nil
}

func NewSimpleBox(
    x, y, w, h, thickness int,
    bg, fg, borderBg, borderFg core.Color,
) (*Box, error) {
    bounds := geom.Bounds{
            Pos: geom.Point{X: x, Y: y},
            W:   w,
            H:   h,
        }
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

func (b *Box) Draw(buf *core.Buffer){
	b.composite.Draw(buf)
	b.border.Draw(buf)
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