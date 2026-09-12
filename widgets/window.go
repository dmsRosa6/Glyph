package widgets

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
	"github.com/dmsRosa6/glyph/primitive"
)

type Window struct {
	mixin.FocusableWrapper
	title *primitive.Text
	raise func()
}

type WindowConfig struct {
	Padding      int
	BoxStyle     framework.Style
	FocusStyle   *framework.Style
	BorderConfig primitive.BorderConfig
	Anchor       framework.Anchor
	Layer        int
	Title        string
	TitleFg      core.Color
}

func NewWindow(bounds *geom.Bounds, cfg WindowConfig) (*Window, error) {
	bn, err := mixin.NewNode(bounds, cfg.Anchor, framework.Style{Bg: core.Transparent, Fg: core.Transparent}, cfg.Layer, "Window")
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
		FocusableWrapper: mixin.NewFocusableWrapper(bn, box),
	}
	if cfg.FocusStyle != nil {
		w.SetFocusStyle(*cfg.FocusStyle)
	}

	if cfg.Title != "" {
		title, err := primitive.NewText(&geom.Point{X: 1, Y: 0}, primitive.TextConfig{
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
	resolved := w.Style()
	inner := w.Inner()
	inner.SetParentStyle(&resolved)

	pos := w.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	wdt, hgt := w.Size()

	buf.PushClip(v.X, v.Y, wdt, hgt)
	defer buf.PopClip()

	inner.Draw(buf, v)
	if w.title != nil {
		w.title.SetParentStyle(&resolved)
		w.title.Draw(buf, v)
	}
}

func (w *Window) SetContext(ctx framework.AppContext) {
	w.FocusableWrapper.SetContext(ctx)
	if w.title != nil {
		w.title.SetContext(ctx)
	}
}

func (w *Window) SetRaiser(raise func()) {
	w.raise = raise
}

func (w *Window) RaiseToFront() {
	if w.raise != nil {
		w.raise()
	}
}

func (w *Window) Focus() {
	w.FocusBehavior.Focus()
	w.RaiseToFront()
}
