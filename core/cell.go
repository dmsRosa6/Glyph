package core

type Cell struct {
    Ch rune
    FG Color
    BG Color
	Dirty bool
}

func NewCell(ch rune, fg, bg Color) *Cell{
	return &Cell{
		Ch: ch,
		FG: fg,
		BG: bg,
		Dirty: true,
	}
}