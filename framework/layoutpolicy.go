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
//
// Arrange returns the subset of children it could NOT position (because
// they didn't satisfy the type assertion the policy needs). Previously
// these were dropped silently with no way for the caller to know why a
// child never showed up where expected; Container now logs a warning for
// each one it gets back.
type LayoutPolicy interface {
	Arrange(children []Drawable, frame geom.Bounds) (skipped []Drawable)
}

// CapacityAwareLayout is an optional extension a LayoutPolicy can
// implement when "does this child fit" depends on more than just that
// child's own bounds -- e.g. StackLayout, where fitting depends on the
// cumulative height of everything already stacked above it. Container's
// AddChild consults this (when present) in addition to the ordinary
// single-child IsInBounds check.
type CapacityAwareLayout interface {
	// Fits reports whether candidate can be added to existing without
	// overflowing frame, according to this policy's arrangement rules.
	Fits(existing []Drawable, candidate Drawable, frame geom.Bounds) bool
}

// layoutTarget is what a LayoutPolicy needs from a child to position it
// directly. Every shape gets this for free by embedding BaseNode.
type layoutTarget interface {
	Size() (w, h int)
	SetComputedPos(x, y int)
	AnchorH() AxisAnchor
}

// xOriginProvider is an optional extension of layoutTarget: a child that
// implements it can hand back its own declared X position (BaseNode
// does, via Bounds()). Policies that need to fall back to "wherever the
// child was constructed at" for NoAnchor children use this; a child that
// doesn't implement it just falls back to 0, same as before.
type xOriginProvider interface {
	Bounds() *geom.Bounds
}

// FreeLayout leaves each child at its own declared position, resolving
// only its own anchor (if any) against the container's frame. This is
// the old Composite behavior.
type FreeLayout struct{}

func (FreeLayout) Arrange(children []Drawable, frame geom.Bounds) (skipped []Drawable) {
	for _, child := range children {
		l, ok := child.(Layoutable)
		if !ok {
			skipped = append(skipped, child)
			continue
		}
		l.Layout(frame)
	}
	return skipped
}

// StackLayout stacks children top-to-bottom in insertion order, ignoring
// each child's declared Y position entirely (X still honors each child's
// horizontal anchor within the container's width; a NoAnchor child now
// keeps its own declared X instead of being silently pinned to 0 -- see
// xOriginProvider above).
type StackLayout struct{}

func (StackLayout) Arrange(children []Drawable, frame geom.Bounds) (skipped []Drawable) {
	y := 0
	for _, child := range children {
		lt, ok := child.(layoutTarget)
		if !ok {
			skipped = append(skipped, child)
			continue
		}
		w, h := lt.Size()
		origX := 0
		if xp, ok := child.(xOriginProvider); ok {
			origX = xp.Bounds().Pos.X
		}
		x := ResolveAxis(lt.AnchorH(), frame.W, w, origX)
		lt.SetComputedPos(x, y)
		y += h
	}
	return skipped
}

// Fits makes StackLayout a CapacityAwareLayout: candidate fits only if
// the sum of every already-tracked child's height, plus candidate's own
// height, stays within frame.H. Children that aren't a layoutTarget are
// skipped in the sum the same way Arrange skips them -- they never
// occupy stack space in the first place.
func (StackLayout) Fits(existing []Drawable, candidate Drawable, frame geom.Bounds) bool {
	total := 0
	for _, c := range existing {
		lt, ok := c.(layoutTarget)
		if !ok {
			continue
		}
		_, h := lt.Size()
		total += h
	}

	lt, ok := candidate.(layoutTarget)
	if !ok {
		// Not a layoutTarget: Arrange will skip it too, so it never
		// consumes stack space -- nothing to reject it for here.
		return true
	}
	_, h := lt.Size()
	total += h

	return total <= frame.H
}
