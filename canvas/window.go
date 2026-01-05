package canvas

import (
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
}

type WindowConfig struct {
    Bounds geom.Bounds

    Box BoxConfig

    Title          string
    TitleXOffset   int
    TitlePosition  TitlePosition
    TitleFg        core.Color
}

func NewWindow(cfg WindowConfig) *Window {

    if cfg.Title != "" {
        innerWidth := cfg.Bounds.W - 2*cfg.Box.Padding

        if cfg.TitleXOffset < 0 ||
            cfg.TitleXOffset+len(cfg.Title) > innerWidth {
            panic("title out of window bounds")
        }
    }

    box := NewBox(cfg.Box)

    var textY int
    switch cfg.TitlePosition {
    case TitleTop:
        textY = cfg.Bounds.Pos.Y
    case TitleBottom:
        textY = cfg.Bounds.Pos.Y + cfg.Bounds.H - 1
    }

    var text *Text
    if cfg.Title != "" {
        text = NewText(
            cfg.Bounds.Pos.X+cfg.Box.Padding+cfg.TitleXOffset,
            textY,
            cfg.Title,
            core.Transparent,
            cfg.TitleFg,
        )
    }

    return &Window{
        box:  box,
        text: text,
    }
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
