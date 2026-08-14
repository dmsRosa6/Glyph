package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// Window is a Bordered plus an optional title. Same v2 shape as Bordered:
// embeds *canvas.Container so propagation/Draw/layer-sort are all
// inherited, and only overrides the three methods that need to redirect
// to the right inner slot.
//
// RECONSTRUCTED -- see the note atop bordered.go.
type Window struct {
	*canvas.Container
	box   *Bordered
	title *Text // nil when cfg.Title == ""
}

type WindowConfig struct {
	Padding      int
	BoxStyle     framework.Style
	BorderConfig BorderConfig
	Anchor       framework.Anchor
	Layer        int
	Title        string
	TitleFg      core.Color
}

func NewWindow(bounds *geom.Bounds, cfg WindowConfig) (*Window, error) {
	outer, err := canvas.NewContainer(bounds, canvas.ContainerConfig{
		Style:  framework.Style{Bg: core.Transparent, Fg: core.Transparent},
		Layer:  cfg.Layer,
		Anchor: cfg.Anchor,
	})
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
	outer.AddChild(box)

	w := &Window{Container: outer, box: box}

	// Plain nil check, not Propagator's reflect-based one: title is a
	// concrete *Text or genuinely absent here, no typed-nil-through-an-
	// interface case to guard against, since we only ever call AddChild
	// with a value we just constructed ourselves.
	if cfg.Title != "" {
		title, err := NewText(&geom.Point{X: 1, Y: 0}, TextConfig{
			Value: cfg.Title,
			Fg:    cfg.TitleFg,
		})
		if err != nil {
			return nil, err
		}
		w.title = title
		outer.AddChild(title)
	}

	return w, nil
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
