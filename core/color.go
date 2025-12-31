package core

type Color struct{
	R uint8
	G uint8
	B uint8
}

func NewColor(r,g,b uint8) *Color{
	return &Color{
		R: r,
		G: g,
		B: b,
	}
}