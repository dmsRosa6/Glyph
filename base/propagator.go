package base

import (
	"reflect"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
)

// Propagator centralizes the "fan this out to every sub-drawable I own"
// pattern that every composite widget (Bordered, Window, FocusableBox,
// Container, ...) was independently hand-rolling. A composite embeds
// Propagator alongside BaseNode, registers each sub-drawable it owns via
// Track, and then its SetParentStyle/SetInvalidator/SetLayer/
// SetLogChannel overrides shrink to one line each that calls the matching
// Propagate* method plus the BaseNode call for the composite's own state.
//
// This makes "I added a new sub-drawable field and forgot to wire one of
// the four propagation methods to it" impossible by construction: there's
// only one place (Track) that needs to be updated, not four.
//
// Track also mirrors Container.AddChild's eager-wiring behavior: if the
// composite has already been wired up (its parent already called
// SetParentStyle/SetInvalidator/SetLogChannel on it before this child was
// tracked), the newly tracked child is immediately wired with that same
// state instead of silently missing it until the next unrelated
// Propagate* call. That's what the FocusableBox bug this replaces was
// running into -- a child added after construction had no path to ever
// receiving an invalidator.
type Propagator struct {
	owned []framework.Drawable

	// Cached so Track can eagerly apply already-known parent state to a
	// child registered after the composite itself was wired up, the same
	// way Container.AddChild eagerly wires new children today.
	parentStyle *framework.Style
	invalidate  func()
	logs        chan<- core.AppLog
}

// Track registers child as owned by this composite node. Safe to call
// with a nil Drawable (including a typed-nil pointer stored in the
// interface, e.g. an optional *Text field that wasn't configured) -- it's
// silently ignored so callers don't need to guard optional sub-drawables
// before calling Track.
//
// If this Propagator already knows its parent style, invalidator, or log
// channel (because PropagateStyle/PropagateInvalidator/
// PropagateLogChannel already ran at least once), that state is applied
// to child immediately.
func (p *Propagator) Track(child framework.Drawable) {
	if isNilDrawable(child) {
		return
	}

	p.owned = append(p.owned, child)

	if p.parentStyle != nil {
		child.SetParentStyle(p.parentStyle)
	}
	if p.invalidate != nil {
		child.SetInvalidator(p.invalidate)
	}
	if p.logs != nil {
		child.SetLogChannel(p.logs)
	}
}

// Untrack removes a previously tracked sub-drawable, e.g. when a
// composite widget swaps one child out for another at runtime. A no-op
// if target was never tracked.
func (p *Propagator) Untrack(target framework.Drawable) {
	for i, c := range p.owned {
		if c == target {
			p.owned = append(p.owned[:i], p.owned[i+1:]...)
			return
		}
	}
}

// Children exposes the tracked sub-drawables read-only, mirroring
// Container.Children / framework.ChildrenLister, so a composite widget
// built on Propagator is automatically walkable by things like
// Canvas.CollectFocusable without needing its own Children() written by
// hand.
func (p *Propagator) Children() []framework.Drawable {
	return p.owned
}

// PropagateStyle fans a resolved parent style out to every tracked
// sub-drawable and remembers it so any child Tracked afterward gets it
// immediately too. Call this from the composite's own SetParentStyle
// override, after also calling BaseNode.SetParentStyle so the composite's
// own Style() resolves correctly.
func (p *Propagator) PropagateStyle(s *framework.Style) {
	p.parentStyle = s
	for _, c := range p.owned {
		c.SetParentStyle(s)
	}
}

// PropagateInvalidator fans the redraw callback out to every tracked
// sub-drawable and remembers it the same way PropagateStyle does. Call
// this from the composite's own SetInvalidator override, after also
// calling BaseNode.SetInvalidator.
func (p *Propagator) PropagateInvalidator(fn func()) {
	p.invalidate = fn
	for _, c := range p.owned {
		c.SetInvalidator(fn)
	}
}

// PropagateLogChannel fans the log channel out to every tracked
// sub-drawable and remembers it. Call this from the composite's own
// SetLogChannel override, after also calling BaseNode.SetLogChannel.
func (p *Propagator) PropagateLogChannel(ch chan<- core.AppLog) {
	p.logs = ch
	for _, c := range p.owned {
		c.SetLogChannel(ch)
	}
}

// PropagateLayer fans a layer value out to every tracked sub-drawable's
// own SetLayer. Unlike style/invalidator/logs this is NOT cached against
// future Track calls -- layer is a one-shot value passed down at the
// moment the composite's own SetLayer is called, not standing state a new
// child should inherit automatically, so a widget wanting that needs to
// re-call this explicitly (or set the layer before Track-ing children
// composed at a fixed relative layer from the start). Stops at the first
// error, matching BaseNode.SetLayer's validation contract (layers must be
// >= 0), and leaves any children before the failing one already updated --
// same non-atomic-on-error contract Container.AddChild already has.
func (p *Propagator) PropagateLayer(l int) error {
	for _, c := range p.owned {
		if err := c.SetLayer(l); err != nil {
			return err
		}
	}
	return nil
}

// isNilDrawable catches the classic Go footgun where a nil *Text (or any
// other nil pointer) stored in a framework.Drawable interface value is
// NOT itself == nil -- the interface has a concrete type, just a nil
// value. Track relies on this so callers can pass an optional,
// unconfigured sub-drawable field straight through without an "if x !=
// nil" guard at every call site.
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
