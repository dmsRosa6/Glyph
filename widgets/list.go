package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type List struct {
	*canvas.Container
	itemPadding int
	rowFocus    *framework.Style
}

type ListConfig struct {
	Style         framework.Style
	ItemPadding   int
	Layer         int
	Anchor        framework.Anchor
	RowFocusStyle *framework.Style
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
	return &List{Container: c, itemPadding: cfg.ItemPadding, rowFocus: cfg.RowFocusStyle}, nil
}

func (l *List) AddItem(height int) (*ListRow, error) {
	w, listH := l.Size()
	if height > listH {
		height = listH
	}
	rowW := w - 2*l.itemPadding
	row, err := newListRow(geom.NewBounds(0, 0, rowW, height), framework.Anchor{H: framework.Center})
	if err != nil {
		return nil, err
	}
	if l.rowFocus != nil {
		row.SetFocusStyle(*l.rowFocus)
	}
	l.AddChild(row)
	return row, nil
}
