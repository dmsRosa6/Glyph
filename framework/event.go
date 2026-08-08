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
	KeyCtrlC //Reserved to kill process cleanly
)

type Event struct {
	Key  Key
	Rune rune
}
