package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
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
	base.BaseNode
	box  *Bordered
	text *Text
}

type WindowConfig struct {
	Padding      int
	BoxStyle     framework.Style
	BorderConfig BorderConfig

	Title         string
	TitleXOffset  int
	TitlePosition TitlePosition
	TitleFg       core.Color

	Anchor framework.Anchor
	Layer  int
}

func NewWindow(bounds *geom.Bounds, cfg WindowConfig) (*Window, error) {
	if cfg.Title != "" {
		innerWidth := bounds.W - 2*cfg.Padding

		if cfg.TitleXOffset < 0 ||
			cfg.TitleXOffset+len(cfg.Title) > innerWidth {
			panic("title out of window bounds")
		}
	}

	base, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.BoxStyle, cfg.Layer)
	if err != nil {
		return nil, err
	}

	// Anchor deliberately omitted -- Window's internal box is "hand-drawn" by the widget itself
	boxBounds := geom.NewBounds(0, 0, bounds.W, bounds.H)
	box, err := NewBox(boxBounds, BoxConfig{
		Padding:      cfg.Padding,
		Style:        cfg.BoxStyle,
		BorderConfig: cfg.BorderConfig,
		Layer:        cfg.Layer, 
	})
	if err != nil {
		return nil, err
	}

	var textY int
	if cfg.TitlePosition == TitleBottom {
		textY = bounds.H - 1
	}

	var text *Text
	if cfg.Title != "" {
		textPos := geom.NewPoint(cfg.Padding+cfg.TitleXOffset, textY)

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
	pos := w.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}

	w.box.Draw(buf, v)
	if w.text != nil {
		w.text.Draw(buf, v)
	}
}

func (w *Window) AddChild(child framework.Drawable) {
	w.box.AddChild(child)
}

func (w *Window) RemoveChild(target framework.Drawable) {
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

func (w *Window) SetParentStyle(s *framework.Style) {
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