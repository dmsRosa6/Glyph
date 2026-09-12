package mixin

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

var idSeq uint64

func nextID(source string) string {
	return fmt.Sprintf("%s#%d", source, atomic.AddUint64(&idSeq, 1))
}

type Node struct {
	bounds *geom.Bounds
	anchor framework.Anchor

	computedPos geom.Point

	style       *framework.Style
	parentStyle *framework.Style

	layer int64

	ctx    framework.AppContext
	source string
	id     string
}

func NewNode(bounds *geom.Bounds, anchor framework.Anchor, style framework.Style, layer int, source string) (Node, error) {
	if !bounds.Valid() {
		return Node{}, errors.New("bounds must have non-negative position, width, and height")
	}

	n := Node{
		bounds:      bounds,
		anchor:      anchor,
		computedPos: *bounds.Pos,
		style:       framework.ResolveStyle(style, *framework.NewTransparentStyle()),
		source:      source,
		id:          nextID(source),
	}

	if err := n.SetLayer(layer); err != nil {
		return Node{}, err
	}

	return n, nil
}

func (n *Node) SetLayer(l int) error {
	if l < 0 {
		return errors.New("layers must be >= 0")
	}
	atomic.StoreInt64(&n.layer, int64(l))
	return nil
}

func (n *Node) GetLayer() int {
	return int(atomic.LoadInt64(&n.layer))
}

func (n *Node) SetParentStyle(s *framework.Style) {
	n.parentStyle = s
}

func (n *Node) Style() framework.Style {
	parent := framework.Style{Bg: core.Transparent, Fg: core.Transparent}
	if n.parentStyle != nil {
		parent = *n.parentStyle
	}
	return *framework.ResolveStyle(*n.style, parent)
}

func (n *Node) ResolvedStyle() *framework.Style {
	s := n.Style()
	return &s
}

func (n *Node) SetContext(ctx framework.AppContext) {
	n.ctx = ctx
}

func (n *Node) Context() framework.AppContext {
	return n.ctx
}

func (n *Node) Invalidate() {
	n.ctx.Redraw()
}

func (n *Node) IsInBounds(parent geom.Bounds) bool {
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

func (n *Node) Layout(parent geom.Bounds) {
	n.computedPos.X = framework.ResolveAxis(n.anchor.H, parent.W, n.bounds.W, n.bounds.Pos.X)
	n.computedPos.Y = framework.ResolveAxis(n.anchor.V, parent.H, n.bounds.H, n.bounds.Pos.Y)
}

func (n *Node) LocalFrame() geom.Bounds {
	return geom.Bounds{Pos: &geom.Point{}, W: n.bounds.W, H: n.bounds.H}
}

func (n *Node) Size() (int, int) {
	return n.bounds.W, n.bounds.H
}

func (n *Node) SetComputedPos(x, y int) {
	n.computedPos.X = x
	n.computedPos.Y = y
}

func (n *Node) AnchorH() framework.AxisAnchor {
	return n.anchor.H
}

func (n *Node) ComputedPos() geom.Point {
	return n.computedPos
}

func (n *Node) Bounds() *geom.Bounds {
	return n.bounds
}

func (n *Node) Resize(w, h int) {
	n.bounds.W = w
	n.bounds.H = h
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) SetID(id string) {
	n.id = id
}

func (n *Node) Source() string {
	return n.source
}

func (n *Node) Logger() framework.Logger {
	return framework.NewLogger(n.ctx.Logs, n.ctx.LogLevel, n.source, n.id)
}

func (n *Node) Fault(err error) {
	if n.ctx.Logs == nil {
		panic(err)
	}
	n.Logger().Fatal(err)
}

func (n *Node) Warn(err error) {
	n.Logger().Warning(err)
}
