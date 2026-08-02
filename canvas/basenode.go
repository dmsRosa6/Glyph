package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

// BaseNode holds everything that used to be copy-pasted into every shape:
// layering, style inheritance, anchored layout, bounds checking, and now
// redraw invalidation. Every concrete shape (Rect, Border, Container,
// Bordered, Text, Window, List) embeds BaseNode and only implements
// Draw (plus AddChild/RemoveChild if it holds children). One
// implementation of each of these behaviors means one place to get it
// right, instead of N places to independently get it wrong.
type BaseNode struct {
	bounds *geom.Bounds
	anchor Anchor

	// computedPos is always a COPY of bounds.Pos, taken once at
	// construction and updated in place by Layout()/SetComputedPos().
	// It never aliases bounds.Pos's underlying pointer, so resolving
	// layout can never corrupt the originally-declared bounds.
	computedPos geom.Point

	style       *Style
	parentStyle *Style
	layer       int

	// invalidate is how this node (or a goroutine holding a reference to
	// it) asks the renderer for a redraw. nil until something wires it
	// up -- Canvas.AddShape / Container.AddChild / Bordered's inner
	// content all propagate it down automatically, same as parentStyle.
	invalidate func()
}

func newBaseNode(bounds *geom.Bounds, anchor Anchor, style Style, layer int) (BaseNode, error) {
	n := BaseNode{
		bounds:      bounds,
		anchor:      anchor,
		computedPos: *bounds.Pos, // copy, deliberately not the same pointer
		style:       ResolveStyle(style, *NewTransparentStyle()),
	}

	if err := n.SetLayer(layer); err != nil {
		return BaseNode{}, err
	}

	return n, nil
}

func (n *BaseNode) SetLayer(l int) error {
	if l < 0 {
		return errors.New("layers must be >= 0")
	}
	n.layer = l
	return nil
}

func (n *BaseNode) GetLayer() int {
	return n.layer
}

func (n *BaseNode) SetParentStyle(s *Style) {
	n.parentStyle = s
}

// Style returns this node's fully resolved style. Never dereferences a
// nil parentStyle -- a node drawn before being attached to a parent just
// resolves against fully transparent instead of panicking.
func (n *BaseNode) Style() Style {
	parent := Style{Bg: core.Transparent, Fg: core.Transparent}
	if n.parentStyle != nil {
		parent = *n.parentStyle
	}
	return *ResolveStyle(*n.style, parent)
}

// ResolvedStyle hands this node's resolved style down as a child's
// parentStyle, so style inheritance chains through every level instead
// of skipping the container's own style.
func (n *BaseNode) ResolvedStyle() *Style {
	s := n.Style()
	return &s
}

// SetInvalidator wires the redraw callback. Safe to call with nil (that's
// the default state -- Invalidate() just becomes a no-op).
func (n *BaseNode) SetInvalidator(fn func()) {
	n.invalidate = fn
}

// Invalidate asks the renderer for a redraw. Safe to call from any
// goroutine, at any time, including before this node is attached to
// anything -- it's a no-op until SetInvalidator has been wired up by a
// parent Container/Bordered or the root Canvas. This is the hook for
// self-refreshing components: hold a reference to a node, mutate its
// own state under its own lock, then call Invalidate().
func (n *BaseNode) Invalidate() {
	if n.invalidate != nil {
		n.invalidate()
	}
}

// IsInBounds checks the shape's declared local bounds against a parent
// frame. Callers must pass a parent bounds in the same local coordinate
// space the shape's own bounds.Pos is expressed in -- see LocalFrame,
// which every container uses for exactly this purpose so a child's
// position is always checked against its immediate container's interior,
// never against some ancestor's unrelated coordinate space.
func (n *BaseNode) IsInBounds(parent geom.Bounds) bool {
	if n.bounds.Pos.X < 0 || n.bounds.Pos.Y < 0 {
		return false
	}
	if n.bounds.Pos.X+n.bounds.W > parent.W {
		return false
	}
	if n.bounds.Pos.Y+n.bounds.H > parent.H {
		return false
	}
	return true
}

func (n *BaseNode) Layout(parent geom.Bounds) {
	n.computedPos.X = resolveAxis(n.anchor.H, parent.W, n.bounds.W, n.bounds.Pos.X)
	n.computedPos.Y = resolveAxis(n.anchor.V, parent.H, n.bounds.H, n.bounds.Pos.Y)
}

// LocalFrame is the zero-origin bounds children should be laid out and
// bounds-checked against: this node's own interior size, not its
// position within its own parent.
func (n *BaseNode) LocalFrame() geom.Bounds {
	return geom.Bounds{Pos: &geom.Point{}, W: n.bounds.W, H: n.bounds.H}
}

// Size reports this node's declared width and height. Used by
// LayoutPolicy implementations that need to arrange children without
// knowing their concrete type.
func (n *BaseNode) Size() (int, int) {
	return n.bounds.W, n.bounds.H
}

// SetComputedPos places this node directly, bypassing anchor resolution.
// Used by layout policies (e.g. StackLayout) that need outright control
// over position rather than resolving it from an anchor.
func (n *BaseNode) SetComputedPos(x, y int) {
	n.computedPos.X = x
	n.computedPos.Y = y
}

// AnchorH exposes the horizontal anchor so a LayoutPolicy can still honor
// it (e.g. centering a child horizontally) even when it's overriding Y
// outright.
func (n *BaseNode) AnchorH() AxisAnchor {
	return n.anchor.H
}