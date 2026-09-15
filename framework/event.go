package framework

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
	KeyBackspace
	KeyDelete
	KeyHome
	KeyEnd
)

type Modifier int

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

type Binding struct {
	Key       Key
	Modifiers Modifier
}

type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonMiddle
	MouseButtonRight
	MouseButtonNone
)

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

type Event struct {
	Kind EventKind

	Key  Key
	Rune rune

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
	case KeyBackspace:
		return "Backspace"
	case KeyDelete:
		return "Delete"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
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
