package canvas

import (
	"errors"
	"fmt"
	"sort"

	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Container struct {
	base.BaseNode
	base.Propagator
	layout framework.LayoutPolicy
}

type ContainerConfig struct {
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
	Layout framework.LayoutPolicy // defaults to FreeLayout if nil
}

func NewContainer(bounds *geom.Bounds, cfg ContainerConfig) (*Container, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	policy := cfg.Layout
	if policy == nil {
		policy = framework.FreeLayout{}
	}

	return &Container{
		BaseNode: bn,
		layout:   policy,
	}, nil
}

func (c *Container) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := c.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	frame := c.LocalFrame()

	// Propagator.Children() returns its backing slice directly (not a
	// copy), so sorting it in place here reorders the same storage
	// Propagate* and Untrack operate on -- same contract Container had
	// with its own []framework.Drawable field before.
	children := c.Propagator.Children()

	sort.SliceStable(children, func(i, j int) bool {
		return children[i].GetLayer() < children[j].GetLayer()
	})

	c.layout.Arrange(children, frame)

	for _, child := range children {
		child.Draw(buf, v)
	}
}

func (c *Container) AddChild(child framework.Drawable) {
	if !child.IsInBounds(c.LocalFrame()) {
		c.Fault("Container", errors.New("shape out of container bounds"))
		return
	}

	// Track appends child and, since PropagateStyle/PropagateInvalidator/
	// PropagateLogChannel have already run at least once whenever this
	// container itself is attached to something, eagerly applies that
	// already-known state -- same eager-wire-on-add behavior AddChild had
	// before, just funneled through one call instead of three.
	c.Propagator.Track(child)
}

func (c *Container) RemoveChild(target framework.Drawable) {
	before := len(c.Propagator.Children())
	c.Propagator.Untrack(target)
	if len(c.Propagator.Children()) < before {
		c.Logger("Container").Debug(fmt.Sprintf("child removed, now %d children", len(c.Propagator.Children())))
	}
}

func (c *Container) SetParentStyle(s *framework.Style) {
	c.BaseNode.SetParentStyle(s)
	// Deliberately c.ResolvedStyle(), not the raw incoming s: children
	// inherit THIS container's fully resolved style, not its parent's,
	// so a Transparent field set at this level still chains correctly
	// instead of skipping a level of inheritance.
	c.Propagator.PropagateStyle(c.ResolvedStyle())
}

func (c *Container) SetInvalidator(fn func()) {
	c.BaseNode.SetInvalidator(fn)
	c.Propagator.PropagateInvalidator(fn)
}

func (c *Container) SetLogChannel(ch chan<- core.AppLog) {
	c.BaseNode.SetLogChannel(ch)
	c.Propagator.PropagateLogChannel(ch)
}
