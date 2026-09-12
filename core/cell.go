package core

// Cell is a small value type -- a rune plus two colors -- not a
// pointer, and there is deliberately no NewCell constructor anymore.
// Buffer stores these directly in a flat slice (see buffer.go) so
// Clear/Set can write values in place with zero heap allocation.
type Cell struct {
	Ch rune
	Fg Color
	Bg Color
}
