package base

import (
	"fmt"
	"reflect"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
)

type Propagator struct {
	owned []framework.Drawable

	parentStyle *framework.Style
	ctx         framework.AppContext
	ctxSet      bool
}

func (p *Propagator) Track(child framework.Drawable) {
	if isNilDrawable(child) {
		return
	}
	p.owned = append(p.owned, child)

	if r, ok := child.(framework.Raisable); ok {
		r.SetRaiser(func() { p.BringToFront(child) })
	}

	if p.parentStyle != nil {
		child.SetParentStyle(p.parentStyle)
	}
	if p.ctxSet {
		child.SetContext(p.ctx)
		registerChild(p.ctx, child)
		warnShadowedKeys(p.ctx, child)
	}
}

func (p *Propagator) Untrack(target framework.Drawable) {
	for i, c := range p.owned {
		if c == target {
			p.owned = append(p.owned[:i], p.owned[i+1:]...)
			if r, ok := target.(framework.Raisable); ok {
				r.SetRaiser(nil)
			}
			unregisterChild(p.ctx, target)
			return
		}
	}
}

func (p *Propagator) Children() []framework.Drawable {
	return p.owned
}

func (p *Propagator) PropagateStyle(s *framework.Style) {
	p.parentStyle = s
	for _, c := range p.owned {
		c.SetParentStyle(s)
	}
}

func (p *Propagator) PropagateContext(ctx framework.AppContext) {
	p.ctx = ctx
	p.ctxSet = true
	for _, c := range p.owned {
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
	if !p.owns(child) {
		return
	}
	max := 0
	for _, c := range p.owned {
		if c == child {
			continue
		}
		if l := c.GetLayer(); l > max {
			max = l
		}
	}
	_ = child.SetLayer(max + 1)
	p.invalidate()
}

// SendToBack is BringToFront's mirror, floored at 0 (SetLayer rejects
// negative layers).
func (p *Propagator) SendToBack(child framework.Drawable) {
	if !p.owns(child) {
		return
	}
	min := child.GetLayer()
	for _, c := range p.owned {
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
	p.invalidate()
}

func (p *Propagator) owns(target framework.Drawable) bool {
	for _, c := range p.owned {
		if c == target {
			return true
		}
	}
	return false
}

func (p *Propagator) invalidate() {
	if p.ctxSet {
		p.ctx.Redraw()
	}
}

// keyLister is satisfied by base.FocusableBaseNode (via BoundKeys).
type keyLister interface {
	BoundKeys() []framework.Key
}

// warnShadowedKeys catches the common case: a widget bound to a
// structural key (Ctrl+C/Enter/Tab/Esc) that's ALSO bound globally will
// never fire, since structural keys always dispatch global-first. Only
// checked at the two moments context first reaches a child -- a global
// binding added later isn't retroactively checked.
func warnShadowedKeys(ctx framework.AppContext, child framework.Drawable) {
	kl, ok := child.(keyLister)
	if !ok {
		return
	}
	for _, k := range kl.BoundKeys() {
		if !framework.IsStructuralKey(k) {
			continue // non-structural keys are widget-first now, can't be shadowed
		}
		if !ctx.GlobalKeyBound(k) {
			continue
		}
		id := ""
		if ident, ok := child.(framework.Identifiable); ok {
			id = ident.ID()
		}
		ctx.Log(*core.NewWarningAppLog(
			fmt.Errorf("key %q is a structural key bound both globally and on this widget -- the global binding always runs first, so this widget's action for %q will never fire", k.String(), k.String()),
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
