package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type List struct {
	items       []*Box
	box         *Box
	layer       int
	parentStyle *Style
	style       *Style
}

type ListConfig struct {
	Box    *Box
	Style  Style
	Layer  int
	Anchor Anchor
	Padding int
}

func NewList(cfg ListConfig) (*List, error) {
	l := &List{
		box:   cfg.Box,
		items: []*Box{},
		layer: cfg.Layer,
		style: &cfg.Style,
	}
	return l, nil
}

func (l *List) Draw(buf *core.Buffer, vec geom.Vector) {
	v := vec

	for _, item := range l.items {
		item.Draw(buf, v)
	}
}

func (l *List) IsInBounds(parent geom.Bounds) bool {
	return l.box.IsInBounds(parent)
}

func (l *List) SetLayer(lay int) error {
	if lay < 0 {
		return errors.New("")
	}
	l.layer = lay
	return nil
}

func (l *List) GetLayer() int {
	return l.layer
}

func (l *List) SetParentStyle(s *Style) {
	l.parentStyle = s
	l.box.SetParentStyle(s)
	for _, item := range l.items {
		item.SetParentStyle(s)
	}
}

func (l *List) AddItem(item *Box) {
	l.items = append(l.items, item)
}
