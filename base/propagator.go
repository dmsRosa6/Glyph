package base

import (
	"reflect"

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

	if p.parentStyle != nil {
		child.SetParentStyle(p.parentStyle)
	}
	if p.ctxSet {
		child.SetContext(p.ctx)
	}
}

func (p *Propagator) Untrack(target framework.Drawable) {
	for i, c := range p.owned {
		if c == target {
			p.owned = append(p.owned[:i], p.owned[i+1:]...)
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
	}
}

func (p *Propagator) PropagateLayer(l int) error {
	for _, c := range p.owned {
		if err := c.SetLayer(l); err != nil {
			return err
		}
	}
	return nil
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
