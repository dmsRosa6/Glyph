package canvas

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

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

	// scrollY is this container's vertical scroll offset in cells,
	// applied ONLY to where children get drawn -- see ScrollY's own
	// doc comment for why that's a different thing from moving the
	// container's own on-screen position. Zero by default, so every
	// existing container is completely unaffected unless something
	// explicitly calls SetScrollY. Read every Draw call on the render
	// goroutine, written from whatever goroutine an input handler runs
	// on -- same concurrent-access shape as mixin.Node's own `layer`
	// field, given the identical plain-atomic-int64 treatment for the
	// identical reason (a full RWMutex costs more than a single int
	// swap needs to, on a genuinely hot path).
	scrollY int64
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

	// drawVec is where each child's OWN ComputedPos gets added on top
	// of -- deliberately a SEPARATE vector from v above, which the
	// PushClip call already used: scrolling has to shift what content
	// lands inside this container's fixed on-screen rectangle without
	// moving that rectangle itself, or ScrollY would just be panning
	// the whole container around the screen instead of scrolling its
	// content within a fixed viewport.
	drawVec := geom.Vector{X: v.X, Y: v.Y - int(atomic.LoadInt64(&c.scrollY))}

	for _, child := range children {
		child.Draw(buf, drawVec)
	}
}

// ScrollY is this container's vertical scroll offset in cells. It is a
// mechanism, not a policy: Container has no opinion about whether, when,
// or how far to scroll -- widgets.List (the first, and so far only,
// consumer) is what decides that, by calling SetScrollY from its own
// auto-follow-focus logic. Nothing here clamps the value to any
// "valid" range either; a caller driving this is expected to do its
// own clamping against whatever it considers valid (List does, in
// setClampedScroll).
func (c *Container) ScrollY() int {
	return int(atomic.LoadInt64(&c.scrollY))
}

func (c *Container) SetScrollY(y int) {
	atomic.StoreInt64(&c.scrollY, int64(y))
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
