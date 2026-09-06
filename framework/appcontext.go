package framework

import "github.com/dmsRosa6/glyph/core"

type AppContext struct {
	Logs       chan<- core.AppLog
	Invalidate func()
	Focus      Navigator
	Signal     func(core.AppSignal)
	Registry   *Registry
	Done       <-chan struct{}
	// LogLevel is the configured minimum severity (AppConfig.LogLevel),
	// threaded through so every framework.Logger built from this
	// context -- BaseNode.Logger(), FocusBehavior's own logger -- can
	// filter Debug/Info noise at the source instead of relying solely
	// on FaultManager's downstream check. See framework.Logger's doc
	// comment for why filtering here, not just downstream, matters.
	LogLevel core.Severity
	// IsGlobalKey takes a Binding, not a bare Key, now that global
	// bindings are keyed on (Key, Modifiers) -- a Key alone can't
	// distinguish "Tab is globally bound" from "Shift+Tab is globally
	// bound," and the shadow-key warning below needs that distinction
	// to avoid a false positive: a widget binding Shift+Tab on itself
	// is NOT shadowed by a global plain-Tab binding, since they're
	// different Bindings and both fire independently.
	IsGlobalKey func(Binding) bool
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

// GlobalKeyBound reports whether b currently has a global App-level
// binding. Used by Propagator's shadow-warning check.
func (c AppContext) GlobalKeyBound(b Binding) bool {
	if c.IsGlobalKey == nil {
		return false
	}
	return c.IsGlobalKey(b)
}

type noopNavigator struct{}

func (noopNavigator) Next()              {}
func (noopNavigator) Prev()              {}
func (noopNavigator) Enter() bool        { return false }
func (noopNavigator) Exit()              {}
func (noopNavigator) Current() Focusable { return nil }
