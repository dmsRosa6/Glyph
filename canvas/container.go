package canvas

import (
	"fmt"
	"sort"
	"sync"

	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Container struct {
	base.BaseNode
	base.Propagator
	layout framework.LayoutPolicy

	// fitMu guards fitDirty and warned together, and is a separate lock
	// from base.Propagator's -- Draw reads fitDirty/warned from the
	// renderer goroutine while AddChild/RemoveChild/Resize write them
	// from whatever goroutine an input handler runs on.
	//
	// The out-of-bounds/doesn't-fit/doesn't-satisfy-layout-policy checks
	// this guards -- including CapacityAwareLayout.Fits's O(n) sum over
	// a freshly allocated "others" slice per child -- used to run in
	// full every single Draw call, for every not-yet-warned child. In
	// steady state (nothing actually overflowing) that's every child,
	// every frame: O(n^2) allocation work for a result that's static
	// until an add, a remove, or a resize of THIS container. fitDirty
	// starts true so the first Draw always computes; after that it's
	// only re-armed by AddChild/RemoveChild/Resize.
	//
	// Known gap: a child resized in place after being added does NOT
	// currently mark this dirty -- base.BaseNode.Resize has no hook
	// back to its parent Container. Not exercised anywhere in this
	// codebase today; would need a real Composable-level
	// resize-notification design, not a patch here.
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
		fitDirty: true,
		warned:   make(map[framework.Drawable]struct{}),
	}, nil
}

func (c *Container) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := c.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	frame := c.LocalFrame()

	// Children() hands back a fresh copy every call (see
	// base.Propagator.Children) -- safe to sort in place here without
	// corrupting insertion order for anyone else reading the tree
	// concurrently.
	children := c.Propagator.Children()

	sort.SliceStable(children, func(i, j int) bool {
		return children[i].GetLayer() < children[j].GetLayer()
	})

	// Positions every child, every frame -- cheap (O(n), no per-child
	// allocation) and has to reflect the current frame/anchor
	// regardless of whether anything structural changed.
	skipped := c.layout.Arrange(children, frame)

	// Diagnostics only, from here down -- none of this affects what
	// gets drawn, only whether a warning is logged. See fitDirty's
	// comment above for why this is gated.
	c.refreshFitWarnings(children, frame, skipped)

	buf.PushClip(v.X, v.Y, frame.W, frame.H)
	defer buf.PopClip()

	for _, child := range children {
		child.Draw(buf, v)
	}
}

// refreshFitWarnings recomputes the offending-child set from scratch
// and logs anything newly offending, but only when fitDirty is set --
// on a clean frame this returns immediately, before the O(n^2) part
// (the CapacityAwareLayout.Fits loop) ever runs.
//
// Recomputing fully on every dirty pass, rather than only ever adding
// to warned, is deliberate: it's what lets a resize that fixes an
// overflow correctly clear that child's warning, and lets it warn
// again if a later resize reintroduces the same problem -- the
// original always-append version couldn't do either.
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
			continue // standing condition, already logged, still true
		}
		c.Warn(f.err)
	}
}

func (c *Container) AddChild(child framework.Drawable) {
	// No checks here -- see refreshFitWarnings. AddChild always tracks.
	c.Propagator.Track(child)
	c.markFitDirty()
}

func (c *Container) RemoveChild(target framework.Drawable) {
	if !c.Propagator.Untrack(target) {
		return
	}
	c.fitMu.Lock()
	delete(c.warned, target) // avoid unbounded growth across add/remove churn
	c.fitMu.Unlock()
	c.markFitDirty()
	c.Logger().Debug(fmt.Sprintf("child removed, now %d children", c.Propagator.Count()))
}

// Resize shadows the promoted base.BaseNode.Resize purely to mark the
// fit-check cache dirty -- a container resize (e.g. terminal resize
// reaching the root via Canvas.ApplySize) is exactly the kind of
// structural change that can flip a child from fitting to not.
func (c *Container) Resize(w, h int) {
	c.BaseNode.Resize(w, h)
	c.markFitDirty()
}

func (c *Container) markFitDirty() {
	c.fitMu.Lock()
	c.fitDirty = true
	c.fitMu.Unlock()
}

func (c *Container) SetParentStyle(s *framework.Style) {
	c.BaseNode.SetParentStyle(s)
	c.Propagator.PropagateStyle(c.ResolvedStyle())
}

func (c *Container) SetContext(ctx framework.AppContext) {
	c.BaseNode.SetContext(ctx)
	c.Propagator.PropagateContext(ctx)
}
