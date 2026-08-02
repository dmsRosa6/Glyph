package canvas

import "github.com/dmsRosa6/glyph/geom"

// List is now just a Container with StackLayout -- there's no bespoke
// items []*Box anymore. Because of that, List.AddChild (promoted
// straight from *Container) accepts ANY Drawable, not only boxed items;
// AddItem below is a convenience for the common "bordered row" case, not
// the only way to put something in a list.
type List struct {
	*Container
	itemStyle   Style
	itemPadding int
}

type ListConfig struct {
	Style       Style
	ItemPadding int
	Anchor      Anchor
	Layer       int
}

func NewList(bounds *geom.Bounds, cfg ListConfig) (*List, error) {
	c, err := NewContainer(bounds, ContainerConfig{
		Style:  cfg.Style,
		Layer:  cfg.Layer,
		Anchor: cfg.Anchor,
		Layout: StackLayout{},
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
	bounds := geom.NewBounds(0, 0, l.bounds.W, height)

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