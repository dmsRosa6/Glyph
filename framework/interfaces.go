package framework

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Drawable interface {
	Draw(buf *core.Buffer, vec geom.Vector)
	IsInBounds(parent geom.Bounds) bool
	SetLayer(l int) error
	GetLayer() int
	SetParentStyle(style *Style)
	SetLogChannel(ch chan<- core.AppLog)
	// SetInvalidator wires up the callback a node (or anything nested
	// inside it) should call after a background update to ask for a
	// redraw. Containers propagate this to their children the same way
	// they already propagate SetParentStyle. Every shape gets a working
	// no-op implementation for free via BaseNode, so this never panics
	// on an unattached node -- it just does nothing until something
	// wires it up (Canvas/Renderer does this automatically).
	SetInvalidator(fn func())
}

type Moveable interface {
	MoveTo(p *geom.Point)
	Translate(v geom.Vector)
}

type Layoutable interface {
	Layout(parent geom.Bounds)
}

type Composable interface {
	AddChild(child Drawable)
	RemoveChild(target Drawable)
}

type Clippable interface {
	Drawable
	SetClip(clip geom.Bounds)
}
type Focusable interface {
	Drawable
	HandleInput(ev Event) (bool, error)
	Focus()
	Blur()
	IsFocused() bool
}

type FocusContainer interface {
	Focusable
	FocusableChildren() []Focusable
}

type ChildrenLister interface {
	Children() []Drawable
}
