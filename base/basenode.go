package base

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type BaseNode struct {
	bounds *geom.Bounds
	anchor framework.Anchor

	computedPos geom.Point

	style       *framework.Style
	parentStyle *framework.Style
	layer       int

	ctx    framework.AppContext
	source string // set once at construction, read by every Logger/Warn/Fault call -- never passed around again
}

func NewBaseNode(bounds *geom.Bounds, anchor framework.Anchor, style framework.Style, layer int, source string) (BaseNode, error) {
	n := BaseNode{
		bounds:      bounds,
		anchor:      anchor,
		computedPos: *bounds.Pos,
		style:       framework.ResolveStyle(style, *framework.NewTransparentStyle()),
		source:      source,
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

func (n *BaseNode) SetParentStyle(s *framework.Style) {
	n.parentStyle = s
}

func (n *BaseNode) Style() framework.Style {
	parent := framework.Style{Bg: core.Transparent, Fg: core.Transparent}
	if n.parentStyle != nil {
		parent = *n.parentStyle
	}
	return *framework.ResolveStyle(*n.style, parent)
}

func (n *BaseNode) ResolvedStyle() *framework.Style {
	s := n.Style()
	return &s
}

func (n *BaseNode) SetContext(ctx framework.AppContext) {
	n.ctx = ctx
}

func (n *BaseNode) Context() framework.AppContext {
	return n.ctx
}

func (n *BaseNode) Invalidate() {
	n.ctx.Redraw()
}

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
	n.computedPos.X = framework.ResolveAxis(n.anchor.H, parent.W, n.bounds.W, n.bounds.Pos.X)
	n.computedPos.Y = framework.ResolveAxis(n.anchor.V, parent.H, n.bounds.H, n.bounds.Pos.Y)
}

func (n *BaseNode) LocalFrame() geom.Bounds {
	return geom.Bounds{Pos: &geom.Point{}, W: n.bounds.W, H: n.bounds.H}
}

func (n *BaseNode) Size() (int, int) {
	return n.bounds.W, n.bounds.H
}

func (n *BaseNode) SetComputedPos(x, y int) {
	n.computedPos.X = x
	n.computedPos.Y = y
}

func (n *BaseNode) AnchorH() framework.AxisAnchor {
	return n.anchor.H
}

func (n *BaseNode) ComputedPos() geom.Point {
	return n.computedPos
}

func (n *BaseNode) Bounds() *geom.Bounds {
	return n.bounds
}

func (n *BaseNode) Resize(w, h int) {
	n.bounds.W = w
	n.bounds.H = h
}

// Logger, Fault, and Warn no longer take a source string -- it's fixed
// at construction (see NewBaseNode) so every log line from this node
// carries the same source without it being retyped at each call site.

func (n *BaseNode) Logger() framework.Logger {
	return framework.NewLogger(n.ctx.Logs, n.source)
}

func (n *BaseNode) Fault(err error) {
	if n.ctx.Logs == nil {
		panic(err)
	}
	n.Logger().Fatal(err)
}

func (n *BaseNode) Warn(err error) {
	n.Logger().Warning(err)
}