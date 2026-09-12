package mixin

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
)

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
	if s, ok := target.(framework.Stoppable); ok {
		s.Stop()
	}
	unregisterChild(ctx, target)
	return true
}

func (p *Propagator) Children() []framework.Drawable {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]framework.Drawable, len(p.owned))
	copy(out, p.owned)
	return out
}

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

func ownsChild(owned []framework.Drawable, target framework.Drawable) bool {
	for _, c := range owned {
		if c == target {
			return true
		}
	}
	return false
}

type keyLister interface {
	BoundKeys() []framework.Binding
}

func warnShadowedKeys(ctx framework.AppContext, child framework.Drawable) {
	kl, ok := child.(keyLister)
	if !ok {
		return
	}
	for _, b := range kl.BoundKeys() {
		if !framework.IsStructuralKey(b.Key) {
			continue
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
