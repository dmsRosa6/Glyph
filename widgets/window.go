package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Window struct {
	base.BaseNode
	base.FocusBehavior
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
		BaseNode:      bn,
		FocusBehavior: base.NewFocusBehavior("Window"),
		box:           box,
	}
	if cfg.FocusStyle != nil {
		w.SetFocusStyle(*cfg.FocusStyle)
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

func (w *Window) Style() framework.Style {
	return w.FocusBehavior.ResolveFocusStyle(w.BaseNode.Style())
}

func (w *Window) Draw(buf *core.Buffer, vec geom.Vector) {
	resolved := w.Style()
	w.box.SetParentStyle(&resolved)
	pos := w.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	wdt, hgt := w.Size()

	buf.PushClip(v.X, v.Y, wdt, hgt)
	defer buf.PopClip()

	w.box.Draw(buf, v)
	if w.title != nil {
		// Bug fix in passing: this used to run unconditionally, one
		// call above where it belongs here -- a titleless Window
		// (cfg.Title == "") left w.title nil and would panic here
		// before ever reaching this same nil check further down.
		w.title.SetParentStyle(&resolved)
		w.title.Draw(buf, v)
	}
}

// Resize resizes both halves Window actually owns: its own BaseNode
// (bounds/anchor bookkeeping, e.g. what IsInBounds checks against) and
// box, the Bordered doing the real drawing. Without this, box's border
// and panel keep whatever size NewWindow originally gave them forever,
// while Window itself reports a new size nothing visible backs up.
func (w *Window) Resize(width, height int) {
	w.BaseNode.Resize(width, height)
	w.box.Resize(width, height)
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
	w.BaseNode.SetParentStyle(s)
	resolved := w.Style()
	w.box.SetParentStyle(&resolved)
}

func (w *Window) SetContext(ctx framework.AppContext) {
	w.BaseNode.SetContext(ctx)
	w.FocusBehavior.SetFocusContext(ctx, w.BaseNode.ID())
	w.box.SetContext(ctx)
	if w.title != nil {
		w.title.SetContext(ctx)
	}
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
	w.FocusBehavior.Focus()
	w.RaiseToFront()
}
