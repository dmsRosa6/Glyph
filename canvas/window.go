package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type TitlePosition int

const (
    TitleTop TitlePosition = iota
    TitleBottom
)

type Window struct {
    box  *Box
    text *Text

    layer int
}

type WindowConfig struct {
    Bounds geom.Bounds

    Box BoxConfig

    Title          string
    TitleXOffset   int
    TitlePosition  TitlePosition
    TitleFg        core.Color
    Layer int
}

func NewWindow(cfg WindowConfig) (*Window, error) {

    if cfg.Title != "" {
        innerWidth := cfg.Bounds.W - 2*cfg.Box.Padding

        if cfg.TitleXOffset < 0 ||
            cfg.TitleXOffset+len(cfg.Title) > innerWidth {
            panic("title out of window bounds")
        }
    }

    var err error
    var box *Box

    box, err = NewBox(cfg.Box)
    if err != nil {
        return nil, err
    }

    var textY int
    switch cfg.TitlePosition {
    case TitleTop:
        textY = cfg.Bounds.Pos.Y
    case TitleBottom:
        textY = cfg.Bounds.Pos.Y + cfg.Bounds.H - 1
    }

    var text *Text
    if cfg.Title != "" {
        text, err = NewText(
            cfg.Bounds.Pos.X+cfg.Box.Padding+cfg.TitleXOffset,
            textY,
            cfg.Layer,
            cfg.Title,
            core.Transparent,
            cfg.TitleFg,
        )
        if err != nil {
            return nil, err
        }
    }

    w :=  &Window{
        box:  box,
        text: text,
    }

    if err = w.SetLayer(cfg.Layer); err != nil {
        return nil, err
    }

    return w, nil
}

func (w *Window) Draw(buf *core.Buffer) {
    w.box.Draw(buf)
    if w.text != nil {
        w.text.Draw(buf)
    }
}

func (w *Window) IsInBounds(parent geom.Bounds) bool {
    return w.box.IsInBounds(parent)
}


func (r *Window) SetLayer(l int) error{
    if l < 0{
		return errors.New("Layers must be greater or equal to 0")
	}
    
    r.box.border.SetLayer(l)
    r.box.composite.SetLayer(l)
    r.layer = l

    return nil
}

func (r *Window) GetLayer() int{
    return r.layer
}
