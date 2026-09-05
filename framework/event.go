package framework

// EventKind discriminates what a framework.Event actually carries.
// Event is one flat struct for both key and mouse input rather than an
// interface or two separate types/channels -- Kind says which half of
// the struct is populated, the other half sits at its zero value. This
// is the whole point of the design: AppActionFunc, FocusableActionFunc,
// and HandleInput's signatures, and the single chan framework.Event
// input.Manager already exposes, don't need to change AT ALL to add
// mouse support. EventKindKey is the zero value, so every existing
// framework.Event{Key: ...} literal anywhere in this codebase (or in
// any app built on it) keeps meaning exactly what it meant before this
// field existed.
type EventKind int

const (
	EventKindKey EventKind = iota
	EventKindMouse
)

func (k EventKind) String() string {
	switch k {
	case EventKindKey:
		return "Key"
	case EventKindMouse:
		return "Mouse"
	default:
		return "Unknown"
	}
}

type Key int

const (
	KeyRune Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEsc
	KeyTab
	KeyCtrlC
)

// Modifier is a bitmask, not more named Key constants -- a single Event
// needs to carry any combination (Ctrl+Shift+Left) without a
// combinatorial explosion of one-off Key values the way KeyCtrlC would
// require if extended that way. KeyCtrlC itself is deliberately left
// untouched: it's a pre-existing, load-bearing special case (App's
// unconditional Ctrl+C-quit binding; term.SafeRawMode disables ISIG
// specifically so this byte arrives as ordinary input instead of a
// real SIGINT, and this is the only thing that lets a keyboard-only
// user exit) with no ambiguity to gain from being reframed as
// KeyRune{'c'} + ModCtrl. Every OTHER Ctrl+<letter> combination decodes
// generically through Modifiers instead -- see input/manager.go's
// handleNormal.
type Modifier int

// ModNone is declared in its own const block deliberately -- putting it
// in the same block as the 1<<iota sequence below would consume iota's
// zero slot and silently shift ModShift/ModAlt/ModCtrl to 2/4/8 instead
// of 1/2/4. Functionally harmless (everything downstream ORs/checks via
// these names, never a raw bit literal), but wrong is wrong -- caught
// by actually running iota through the compiler rather than assuming.
const ModNone Modifier = 0

const (
	ModShift Modifier = 1 << iota
	ModAlt
	ModCtrl
)

func (m Modifier) Has(flag Modifier) bool {
	return m&flag != 0
}

func (m Modifier) String() string {
	if m == 0 {
		return ""
	}
	s := ""
	if m.Has(ModCtrl) {
		s += "Ctrl+"
	}
	if m.Has(ModAlt) {
		s += "Alt+"
	}
	if m.Has(ModShift) {
		s += "Shift+"
	}
	return s
}

// Binding is the (Key, Modifiers) pair a handler is registered under --
// EXACT match, not a filter applied after lookup. This is what makes
// Shift+Tab and plain Tab independently bindable: a global or per-widget
// binding on Binding{Key: KeyTab} matches ONLY Modifiers == ModNone;
// Shift+Tab (Modifiers: ModShift) needs its own separate
// Binding{Key: KeyTab, Modifiers: ModShift} entry, rather than requiring
// one handler to inspect ev.Modifiers itself to tell the two apart.
//
// This also closes a real correctness gap for free: a plain
// BindAction(KeyRune, handler) -- the natural way to build a
// text-input-style widget -- binds Binding{KeyRune, ModNone}, so it will
// NOT match a Ctrl+H event even though Ctrl+H's Rune is now the
// printable 'h' (see handleNormal's doc comment in input/manager.go).
// Before Binding existed, a KeyRune-keyed map matched every modifier
// variant of KeyRune, so a naive unicode.IsPrint(ev.Rune) filter in a
// text-input handler would have silently accepted Ctrl+H as literal
// text; with exact (Key, Modifiers) matching, that handler's binding
// simply never fires for it.
type Binding struct {
	Key       Key
	Modifiers Modifier
}

// MouseButton identifies which physical button a mouse Event concerns.
// MouseButtonNone covers reports that aren't about a specific button at
// all -- wheel events, and (per the SGR mouse protocol) any report
// where the "button" bits are ambiguous by the protocol's own design.
type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonMiddle
	MouseButtonRight
	MouseButtonNone
)

// MouseAction is what happened, as opposed to MouseButton (which one).
type MouseAction int

const (
	MousePress MouseAction = iota
	MouseRelease
	MouseDrag
	MouseWheelUp
	MouseWheelDown
)

func (a MouseAction) String() string {
	switch a {
	case MousePress:
		return "press"
	case MouseRelease:
		return "release"
	case MouseDrag:
		return "drag"
	case MouseWheelUp:
		return "wheel-up"
	case MouseWheelDown:
		return "wheel-down"
	default:
		return "unknown"
	}
}

// Event is intentionally one flat struct for both key and mouse input.
// See EventKind's doc comment for why. Modifiers applies to both kinds
// -- Ctrl+Left is a key Event with ModCtrl set; Shift+Click is a mouse
// Event with ModShift set.
type Event struct {
	Kind EventKind

	// Populated when Kind == EventKindKey. Zero values (KeyRune, rune
	// 0) when Kind == EventKindMouse -- see app.go's dispatch loop for
	// why mouse events are routed on Kind BEFORE anything ever looks at
	// Key, so a mouse event can never be mistaken for a KeyRune keypress
	// by a widget's key-indexed action map.
	Key  Key
	Rune rune

	// Populated when Kind == EventKindMouse. Zero values otherwise.
	// MouseX/MouseY are 0-indexed, matching every other coordinate in
	// this codebase (geom.Point, core.Buffer, BaseNode.ComputedPos are
	// all 0-indexed from the top-left) -- the wire protocol itself is
	// 1-indexed; input.Manager converts on decode.
	MouseButton MouseButton
	MouseAction MouseAction
	MouseX      int
	MouseY      int

	Modifiers Modifier
}

func (k Key) String() string {
	switch k {
	case KeyRune:
		return "Rune"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyEnter:
		return "Enter"
	case KeyEsc:
		return "Escape"
	case KeyTab:
		return "Tab"
	case KeyCtrlC:
		return "Ctrl+C"
	default:
		return "Unknown"
	}
}

func IsStructuralKey(k Key) bool {
	switch k {
	case KeyCtrlC, KeyEnter, KeyTab, KeyEsc:
		return true
	default:
		return false
	}
}
