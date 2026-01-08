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
    Bounds geom.Bounds

    Border BorderConfig

    Padding int

    Bg, Fg core.Color

    Layer int
}


//TODO Add colors to the composite
func NewBox(cfg BoxConfig) (*Box,error) {
    if cfg.Padding < 0 {
        panic("padding must be >= 0")
    }

    b := &Box{}
    var br *Border
    var err error
    var c *Composite

    br, err = NewBorder(cfg.Border)    
    if err != nil {
        return nil ,err
    }

    b.border = br

    c, err = NewComposite(CompositeConfig{
        Bounds: geom.Bounds{
            Pos: geom.Point{
                X: cfg.Bounds.Pos.X + cfg.Padding,
                Y: cfg.Bounds.Pos.Y + cfg.Padding,
            },
            W: cfg.Bounds.W - 2*cfg.Padding,
            H: cfg.Bounds.H - 2*cfg.Padding,
        },
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
    return NewBox(BoxConfig{
        Bounds: geom.Bounds{
            Pos: geom.Point{X: x, Y: y},
            W:   w,
            H:   h,
        },
        Padding: thickness,
        Border: BorderConfig{
            Bounds: geom.Bounds{
                Pos: geom.Point{X: x, Y: y},
                W:   w,
                H:   h,
            },
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
	b.border.Draw(buf)
	b.composite.Draw(buf)
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