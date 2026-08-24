package framework

import "github.com/dmsRosa6/glyph/core"

type AppContext struct {
	Logs       chan<- core.AppLog
	Invalidate func()
	Focus      Navigator
	Signal     func(core.AppSignal)
	Done       <-chan struct{}
	Registry   *Registry
}

func (c AppContext) Log(l core.AppLog) {
	if c.Logs != nil {
		c.Logs <- l
	}
}

func (c AppContext) Lifecycle() <-chan struct{} {
	return c.Done
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

// Nodes returns the shared, app-wide node registry so any widget --
// most usefully from inside an AppActionFunc or FocusableActionFunc --
// can look up another widget elsewhere in the tree by ID. Safe to call
// and safe to chain straight off of even when Registry was never set;
// every Registry method is nil-receiver-safe.
func (c AppContext) Nodes() *Registry {
	return c.Registry
}

type noopNavigator struct{}

func (noopNavigator) Next()              {}
func (noopNavigator) Prev()              {}
func (noopNavigator) Enter() bool        { return false }
func (noopNavigator) Exit()              {}
func (noopNavigator) Current() Focusable { return nil }
