package core

// Cell is a small value type -- a rune plus two colors -- used mainly on the buffer
type Cell struct {
	Ch rune
	Fg Color
	Bg Color
}
