package mixin

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/dmsRosa6/glyph/framework"
)

// FocusBehavior is focus/input-handling behavior as a mixin -- it owns
// no Node, so a struct can embed both Node (or *canvas.Container) and
// FocusBehavior at the same depth with no method-set collision.
type FocusBehavior struct {
	// actionsMu is a *sync.RWMutex (not a value) so every copy of a
	// constructed FocusBehavior shares the same lock -- needed once
	// it's embedded by value into a composite.
	actionsMu  *sync.RWMutex
	actions    map[framework.Binding]FocusableActionFunc
	focused    bool
	focusStyle *framework.Style

	ctx    framework.AppContext
	source string
	id     string
}

// FocusableActionContext is what a bound action receives: the
// FocusBehavior it's bound to, and the triggering Event.
type FocusableActionContext struct {
	behavior *FocusBehavior
	ev       framework.Event
}

type FocusableActionFunc func(action FocusableActionContext) (bool, error)

func (a FocusableActionContext) Behavior() *FocusBehavior {
	return a.behavior
}

func (a FocusableActionContext) Event() framework.Event {
	return a.ev
}

func (a FocusableActionContext) Nodes() *framework.Registry {
	return a.behavior.ctx.Nodes()
}

func NewFocusBehavior(source string) FocusBehavior {
	return FocusBehavior{
		actionsMu: &sync.RWMutex{},
		actions:   make(map[framework.Binding]FocusableActionFunc),
		source:    source,
	}
}

// SetFocusContext wires this behavior's redraw/log hook. id is
// captured once, at the moment context first reaches this behavior.
func (f *FocusBehavior) SetFocusContext(ctx framework.AppContext, id string) {
	f.ctx = ctx
	f.id = id
}

func (f *FocusBehavior) SetFocusStyle(s framework.Style) {
	f.focusStyle = &s
}

// ResolveFocusStyle blends focusStyle over base when focused.
func (f *FocusBehavior) ResolveFocusStyle(base framework.Style) framework.Style {
	if f.focused && f.focusStyle != nil {
		return *framework.ResolveStyle(*f.focusStyle, base)
	}
	return base
}

// BindAction binds fn to k with no modifiers. Use BindActionMod for a
// specific modifier combination.
func (f *FocusBehavior) BindAction(k framework.Key, fn FocusableActionFunc) {
	f.BindActionMod(k, framework.ModNone, fn)
}

func (f *FocusBehavior) BindActionMod(k framework.Key, mods framework.Modifier, fn FocusableActionFunc) {
	f.actionsMu.Lock()
	defer f.actionsMu.Unlock()
	f.actions[framework.Binding{Key: k, Modifiers: mods}] = fn
}

// BoundKeys lists every binding this behavior has an action for. Used
// by Propagator's shadow-key warning.
func (f *FocusBehavior) BoundKeys() []framework.Binding {
	f.actionsMu.RLock()
	defer f.actionsMu.RUnlock()
	keys := make([]framework.Binding, 0, len(f.actions))
	for b := range f.actions {
		keys = append(keys, b)
	}
	return keys
}

func (f *FocusBehavior) HandleInput(ev framework.Event) (bool, error) {
	f.actionsMu.RLock()
	fn, ok := f.actions[framework.Binding{Key: ev.Key, Modifiers: ev.Modifiers}]
	f.actionsMu.RUnlock()
	if !ok {
		return false, nil
	}
	refresh, err := f.invoke(fn, ev)
	if err != nil {
		f.logger().Warning(err)
	}
	if refresh {
		f.ctx.Redraw()
	}
	return true, err
}

// invoke calls fn with a recover in place, since fn is caller-supplied
// and a panic in any goroutine kills the whole process. Recovered
// panics are logged as Fatal, which fault.FaultManager promotes to a
// clean SIGTERM shutdown instead of a broken terminal.
func (f *FocusBehavior) invoke(fn FocusableActionFunc, ev framework.Event) (refresh bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			f.logger().Fatal(fmt.Errorf("recovered panic in bound action for key %q: %v\n%s", ev.Key.String(), r, debug.Stack()))
			refresh, err = false, nil
		}
	}()
	return fn(FocusableActionContext{behavior: f, ev: ev})
}

func (f *FocusBehavior) Focus() {
	if f.focused {
		return
	}
	f.focused = true
	f.logger().Debug("focused")
	f.ctx.Redraw()
}

func (f *FocusBehavior) Blur() {
	if !f.focused {
		return
	}
	f.focused = false
	f.logger().Debug("blurred")
	f.ctx.Redraw()
}

func (f *FocusBehavior) IsFocused() bool {
	return f.focused
}

func (f *FocusBehavior) logger() framework.Logger {
	return framework.NewLogger(f.ctx.Logs, f.ctx.LogLevel, f.source, f.id)
}
