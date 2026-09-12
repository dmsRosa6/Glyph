package framework

import "github.com/dmsRosa6/glyph/core"

type AppContext struct {
	Logs        chan<- core.AppLog
	Invalidate  func()
	Focus       Navigator
	Signal      func(core.AppSignal)
	Registry    *Registry
	Done        <-chan struct{}
	LogLevel    core.Severity
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

func (c AppContext) Lifecycle() <-chan struct{} {
	return c.Done
}

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
