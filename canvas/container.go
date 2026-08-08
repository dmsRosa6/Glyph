package canvas

import (
	"sort"

	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Container struct {
	base.BaseNode
	children []framework.Drawable
	layout   framework.LayoutPolicy
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
		children: []framework.Drawable{},
		layout:   policy,
	}, nil
}

func (c *Container) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := c.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	frame := c.LocalFrame()

	sort.SliceStable(c.children, func(i, j int) bool {
		return c.children[i].GetLayer() < c.children[j].GetLayer()
	})

	c.layout.Arrange(c.children, frame)

	for _, child := range c.children {
		child.Draw(buf, v)
	}
}

func (c *Container) AddChild(child framework.Drawable) {
	if !child.IsInBounds(c.LocalFrame()) {
		panic("shape out of container bounds")
	}

	child.SetParentStyle(c.ResolvedStyle())
	child.SetInvalidator(c.Invalidate)

	c.children = append(c.children, child)
}

func (c *Container) RemoveChild(target framework.Drawable) {
	for i, child := range c.children {
		if child == target {
			c.children = append(c.children[:i], c.children[i+1:]...)
			return
		}
	}
}

func (c *Container) SetParentStyle(s *framework.Style) {
	c.BaseNode.SetParentStyle(s)
	for _, child := range c.children {
		child.SetParentStyle(c.ResolvedStyle())
	}
}

func (c *Container) SetInvalidator(fn func()) {
	c.BaseNode.SetInvalidator(fn)
	for _, child := range c.children {
		child.SetInvalidator(fn)
	}
}

// Children exposes this container's children for callers that can't
// import canvas directly (e.g. canvas.go's collectFocusable, once
// widgets like Bordered live in a separate package) -- see
// framework.ChildrenLister.
func (c *Container) Children() []framework.Drawable {
	return c.children
}