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
    bounds *geom.Bounds

    titleOffset int

    layout *Layout
    layer int
}

type WindowConfig struct {
    BoxConfig BoxConfig

    Title          string
    TitleXOffset   int
    TitlePosition  TitlePosition
    TitleFg        core.Color

    Anchor Anchor
    Layer int
}

func NewWindow(bounds *geom.Bounds, cfg WindowConfig) (*Window, error) {

    if cfg.Title != "" {
        innerWidth := bounds.W - 2*cfg.BoxConfig.Padding

        if cfg.TitleXOffset < 0 ||
            cfg.TitleXOffset+len(cfg.Title) > innerWidth {
            panic("title out of window bounds")
        }
    }

    var err error
    var box *Box

    box, err = NewBox(bounds, cfg.BoxConfig)
    if err != nil {
        return nil, err
    }

    var textY int
    switch cfg.TitlePosition {
    case TitleTop:
        textY = 0
    case TitleBottom:
        textY =  bounds.H - 1
    }

    var text *Text
    if cfg.Title != "" {

        textPos := geom.NewPoint(bounds.Pos.X+cfg.BoxConfig.Padding+cfg.TitleXOffset, textY)
        
        textCfg := TextConfig{
            Value : cfg.Title,
            Fg : cfg.TitleFg,
            Bg : core.Transparent,
            Anchor : cfg.Anchor,
            Layer : cfg.Layer,
        }
        
        text, err = NewText(textPos,textCfg)
        if err != nil {
            return nil, err
        }
    }

    w :=  &Window{
        box:  box,
        text: text,
    }

    w.bounds = bounds
    w.titleOffset = cfg.TitleXOffset
    w.layout = &Layout{
        computedPos: bounds.Pos,
        anchor: &cfg.Anchor,
    }

    if err = w.SetLayer(cfg.Layer); err != nil {
        return nil, err
    }

    return w, nil
}

func (w *Window) Draw(buf *core.Buffer, vec geom.Vector) {
    
    v := geom.Vector{}

    v.AddVector(vec)
    v.AddVector(*geom.VectorFromPoint(*w.layout.computedPos))

    w.box.Draw(buf, vec)
    if w.text != nil {
        w.text.Draw(buf, v)
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

func (b *Window) AddChild(child Drawable){
    b.box.AddChild(child)
}

func (b *Window) RemoveChild(target Drawable) {
	b.box.AddChild(target)
}

func (w *Window) Layout(parent geom.Bounds) {
    w.layout.computedPos.X = resolveAxis(w.layout.anchor.H, parent.Pos.X, parent.W, w.bounds.W, w.bounds.Pos.X)
    w.layout.computedPos.Y = resolveAxis(w.layout.anchor.V, parent.Pos.Y, parent.H, w.bounds.H, w.bounds.Pos.Y)
}
