package framework

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

type Event struct {
	Key  Key
	Rune rune
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
