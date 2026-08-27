package base

import "github.com/dmsRosa6/glyph/framework"

// FocusBehavior is focus/input-handling behavior as a true mixin: it
// owns no BaseNode of its own. That's the whole point -- the old
// FocusableBaseNode embedded a BaseNode directly, which meant it
// competed with base.Propagator (or *canvas.Container, which embeds
// both BaseNode and Propagator) for "the thing that owns BaseNode."
// Two embedded types both claiming to be the source of BaseNode's
// promoted methods (SetLayer, Style, SetContext, ...) is a compile
// error, which is why no composite could embed FocusableBaseNode
// alongside a Container -- every composite that needed both had to
// hold one of them as a plain field and hand-forward every method the
// field needed (see Window/FocusableBox/ListRow).
//
// FocusBehavior sidesteps this by contributing a disjoint set of
// method names: BindAction, BoundKeys, HandleInput, Focus, Blur,
// IsFocused, SetFocusStyle, ResolveFocusStyle. None of these exist on
// BaseNode or on Container/Propagator, so a struct can embed BaseNode
// (or *canvas.Container) AND FocusBehavior at the same depth with zero
// ambiguity, and every one of those methods is promoted for free.
//
// Like base.Propagator, FocusBehavior keeps its own small copy of
// AppContext (via SetFocusContext -- a deliberately different name
// than BaseNode.SetContext, so promotion never has two candidates to
// choose between) rather than reaching into a sibling BaseNode it
// doesn't own.
type FocusBehavior struct {
	actions    map[framework.Key]FocusableActionFunc
	focused    bool
	focusStyle *framework.Style

	ctx    framework.AppContext
	source string
	id     string
}

// FocusableActionContext is what a bound action receives: the
// FocusBehavior it's bound to (for reaching the node Registry, or
// introspecting bound keys) and the triggering Event. It can't hand
// back a pointer to the owning widget the way the old version did --
// FocusBehavior has no way to know what it's embedded in.
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
		actions: make(map[framework.Key]FocusableActionFunc),
		source:  source,
	}
}

// SetFocusContext wires this behavior's redraw/log hook. id is
// captured once, at the moment context first reaches this behavior --
// the same "not retroactive" caveat Propagator's shadow-key check
// already documents. Fine in practice: SetID is always called at
// construction time, well before a node is attached to the tree.
func (f *FocusBehavior) SetFocusContext(ctx framework.AppContext, id string) {
	f.ctx = ctx
	f.id = id
}

func (f *FocusBehavior) SetFocusStyle(s framework.Style) {
	f.focusStyle = &s
}

// ResolveFocusStyle blends focusStyle over base when focused, else
// returns base unchanged. Composites call this from their own Style()
// override -- the one method this mixin can't eliminate, since
// blending needs both BaseNode's resolved style AND this behavior's
// focused/focusStyle state, and no single embedded type has both.
func (f *FocusBehavior) ResolveFocusStyle(base framework.Style) framework.Style {
	if f.focused && f.focusStyle != nil {
		return *framework.ResolveStyle(*f.focusStyle, base)
	}
	return base
}

func (f *FocusBehavior) BindAction(k framework.Key, fn FocusableActionFunc) {
	f.actions[k] = fn
}

// BoundKeys lists every key this behavior currently has an action
// bound to. Used by Propagator's shadow-warning check.
func (f *FocusBehavior) BoundKeys() []framework.Key {
	keys := make([]framework.Key, 0, len(f.actions))
	for k := range f.actions {
		keys = append(keys, k)
	}
	return keys
}

func (f *FocusBehavior) HandleInput(ev framework.Event) (bool, error) {
	fn, ok := f.actions[ev.Key]
	if !ok {
		return false, nil
	}
	refresh, err := fn(FocusableActionContext{behavior: f, ev: ev})
	if err != nil {
		f.logger().Warning(err)
	}
	if refresh {
		f.ctx.Redraw()
	}
	return true, err
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
	return framework.NewLogger(f.ctx.Logs, f.source, f.id)
}
