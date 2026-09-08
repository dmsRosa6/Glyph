package base

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
)

// Propagator is safe for concurrent use: Track/Untrack/BringToFront/
// SendToBack run from whatever goroutine owns input handling, while
// Container.Draw reads (and used to sort in place) the same data from
// the renderer's goroutine. mu guards owned, parentStyle, ctx, and
// ctxSet together -- they're read and mutated as a related group (see
// Track/PropagateContext), not independently.
type Propagator struct {
	mu    sync.RWMutex
	owned []framework.Drawable

	parentStyle *framework.Style
	ctx         framework.AppContext
	ctxSet      bool
}

func (p *Propagator) Track(child framework.Drawable) {
	if isNilDrawable(child) {
		return
	}

	// Mutate owned and snapshot the fields Track needs under the lock,
	// then release before calling out to child's own methods. Calling
	// interface methods on an unknown Drawable while holding this lock
	// would be a self-inflicted deadlock risk the moment any Drawable's
	// own method ever needed to call back into this Propagator.
	p.mu.Lock()
	p.owned = append(p.owned, child)
	parentStyle := p.parentStyle
	ctx, ctxSet := p.ctx, p.ctxSet
	p.mu.Unlock()

	if r, ok := child.(framework.Raisable); ok {
		r.SetRaiser(func() { p.BringToFront(child) })
	}

	if parentStyle != nil {
		child.SetParentStyle(parentStyle)
	}
	if ctxSet {
		child.SetContext(ctx)
		registerChild(ctx, child)
		warnShadowedKeys(ctx, child)
	}
}

// Untrack reports whether it actually removed something, so
// Container.RemoveChild doesn't need to call Children() (which always
// allocates, see below) just to diff a length.
func (p *Propagator) Untrack(target framework.Drawable) (removed bool) {
	p.mu.Lock()
	idx := -1
	for i, c := range p.owned {
		if c == target {
			idx = i
			break
		}
	}
	if idx == -1 {
		p.mu.Unlock()
		return false
	}
	p.owned = append(p.owned[:idx], p.owned[idx+1:]...)
	ctx := p.ctx
	p.mu.Unlock()

	if r, ok := target.(framework.Raisable); ok {
		r.SetRaiser(nil)
	}
	// Stop before unregisterChild -- unregisterChild also walks and
	// unregisters grandchildren via ChildrenLister, and a Stoppable
	// widget's own Draw/Invalidate calls should have already stopped
	// arriving by the time this node disappears from the Registry, not
	// after.
	if s, ok := target.(framework.Stoppable); ok {
		s.Stop()
	}
	unregisterChild(ctx, target)
	return true
}

// Children returns a defensive copy, not the live backing array.
// Previously this returned p.owned directly, and Container.Draw sorted
// that returned slice by layer every frame -- since it was the same
// backing array as p.owned, the sort silently and permanently
// overwrote insertion order as a side effect of rendering. Every
// caller now gets its own snapshot to read or reorder freely.
//
// Cost note: this is an O(n) allocation on every call, and
// Container.Draw calls it once per container, every single frame --
// even a frame where nothing in that container changed. Fine at the
// tree sizes this framework has actually been exercised with; worth
// knowing before building deep/wide trees, since it's the one
// per-frame allocation left in the render path after Buffer's
// flat-storage rewrite and Renderer.Flush's diffing (see core.Buffer
// and render.Renderer.Flush). Not changed here: avoiding it needs a
// real dirty-tracking scheme (invalidate a cached copy on Track/
// Untrack/BringToFront/SendToBack, correctly, under the same mu this
// method already takes) -- meaningfully more moving parts than this
// method has today, for a cost nothing has actually reported hitting
// yet. A defensive copy that's simple and provably correct beats a
// cache that's fast and subtly wrong.
func (p *Propagator) Children() []framework.Drawable {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]framework.Drawable, len(p.owned))
	copy(out, p.owned)
	return out
}

// Count is Children() without the allocation, for call sites that only
// want a size (e.g. a debug log line) and don't need the contents.
func (p *Propagator) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.owned)
}

func (p *Propagator) PropagateStyle(s *framework.Style) {
	p.mu.Lock()
	p.parentStyle = s
	children := append([]framework.Drawable(nil), p.owned...)
	p.mu.Unlock()

	for _, c := range children {
		c.SetParentStyle(s)
	}
}

func (p *Propagator) PropagateContext(ctx framework.AppContext) {
	p.mu.Lock()
	p.ctx = ctx
	p.ctxSet = true
	children := append([]framework.Drawable(nil), p.owned...)
	p.mu.Unlock()

	for _, c := range children {
		c.SetContext(ctx)
		registerChild(ctx, c)
		warnShadowedKeys(ctx, c)
	}
}

// BringToFront moves child to draw after all its current siblings in
// this Propagator -- visually on top, since Container.Draw sorts
// ascending by layer. Sets child's layer to one above the current max
// among its siblings. No-op if child isn't currently tracked here.
// Scope is this Propagator's own children only -- same as CSS z-index
// only ever comparing within its own stacking context.
func (p *Propagator) BringToFront(child framework.Drawable) {
	p.mu.RLock()
	owned := append([]framework.Drawable(nil), p.owned...)
	ctx, ctxSet := p.ctx, p.ctxSet
	p.mu.RUnlock()

	if !ownsChild(owned, child) {
		return
	}

	max := 0
	for _, c := range owned {
		if c == child {
			continue
		}
		if l := c.GetLayer(); l > max {
			max = l
		}
	}
	_ = child.SetLayer(max + 1)

	if ctxSet {
		ctx.Redraw()
	}
}

// SendToBack is BringToFront's mirror, floored at 0 (SetLayer rejects
// negative layers).
func (p *Propagator) SendToBack(child framework.Drawable) {
	p.mu.RLock()
	owned := append([]framework.Drawable(nil), p.owned...)
	ctx, ctxSet := p.ctx, p.ctxSet
	p.mu.RUnlock()

	if !ownsChild(owned, child) {
		return
	}

	min := child.GetLayer()
	for _, c := range owned {
		if c == child {
			continue
		}
		if l := c.GetLayer(); l < min {
			min = l
		}
	}
	if min <= 0 {
		_ = child.SetLayer(0)
	} else {
		_ = child.SetLayer(min - 1)
	}

	if ctxSet {
		ctx.Redraw()
	}
}

// ownsChild is a plain helper over an already-fetched snapshot, rather
// than a method that re-locks -- BringToFront/SendToBack already hold
// the one snapshot they need for both the membership check and the
// layer scan.
func ownsChild(owned []framework.Drawable, target framework.Drawable) bool {
	for _, c := range owned {
		if c == target {
			return true
		}
	}
	return false
}

// keyLister is satisfied by base.FocusableBaseNode (via BoundKeys).
type keyLister interface {
	BoundKeys() []framework.Binding
}

// warnShadowedKeys catches the common case: a widget bound to a
// structural key (Ctrl+C/Enter/Tab/Esc) that's ALSO bound globally
// under the SAME (key, modifiers) combination will never fire, since
// structural keys always dispatch global-first. Only checked at the two
// moments context first reaches a child -- a global binding added later
// isn't retroactively checked.
//
// The match is exact per framework.Binding: a widget binding
// Binding{KeyTab, ModShift} is NOT shadowed by an unrelated global
// Binding{KeyTab, ModNone} (plain Tab) -- they're different bindings
// that fire independently. Checking IsStructuralKey/GlobalKeyBound
// against the bare Key here would produce a false positive for exactly
// the Shift+Tab-vs-Tab case this whole Binding scheme exists to support.
func warnShadowedKeys(ctx framework.AppContext, child framework.Drawable) {
	kl, ok := child.(keyLister)
	if !ok {
		return
	}
	for _, b := range kl.BoundKeys() {
		if !framework.IsStructuralKey(b.Key) {
			continue // non-structural keys are widget-first now, can't be shadowed
		}
		if !ctx.GlobalKeyBound(b) {
			continue
		}
		id := ""
		if ident, ok := child.(framework.Identifiable); ok {
			id = ident.ID()
		}
		desc := b.Modifiers.String() + b.Key.String()
		ctx.Log(*core.NewWarningAppLog(
			fmt.Errorf("key %q is a structural key bound both globally and on this widget -- the global binding always runs first, so this widget's action for %q will never fire", desc, desc),
			"Propagator",
		).WithID(id))
	}
}

func registerChild(ctx framework.AppContext, child framework.Drawable) {
	id, ok := child.(framework.Identifiable)
	if !ok {
		return
	}
	if ctx.Nodes().Register(id.ID(), child) {
		ctx.Log(*core.NewWarningAppLog(
			fmt.Errorf("duplicate node id %q registered -- previous node under this id is now unreachable via Nodes().Find", id.ID()),
			"Propagator",
		).WithID(id.ID()))
	}
}

func unregisterChild(ctx framework.AppContext, child framework.Drawable) {
	if id, ok := child.(framework.Identifiable); ok {
		ctx.Nodes().Unregister(id.ID())
	}
	if cl, ok := child.(framework.ChildrenLister); ok {
		for _, gc := range cl.Children() {
			unregisterChild(ctx, gc)
		}
	}
}

func isNilDrawable(d framework.Drawable) bool {
	if d == nil {
		return true
	}
	v := reflect.ValueOf(d)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}
