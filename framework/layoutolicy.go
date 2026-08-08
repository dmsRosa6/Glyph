package framework

import (
	"github.com/dmsRosa6/glyph/geom"
)

// LayoutPolicy decides where each child sits within a Container's
// interior. Container calls Arrange once per Draw, before drawing each
// child, so a policy reacts to children being added, removed, or
// resized between frames with no manual bookkeeping in the container.
//
// This is what replaces having a separate bespoke type for every
// arrangement (Composite for free-form, List for stacked). Adding a new
// arrangement later -- a horizontal row, a grid -- is a new LayoutPolicy,
// not a new container type.
type LayoutPolicy interface {
	Arrange(children []Drawable, frame geom.Bounds)
}

// layoutTarget is what a LayoutPolicy needs from a child to position it
// directly. Every shape gets this for free by embedding BaseNode.
type layoutTarget interface {
	Size() (w, h int)
	SetComputedPos(x, y int)
	AnchorH() AxisAnchor
}

// FreeLayout leaves each child at its own declared position, resolving
// only its own anchor (if any) against the container's frame. This is
// the old Composite behavior.
type FreeLayout struct{}

func (FreeLayout) Arrange(children []Drawable, frame geom.Bounds) {
	for _, child := range children {
		if l, ok := child.(Layoutable); ok {
			l.Layout(frame)
		}
	}
}

// StackLayout stacks children top-to-bottom in insertion order, ignoring
// each child's declared Y position entirely (X still honors each child's
// horizontal anchor within the container's width). This is the old List
// behavior, generalized to work on any Drawable rather than only *Box.
//
// Note: this only checks that each item individually fits within the
// container's height (via IsInBounds at AddChild time) -- it doesn't
// currently reject an AddChild that would make the *cumulative* stack
// taller than the container. Overflow just draws past the bottom edge.
// Worth tightening if you hit it in practice.
type StackLayout struct{}

func (StackLayout) Arrange(children []Drawable, frame geom.Bounds) {
	y := 0
	for _, child := range children {
		lt, ok := child.(layoutTarget)
		if !ok {
			continue
		}
		w, h := lt.Size()
		x := ResolveAxis(lt.AnchorH(), frame.W, w, 0)
		lt.SetComputedPos(x, y)
		y += h
	}
}