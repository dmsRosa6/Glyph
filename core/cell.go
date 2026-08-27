package core

type Cell struct {
	Ch rune
	Fg Color
	Bg Color
}

func NewCell(ch rune, bg, fg Color) *Cell {
	return &Cell{
		Ch: ch,
		Fg: fg,
		Bg: bg,
	}
}
