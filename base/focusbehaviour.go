package base

import (
	"sync"

	"github.com/dmsRosa6/glyph/framework"
)

// FocusBehavior is focus/input-handling behavior as a true mixin: it
// owns no BaseNode of its own. That's the whole point -- embedding a
// BaseNode directly (the old FocusableBaseNode's design) meant it
// competed with base.Propagator (or *canvas.Container, which embeds
// both BaseNode and Propagator) for "the thing that owns BaseNode."
// Two embedded types both claiming to be the source of BaseNode's
// promoted methods (SetLayer, Style, SetContext, ...) is a compile
// error, which is why no composite could embed both directly -- every
// composite that needed both had to hold one as a plain field and
// hand-forward every method it needed (see Window/FocusableBox/
// ListRow's own doc comments for their specific reasons that field is
// still there for a DIFFERENT reason today).
//
// FocusBehavior sidesteps this by contributing a disjoint set of
// method names: BindAction, BindActionMod, BoundKeys, HandleInput,
// Focus, Blur, IsFocused, SetFocusStyle, ResolveFocusStyle. None of
// these exist on BaseNode or on Container/Propagator, so a struct can
// embed BaseNode (or *canvas.Container) AND FocusBehavior at the same
// depth with zero ambiguity, and every one of those methods is
// promoted for free.
//
// Like base.Propagator, FocusBehavior keeps its own small copy of
// AppContext (via SetFocusContext -- a deliberately different name
// than BaseNode.SetContext, so promotion never has two candidates to
// choose between) rather than reaching into a sibling BaseNode it
// doesn't own.
type FocusBehavior struct {
	// actionsMu guards actions. BindAction/BindActionMod are public,
	// callable from any goroutine at any time (same reasoning as
	// app.App.bindingsMu), while HandleInput reads actions on every
	// input dispatch and BoundKeys reads it whenever Propagator's
	// shadow-key check runs -- i.e. whenever this widget is attached to
	// a container, possibly from a background goroutine.
	//
	// A *sync.RWMutex, not a value: FocusBehavior itself is a value
	// type embedded BY VALUE into every composite that uses it (Window,
	// FocusableBox, ListRow, FocusableBaseNode all copy one out of
	// NewFocusBehavior once, at construction time). go vet's copylocks
	// check flags any struct containing a value-type mutex that's ever
	// returned/copied by value, even a provably-safe one-time
	// zero-value copy like NewFocusBehavior's return. A pointer field
	// sidesteps that entirely: every copy of a constructed
	// FocusBehavior shares the SAME underlying mutex, which is exactly
	// what's needed once it's embedded into a composite and accessed
	// through a pointer receiver from then on -- the standard idiom for
	// "value type that still needs an internal mutex."
	actionsMu  *sync.RWMutex
	actions    map[framework.Binding]FocusableActionFunc
	focused    bool
	focusStyle *framework.Style

	ctx    framework.AppContext
	source string
	id     string
}

// FocusableActionContext is what a bound action receives: the
// FocusBehavior it's bound to (for reaching the node Registry, or
// introspecting bound keys) and the triggering Event. It can't hand
// back a pointer to the owning widget -- FocusBehavior has no way to
// know what it's embedded in.
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

// BindAction binds fn to k with no modifiers (Binding{Key: k,
// Modifiers: framework.ModNone}) -- the common case, e.g. plain typed
// characters on KeyRune. Use BindActionMod for a specific modifier
// combination.
func (f *FocusBehavior) BindAction(k framework.Key, fn FocusableActionFunc) {
	f.BindActionMod(k, framework.ModNone, fn)
}

// BindActionMod binds fn to an exact (k, mods) combination. Exact
// match only -- see framework.Binding's doc comment for why (this is
// what lets a widget bind Shift+Tab independently from plain Tab, and
// is also what stops a plain KeyRune binding from accidentally firing
// on a Ctrl+<letter> event).
func (f *FocusBehavior) BindActionMod(k framework.Key, mods framework.Modifier, fn FocusableActionFunc) {
	f.actionsMu.Lock()
	defer f.actionsMu.Unlock()
	f.actions[framework.Binding{Key: k, Modifiers: mods}] = fn
}

// BoundKeys lists every (key, modifiers) binding this behavior
// currently has an action for. Used by Propagator's shadow-warning
// check.
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
	return framework.NewLogger(f.ctx.Logs, f.ctx.LogLevel, f.source, f.id)
}
