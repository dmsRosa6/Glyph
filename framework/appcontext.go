package framework

import "github.com/dmsRosa6/glyph/core"

type AppContext struct {
	Logs        chan<- core.AppLog
	Invalidate  func()
	Focus       Navigator
	Signal      func(core.AppSignal)
	Registry    *Registry
	Done        <-chan struct{}
	IsGlobalKey func(Key) bool
}

func (c AppContext) Log(l core.AppLog) {
	if c.Logs != nil {
		c.Logs <- l
	}
}

func (c AppContext) Redraw() {
	if c.Invalidate != nil {
		c.Invalidate()
	}
}

func (c AppContext) SignalApp(sig core.AppSignal) {
	if c.Signal != nil {
		c.Signal(sig)
	}
}

func (c AppContext) Nav() Navigator {
	if c.Focus != nil {
		return c.Focus
	}
	return noopNavigator{}
}

func (c AppContext) Nodes() *Registry {
	return c.Registry
}

// Lifecycle returns a channel that closes when the app is stopping.
// Nil-safe: selecting on a nil channel blocks forever, the correct
// "no lifecycle to hook into" behavior for a widget driven outside a
// running App.
func (c AppContext) Lifecycle() <-chan struct{} {
	return c.Done
}

// GlobalKeyBound reports whether k currently has a global App-level
// binding. Used by Propagator's shadow-warning check.
func (c AppContext) GlobalKeyBound(k Key) bool {
	if c.IsGlobalKey == nil {
		return false
	}
	return c.IsGlobalKey(k)
}

type noopNavigator struct{}

func (noopNavigator) Next()              {}
func (noopNavigator) Prev()              {}
func (noopNavigator) Enter() bool        { return false }
func (noopNavigator) Exit()              {}
func (noopNavigator) Current() Focusable { return nil }
