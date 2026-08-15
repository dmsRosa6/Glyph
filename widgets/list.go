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
// ItemPadding is a horizontal inset per row. StackLayout.Arrange only
// ever calls ResolveAxis with a hardcoded 0 as the "original" X for
// NoAnchor children -- it never reads back a child's own declared
// Pos.X -- so a row built with Pos.X = itemPadding and no anchor would
// silently have that X thrown away and pinned to 0 every frame. Rows are
// therefore built narrower (W - 2*itemPadding) and anchored H: Center
// instead of positioned by raw X, since Center/Start/End are the only
// anchors StackLayout actually honors.
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

// AddItem creates a new row of the given height, horizontally centered
// with ItemPadding of inset on each side, adds it via the ordinary
// AddChild path, and returns it so the caller can add their own content
// -- exactly the main.go listDemo pattern
// (`row, _ := list.AddItem(4); row.AddChild(text)`).
func (l *List) AddItem(height int) (*canvas.Container, error) {
	w, listH := l.Size()
	if height > listH {
		height = listH
	}
	rowW := w - 2*l.itemPadding
	row, err := canvas.NewContainer(geom.NewBounds(0, 0, rowW, height), canvas.ContainerConfig{
		Style:  framework.Style{Bg: core.Transparent, Fg: core.Transparent},
		Anchor: framework.Anchor{H: framework.Center},
	})
	if err != nil {
		return nil, err
	}
	l.AddChild(row)
	return row, nil
}
