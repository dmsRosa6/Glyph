package core

type Color struct{
	R uint8
	G uint8
	B uint8
	IsTransparent bool //TODO check is this could be an alpha maybe it would be cool
}

func NewColor(r,g,b uint8) Color{
	return Color{
		R: r,
		G: g,
		B: b,
		IsTransparent: false,
	}
}

var Transparent Color = Color{IsTransparent: true}
var Black Color = Color{IsTransparent: false}
var White Color = Color{R:255,G:255,B:255,IsTransparent: false}