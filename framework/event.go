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
	KeyCtrlC //Proabaly should be opinionated and be reserved to kill process cleanly
)

type Event struct {
	Key  Key
	Rune rune
}
