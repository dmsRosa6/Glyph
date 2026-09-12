package framework

import (
	"github.com/dmsRosa6/glyph/geom"
)

type LayoutPolicy interface {
	Arrange(children []Drawable, frame geom.Bounds) (skipped []Drawable)
}

type CapacityAwareLayout interface {
	Fits(existing []Drawable, candidate Drawable, frame geom.Bounds) bool
}

type layoutTarget interface {
	Size() (w, h int)
	SetComputedPos(x, y int)
	AnchorH() AxisAnchor
}

type xOriginProvider interface {
	Bounds() *geom.Bounds
}

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
		return true
	}
	_, h := lt.Size()
	total += h

	return total <= frame.H
}
