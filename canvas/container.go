package canvas

import (
	"fmt"
	"sort"
	"sync"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

type Container struct {
	mixin.Node
	mixin.Propagator
	layout framework.LayoutPolicy

	fitMu    sync.Mutex
	fitDirty bool
	warned   map[framework.Drawable]struct{}
}

type ContainerConfig struct {
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
	Layout framework.LayoutPolicy
}

func NewContainer(bounds *geom.Bounds, cfg ContainerConfig) (*Container, error) {
	bn, err := mixin.NewNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Container")
	if err != nil {
		return nil, err
	}

	policy := cfg.Layout
	if policy == nil {
		policy = framework.FreeLayout{}
	}

	return &Container{
		Node:     bn,
		layout:   policy,
		fitDirty: true,
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

	c.refreshFitWarnings(children, frame, skipped)

	buf.PushClip(v.X, v.Y, frame.W, frame.H)
	defer buf.PopClip()

	for _, child := range children {
		child.Draw(buf, v)
	}
}

func (c *Container) refreshFitWarnings(children []framework.Drawable, frame geom.Bounds, skipped []framework.Drawable) {
	c.fitMu.Lock()
	dirty := c.fitDirty
	c.fitMu.Unlock()
	if !dirty {
		return
	}

	type finding struct {
		child framework.Drawable
		err   error
	}
	var findings []finding
	offending := make(map[framework.Drawable]struct{}, len(children))

	for _, s := range skipped {
		offending[s] = struct{}{}
		findings = append(findings, finding{s, fmt.Errorf("child of type %T does not satisfy this container's layout policy and was not positioned", s)})
	}

	for i, child := range children {
		if _, already := offending[child]; already {
			continue
		}
		if !child.IsInBounds(frame) {
			offending[child] = struct{}{}
			findings = append(findings, finding{child, fmt.Errorf("child of type %T is out of container bounds and will be visually clipped", child)})
			continue
		}
		if cal, ok := c.layout.(framework.CapacityAwareLayout); ok {
			others := make([]framework.Drawable, 0, len(children)-1)
			others = append(others, children[:i]...)
			others = append(others, children[i+1:]...)
			if !cal.Fits(others, child, frame) {
				offending[child] = struct{}{}
				findings = append(findings, finding{child, fmt.Errorf("child of type %T does not fit in container's remaining space and will be visually clipped", child)})
			}
		}
	}

	c.fitMu.Lock()
	prev := c.warned
	c.warned = offending
	c.fitDirty = false
	c.fitMu.Unlock()

	for _, f := range findings {
		if _, already := prev[f.child]; already {
			continue
		}
		c.Warn(f.err)
	}
}

func (c *Container) AddChild(child framework.Drawable) {
	c.Propagator.Track(child)
	c.markFitDirty()
}

func (c *Container) RemoveChild(target framework.Drawable) {
	if !c.Propagator.Untrack(target) {
		return
	}
	c.fitMu.Lock()
	delete(c.warned, target)
	c.fitMu.Unlock()
	c.markFitDirty()
	c.Logger().Debug(fmt.Sprintf("child removed, now %d children", c.Propagator.Count()))
}

func (c *Container) Resize(w, h int) {
	c.Node.Resize(w, h)
	c.markFitDirty()
}

func (c *Container) markFitDirty() {
	c.fitMu.Lock()
	c.fitDirty = true
	c.fitMu.Unlock()
}

func (c *Container) SetParentStyle(s *framework.Style) {
	c.Node.SetParentStyle(s)
	c.Propagator.PropagateStyle(c.ResolvedStyle())
}

func (c *Container) SetContext(ctx framework.AppContext) {
	c.Node.SetContext(ctx)
	c.Propagator.PropagateContext(ctx)
}
