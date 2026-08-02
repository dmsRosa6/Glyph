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

// Window is a Bordered box with a title Text overlaid on the frame.
// The title genuinely isn't reducible to Container/Bordered alone (it
// sits on the border itself, outside the padded content area), so this
// stays a small dedicated composition -- but it's built on Bordered
// instead of owning a border and a content container directly.
type Window struct {
	BaseNode
	box  *Bordered
	text *Text
}

type WindowConfig struct {
	BoxConfig BoxConfig

	Title         string
	TitleXOffset  int
	TitlePosition TitlePosition
	TitleFg       core.Color

	Anchor Anchor
	Layer  int
}

func NewWindow(bounds *geom.Bounds, cfg WindowConfig) (*Window, error) {
	if cfg.Title != "" {
		innerWidth := bounds.W - 2*cfg.BoxConfig.Padding

		if cfg.TitleXOffset < 0 ||
			cfg.TitleXOffset+len(cfg.Title) > innerWidth {
			panic("title out of window bounds")
		}
	}

	base, err := newBaseNode(bounds, cfg.Anchor, cfg.BoxConfig.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	boxBounds := geom.NewBounds(0, 0, bounds.W, bounds.H)
	box, err := NewBox(boxBounds, cfg.BoxConfig)
	if err != nil {
		return nil, err
	}

	var textY int
	if cfg.TitlePosition == TitleBottom {
		textY = bounds.H - 1
	}

	var text *Text
	if cfg.Title != "" {
		textPos := geom.NewPoint(cfg.BoxConfig.Padding+cfg.TitleXOffset, textY)

		text, err = NewText(textPos, TextConfig{
			Value: cfg.Title,
			Fg:    cfg.TitleFg,
			Layer: cfg.Layer,
		})
		if err != nil {
			return nil, err
		}
	}

	w := &Window{
		BaseNode: base,
		box:      box,
		text:     text,
	}

	box.SetParentStyle(w.ResolvedStyle())
	if text != nil {
		text.SetParentStyle(box.ResolvedStyle())
	}

	return w, nil
}

func (w *Window) Draw(buf *core.Buffer, vec geom.Vector) {
	v := geom.Vector{X: vec.X + w.computedPos.X, Y: vec.Y + w.computedPos.Y}

	w.box.Draw(buf, v)
	if w.text != nil {
		w.text.Draw(buf, v)
	}
}

func (w *Window) AddChild(child Drawable) {
	w.box.AddChild(child)
}

func (w *Window) RemoveChild(target Drawable) {
	w.box.RemoveChild(target)
}

func (w *Window) SetLayer(l int) error {
	if err := w.BaseNode.SetLayer(l); err != nil {
		return err
	}
	if w.text != nil {
		if err := w.text.SetLayer(l); err != nil {
			return err
		}
	}
	return w.box.SetLayer(l)
}

func (w *Window) SetParentStyle(s *Style) {
	w.BaseNode.SetParentStyle(s)
	w.box.SetParentStyle(w.ResolvedStyle())
	if w.text != nil {
		w.text.SetParentStyle(w.box.ResolvedStyle())
	}
}

func (w *Window) SetInvalidator(fn func()) {
	w.BaseNode.SetInvalidator(fn)
	w.box.SetInvalidator(fn)
	if w.text != nil {
		w.text.SetInvalidator(fn)
	}
}
