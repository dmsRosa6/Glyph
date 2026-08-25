package canvas

import (
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

	// warned dedupes out-of-bounds/doesn't-fit warnings per child so a
	// standing condition (a window that's just always too small for its
	// content) logs once, not once per frame.
	warned map[framework.Drawable]struct{}
}

type ContainerConfig struct {
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
	Layout framework.LayoutPolicy
}

func NewContainer(bounds *geom.Bounds, cfg ContainerConfig) (*Container, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Container")
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
		warned:   make(map[framework.Drawable]struct{}),
	}, nil
}

func (c *Container) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := c.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	frame := c.LocalFrame()

	children := c.Propagator.Children()

	sort.SliceStable(children, func(i, j int) bool {
		return children[i].GetLayer() < children[j].GetLayer()
	})

	skipped := c.layout.Arrange(children, frame)
	for _, s := range skipped {
		c.Warn(fmt.Errorf("child of type %T does not satisfy this container's layout policy and was not positioned", s))
	}

	for i, child := range children {
		if _, already := c.warned[child]; already {
			continue
		}

		if !child.IsInBounds(frame) {
			c.Warn(fmt.Errorf("child of type %T is out of container bounds and will be visually clipped", child))
			c.warned[child] = struct{}{}
			continue
		}

		if cal, ok := c.layout.(framework.CapacityAwareLayout); ok {
			others := make([]framework.Drawable, 0, len(children)-1)
			others = append(others, children[:i]...)
			others = append(others, children[i+1:]...)
			if !cal.Fits(others, child, frame) {
				c.Warn(fmt.Errorf("child of type %T does not fit in container's remaining space and will be visually clipped", child))
				c.warned[child] = struct{}{}
			}
		}
	}

	buf.PushClip(v.X, v.Y, frame.W, frame.H)
	defer buf.PopClip()

	for _, child := range children {
		child.Draw(buf, v)
	}
}

func (c *Container) AddChild(child framework.Drawable) {
	// No checks here anymore -- see Draw. AddChild always tracks.
	c.Propagator.Track(child)
}

func (c *Container) RemoveChild(target framework.Drawable) {
	before := len(c.Propagator.Children())
	c.Propagator.Untrack(target)
	if len(c.Propagator.Children()) < before {
		delete(c.warned, target) // avoid unbounded growth across add/remove churn
		c.Logger().Debug(fmt.Sprintf("child removed, now %d children", len(c.Propagator.Children())))
	}
}

func (c *Container) SetParentStyle(s *framework.Style) {
	c.BaseNode.SetParentStyle(s)
	c.Propagator.PropagateStyle(c.ResolvedStyle())
}

func (c *Container) SetContext(ctx framework.AppContext) {
	c.BaseNode.SetContext(ctx)
	c.Propagator.PropagateContext(ctx)
}
