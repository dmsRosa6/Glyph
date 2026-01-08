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

var (
    Transparent = Color{IsTransparent: true}
    Black       = Color{R: 0, G: 0, B: 0, IsTransparent: false}
    White       = Color{R: 255, G: 255, B: 255, IsTransparent: false}
    Red         = Color{R: 255, G: 0, B: 0, IsTransparent: false}
    Green       = Color{R: 0, G: 255, B: 0, IsTransparent: false}
    Blue        = Color{R: 0, G: 0, B: 255, IsTransparent: false}
    Yellow      = Color{R: 255, G: 255, B: 0, IsTransparent: false}
    Gray      	= Color{R: 200, G: 200, B: 200, IsTransparent: false}

)
