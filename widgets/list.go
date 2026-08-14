package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// List stacks item rows top-to-bottom. layoutpolicy.go's StackLayout doc
// comment ("This is the old List behavior, generalized") is the evidence
// this is built on: List isn't its own composite hand-rolling
// style/invalidator/layer fan-out at all -- it's a thin wrapper around
// canvas.Container configured with StackLayout, so AddItem's rows go
// through the ordinary Container.AddChild path.
//
// That means List needed ZERO changes for this refactor, not because I
// applied Propagator to it, but because it was never a second copy of the
// hand-rolled pattern in the first place -- it rides on canvas.Container,
// which step 5 already fixed. I'm fairly confident about that shape;
// ItemPadding's exact meaning (horizontal inset per row, shown here, vs.
// vertical gap between rows) is a guess -- StackLayout as it stands has
// no gap concept, so I went with horizontal inset, matching Bordered's
// padding precedent. Worth confirming against the real file.
//
// RECONSTRUCTED -- see the note atop bordered.go.
type List struct {
	*canvas.Container
	itemPadding int
}

type ListConfig struct {
	Style       framework.Style
	ItemPadding int
	Layer       int
	Anchor      framework.Anchor
}

func NewList(bounds *geom.Bounds, cfg ListConfig) (*List, error) {
	c, err := canvas.NewContainer(bounds, canvas.ContainerConfig{
		Style:  cfg.Style,
		Layer:  cfg.Layer,
		Anchor: cfg.Anchor,
		Layout: framework.StackLayout{},
	})
	if err != nil {
		return nil, err
	}
	return &List{Container: c, itemPadding: cfg.ItemPadding}, nil
}

// AddItem creates a new row of the given height, inset horizontally by
// ItemPadding, adds it via the ordinary AddChild path, and returns it so
// the caller can add their own content -- exactly the main.go listDemo
// pattern (`row, _ := list.AddItem(4); row.AddChild(text)`).
func (l *List) AddItem(height int) (*canvas.Container, error) {
	_, listH := l.Size()
	if height > listH {
		height = listH
	}
	w, _ := l.Size()
	rowW := w - 2*l.itemPadding
	row, err := canvas.NewContainer(geom.NewBounds(l.itemPadding, 0, rowW, height), canvas.ContainerConfig{
		Style: framework.Style{Bg: core.Transparent, Fg: core.Transparent},
	})
	if err != nil {
		return nil, err
	}
	l.AddChild(row)
	return row, nil
}
