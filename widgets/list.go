package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// List is a container with a stacked layout, plus AddItem for adding
// bordered, padded rows. Fixed by default (overflow clips, like any
// container). Set ListConfig.Scrollable to auto-scroll and keep the
// focused row visible as focus moves between rows. List isn't
// Focusable itself, so it doesn't bind a scroll key or mouse wheel --
// ScrollBy/ScrollTo are the manual escape hatch for an app that wants
// one.
type List struct {
	*canvas.Container
	itemPadding int
	rowFocus    *framework.Style
	scrollable  bool
}

type ListConfig struct {
	Style         framework.Style
	ItemPadding   int
	Layer         int
	Anchor        framework.Anchor
	RowFocusStyle *framework.Style
	Scrollable    bool
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
		itemPadding: cfg.ItemPadding,
		rowFocus:    cfg.RowFocusStyle,
		scrollable:  cfg.Scrollable,
	}, nil
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

// Draw runs followFocus before the promoted Container.Draw, so a
// scroll adjustment lands in the same frame the focus change is drawn.
func (l *List) Draw(buf *core.Buffer, vec geom.Vector) {
	l.followFocus()
	l.Container.Draw(buf, vec)
}

func (l *List) ScrollBy(delta int) {
	if !l.scrollable {
		return
	}
	l.setClampedScroll(l.ScrollY() + delta)
}

func (l *List) ScrollTo(y int) {
	if !l.scrollable {
		return
	}
	l.setClampedScroll(y)
}

// followFocus scrolls just far enough to keep the focused row visible.
func (l *List) followFocus() {
	if !l.scrollable {
		return
	}

	children := l.Children()
	_, viewH := l.Size()

	y := 0
	for _, c := range children {
		h := rowHeight(c)
		if f, ok := c.(framework.Focusable); ok && f.IsFocused() {
			l.keepVisible(y, y+h, viewH)
			return
		}
		y += h
	}
}

// keepVisible nudges scroll just far enough to bring [top, bottom)
// fully into the viewport; does nothing if it's already visible.
func (l *List) keepVisible(top, bottom, viewH int) {
	current := l.ScrollY()
	switch {
	case top < current:
		l.setClampedScroll(top)
	case bottom > current+viewH:
		l.setClampedScroll(bottom - viewH)
	}
}

func (l *List) setClampedScroll(y int) {
	_, viewH := l.Size()
	max := l.totalContentHeight() - viewH
	if max < 0 {
		max = 0
	}
	switch {
	case y < 0:
		y = 0
	case y > max:
		y = max
	}
	l.SetScrollY(y)
}

func (l *List) totalContentHeight() int {
	total := 0
	for _, c := range l.Children() {
		total += rowHeight(c)
	}
	return total
}

type sizer interface {
	Size() (int, int)
}

func rowHeight(d framework.Drawable) int {
	if s, ok := d.(sizer); ok {
		_, h := s.Size()
		return h
	}
	return 0
}
