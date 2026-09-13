package mixin

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/dmsRosa6/glyph/framework"
)

type FocusBehavior struct {
	actionsMu  *sync.RWMutex
	actions    map[framework.Binding]FocusableActionFunc
	focused    bool
	focusStyle *framework.Style

	ctx    framework.AppContext
	source string
	id     string
}

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

func (f *FocusBehavior) SetFocusContext(ctx framework.AppContext, id string) {
	f.ctx = ctx
	f.id = id
}

func (f *FocusBehavior) SetFocusStyle(s framework.Style) {
	f.focusStyle = &s
}

func (f *FocusBehavior) ResolveFocusStyle(base framework.Style) framework.Style {
	if f.focused && f.focusStyle != nil {
		return *framework.ResolveStyle(*f.focusStyle, base)
	}
	return base
}

func (f *FocusBehavior) BindAction(k framework.Key, fn FocusableActionFunc) {
	f.BindActionMod(k, framework.ModNone, fn)
}

func (f *FocusBehavior) BindActionMod(k framework.Key, mods framework.Modifier, fn FocusableActionFunc) {
	f.actionsMu.Lock()
	defer f.actionsMu.Unlock()
	f.actions[framework.Binding{Key: k, Modifiers: mods}] = fn
}

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

// invoke calls fn with a recover in place. fn is caller-supplied --
// BindAction/BindActionMod take an arbitrary closure -- so a bug in
// application code panicking here would otherwise take down whichever
// goroutine is running input dispatch, and a panic in ANY goroutine
// kills the whole process regardless of which one raised it. That
// means the terminal's raw mode would never get a chance to be
// restored: the person using the app is left with a genuinely broken
// terminal (no echo, no line buffering), not just a crashed program.
//
// Recovering here and reporting it as a Fatal-severity log reuses
// machinery this framework already has for exactly this situation:
// fault.FaultManager promotes any Fatal log straight to a SIGTERM,
// which App.Stop() turns into a normal, clean shutdown -- including
// Renderer.Stop()'s terminal restore. A panicking handler still ends
// the app; it just does so gracefully instead of catastrophically.
// The panic is deliberately reported via the SAME logger the bound
// action's own errors already go through (f.logger()), not a
// separate path, so both land in one place a person actually looking
// at logs would check.
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
