package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type List struct {
	*canvas.Container
	itemStyle   framework.Style
	itemPadding int
}

type ListConfig struct {
	Style       framework.Style
	ItemPadding int
	Anchor      framework.Anchor
	Layer       int
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

	return &List{
		Container:   c,
		itemStyle:   cfg.Style,
		itemPadding: cfg.ItemPadding,
	}, nil
}

// AddItem creates a bordered, padded box of the given height (full list
// width) and stacks it as the next item. If you don't want the
// per-item border/padding, skip this and call list.AddChild directly
// with whatever Drawable you want stacked instead.
func (l *List) AddItem(height int) (*Bordered, error) {
	w, _ := l.Size()
	bounds := geom.NewBounds(0, 0, w, height)

	box, err := NewBox(bounds, BoxConfig{
		Padding:      l.itemPadding,
		Style:        l.itemStyle,
		Layer:        l.GetLayer(),
		BorderConfig: DefaultBorderConfig(),
	})
	if err != nil {
		return nil, err
	}

	l.AddChild(box)
	return box, nil
}