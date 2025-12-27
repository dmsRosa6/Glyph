package core

// TODO maybe we can use 16 bits for colors
type Color struct{
	R int8
	G int8
	B int8
}

func NewColor(r,g,b int8) *Color{
	return &Color{
		R: r,
		G: g,
		B: b,
	}
}