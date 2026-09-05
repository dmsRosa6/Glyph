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
	SetContext(ctx AppContext)
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

type Navigator interface {
	Next()
	Prev()
	Enter() bool
	Exit()
	Current() Focusable
}

type Identifiable interface {
	ID() string
}

type Raisable interface {
	SetRaiser(raise func())
}

// MouseHandler is implemented by a Drawable that wants to react to
// mouse input. Deliberately separate from Focusable.HandleInput: a key
// event is dispatched to whichever widget currently holds keyboard
// focus, but a mouse event is inherently positional -- it targets
// whatever is under Event.MouseX/MouseY, not whatever's focused.
//
// Nothing in this codebase currently performs that positional dispatch
// (hit-testing: mapping a screen coordinate to the Drawable whose
// ABSOLUTE screen bounds contain it). BaseNode.ComputedPos is only
// relative to its own parent -- there's no existing way to ask "what
// are this widget's bounds in absolute screen space" without a tree
// walk accumulating every ancestor's offset, and Container doesn't
// track that today. This interface exists so that whenever hit-testing
// is designed, it has a settled target to dispatch into rather than
// inventing this signature at the same time as the tree-walk logic --
// see input.Manager's mouse decoding and app.App.Run's dispatch loop,
// which currently logs decoded mouse events but does not route them
// here yet.
type MouseHandler interface {
	Drawable
	HandleMouse(ev Event) (bool, error)
}
