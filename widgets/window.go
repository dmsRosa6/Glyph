package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Window struct {
	base.FocusableBaseNode
	box   *Bordered
	title *Text

	raise func()
}

type WindowConfig struct {
	Padding      int
	BoxStyle     framework.Style
	FocusStyle   *framework.Style
	BorderConfig BorderConfig
	Anchor       framework.Anchor
	Layer        int
	Title        string
	TitleFg      core.Color
}

func NewWindow(bounds *geom.Bounds, cfg WindowConfig) (*Window, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, framework.Style{Bg: core.Transparent, Fg: core.Transparent}, cfg.Layer, "Window")
	if err != nil {
		return nil, err
	}

	box, err := NewBox(geom.NewBounds(0, 0, bounds.W, bounds.H), BoxConfig{
		Padding:      cfg.Padding,
		Style:        cfg.BoxStyle,
		BorderConfig: cfg.BorderConfig,
	})
	if err != nil {
		return nil, err
	}

	w := &Window{
		FocusableBaseNode: base.NewFocusableBaseNode(bn),
		box:               box,
	}
	if cfg.FocusStyle != nil {
		w.FocusableBaseNode.SetFocusStyle(*cfg.FocusStyle)
	}

	if cfg.Title != "" {
		title, err := NewText(&geom.Point{X: 1, Y: 0}, TextConfig{
			Value: cfg.Title,
			Fg:    cfg.TitleFg,
		})
		if err != nil {
			return nil, err
		}
		w.title = title
	}

	return w, nil
}

func (w *Window) Draw(buf *core.Buffer, vec geom.Vector) {
	resolved := w.FocusableBaseNode.Style()
	w.box.SetParentStyle(&resolved)
	w.title.SetParentStyle(&resolved)
	pos := w.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	wdt, hgt := w.Size()

	buf.PushClip(v.X, v.Y, wdt, hgt)
	defer buf.PopClip()

	w.box.Draw(buf, v)
	if w.title != nil {
		w.title.Draw(buf, v)
	}
}

func (w *Window) AddChild(child framework.Drawable) {
	w.box.AddChild(child)
}

func (w *Window) RemoveChild(target framework.Drawable) {
	w.box.RemoveChild(target)
}

func (w *Window) Children() []framework.Drawable {
	return w.box.Children()
}

func (w *Window) SetParentStyle(s *framework.Style) {
	w.FocusableBaseNode.SetParentStyle(s)
	resolved := w.FocusableBaseNode.Style()
	w.box.SetParentStyle(&resolved)
}

func (w *Window) SetContext(ctx framework.AppContext) {
	w.FocusableBaseNode.SetContext(ctx)
	w.box.SetContext(ctx)
	if w.title != nil {
		w.title.SetContext(ctx)
	}
}

func (w *Window) SetLayer(l int) error {
	return w.FocusableBaseNode.SetLayer(l)
}

// SetRaiser implements framework.Raisable. Wired automatically the
// instant this Window is tracked by any Propagator.
func (w *Window) SetRaiser(raise func()) {
	w.raise = raise
}

func (w *Window) RaiseToFront() {
	if w.raise != nil {
		w.raise()
	}
}

func (w *Window) FocusableChildren() []framework.Focusable {
	var out []framework.Focusable
	for _, c := range w.box.Children() {
		if f, ok := c.(framework.Focusable); ok {
			out = append(out, f)
		}
	}
	return out
}

func (w *Window) Focus() {
	w.FocusableBaseNode.Focus()
	w.RaiseToFront()
}
