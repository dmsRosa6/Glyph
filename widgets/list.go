package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// List is a container with a stacked layout, plus AddItem for adding
// bordered, padded rows.
//
// A List is FIXED by default: rows beyond the declared height are
// simply clipped, the same as every other Container already does --
// nothing new. Set ListConfig.Scrollable to opt into the other mode:
// as focus moves between rows (Tab/Shift+Tab, or
// FocusManager.Enter/Exit drilling into one), the list auto-scrolls
// just enough to keep the newly-focused row fully visible. It does
// NOT bind its own scroll keys or a mouse wheel -- List itself isn't
// Focusable (its ROWS are), so it has no natural place to receive
// input directly, and deciding how an app wants manual scrolling
// triggered (PageUp/PageDown? a wheel?) is an app-level call, not a
// framework one. ScrollBy/ScrollTo exist as the manual escape hatch:
// an app wires its own key or mouse binding (the same way
// examples/mouse-paint-demo already wires its own BindMouse) and
// calls one of those from it.
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
	// Scrollable opts into auto-scroll-to-keep-focus-visible -- see
	// the type's own doc comment. False (the default) is a plain fixed
	// list.
	Scrollable bool
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

// Draw shadows the promoted *canvas.Container.Draw purely to run
// followFocus first -- so a scroll adjustment triggered by a focus
// change lands in the SAME frame that change is drawn in, not a frame
// late. followFocus itself is a no-op on a non-Scrollable List.
func (l *List) Draw(buf *core.Buffer, vec geom.Vector) {
	l.followFocus()
	l.Container.Draw(buf, vec)
}

// ScrollBy/ScrollTo are List's manual escape hatch for an app that
// wants direct scroll control on top of the automatic follow-focus
// behavior -- see the type's own doc comment for why List doesn't wire
// a key or mouse binding to either of these itself. Both are no-ops on
// a non-Scrollable List.
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

// followFocus scans this List's current rows for whichever one is
// focused and, if it's not already fully inside the visible viewport,
// scrolls just far enough to bring it in. Purely reactive, recomputed
// fresh every call from (row positions, which row is focused, current
// scroll) rather than kept as separately-tracked state that could
// drift out of sync with any of those.
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

// keepVisible nudges the scroll offset just far enough to bring
// [top, bottom) fully inside the current viewport: scrolls up if the
// row is above it, down if below, and does nothing if it's already
// fully visible -- deliberately not re-centering it every time, which
// would make the list jump around more than necessary whenever focus
// moves by exactly one row within an already-fine viewport. A row
// taller than the whole viewport shows from its top rather than its
// bottom, since the "scrolled up" branch is checked first.
func (l *List) keepVisible(top, bottom, viewH int) {
	current := l.ScrollY()
	switch {
	case top < current:
		l.setClampedScroll(top)
	case bottom > current+viewH:
		l.setClampedScroll(bottom - viewH)
	}
}

// setClampedScroll is the one place that actually calls SetScrollY,
// so every caller (followFocus, ScrollBy, ScrollTo) gets the same
// [0, maxScroll] clamp for free rather than three copies of it.
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

// totalContentHeight sums every current row's height on demand rather
// than tracking a running total incrementally -- List has no
// RemoveItem of its own, but nothing stops a caller reaching the
// promoted Container.RemoveChild directly on a tracked row, and a
// live recomputation can never drift out of sync with whatever
// actually happened to the child set the way an incrementally
// maintained counter could.
func (l *List) totalContentHeight() int {
	total := 0
	for _, c := range l.Children() {
		total += rowHeight(c)
	}
	return total
}

// sizer is satisfied by every ListRow (Size() is promoted from
// mixin.Node via mixin.FocusableWrapper) -- a small structural
// interface rather than a concrete-type check, so anything else with
// the same Size() method added directly via the promoted
// Container.AddChild is measured the same way instead of silently
// contributing 0 to the scroll math.
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
