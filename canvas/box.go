package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Box struct{
	border *Border
	composite *Composite
}

type BoxConfig struct {
    Bounds geom.Bounds

    Border BorderConfig

    Padding int

    Bg, Fg core.Color
}


//TODO Add colors to the composite
func NewBox(cfg BoxConfig) *Box {
    if cfg.Padding < 0 {
        panic("padding must be >= 0")
    }

    b := &Box{}

    b.border = NewBorder(cfg.Border)

    b.composite = NewComposite(CompositeConfig{
        Bounds: geom.Bounds{
            Pos: geom.Point{
                X: cfg.Bounds.Pos.X + cfg.Padding,
                Y: cfg.Bounds.Pos.Y + cfg.Padding,
            },
            W: cfg.Bounds.W - 2*cfg.Padding,
            H: cfg.Bounds.H - 2*cfg.Padding,
        },
    })

    return b
}

func NewSimpleBox(
    x, y, w, h, thickness int,
    bg, fg, borderBg, borderFg core.Color,
) *Box {
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