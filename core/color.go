package core

type Color struct {
	R             uint8
	G             uint8
	B             uint8
	IsTransparent bool
}

func NewColor(r, g, b uint8) Color {
	return Color{
		R:             r,
		G:             g,
		B:             b,
		IsTransparent: false,
	}
}

var (
	Transparent = Color{IsTransparent: true}

	// Reds
	Maroon      = Color{R: 128, G: 0, B: 0}
	DarkRed     = Color{R: 139, G: 0, B: 0}
	Brown       = Color{R: 165, G: 42, B: 42}
	Firebrick   = Color{R: 178, G: 34, B: 34}
	Crimson     = Color{R: 220, G: 20, B: 60}
	Red         = Color{R: 255, G: 0, B: 0}
	Tomato      = Color{R: 255, G: 99, B: 71}
	Coral       = Color{R: 255, G: 127, B: 80}
	IndianRed   = Color{R: 205, G: 92, B: 92}
	LightCoral  = Color{R: 240, G: 128, B: 128}
	DarkSalmon  = Color{R: 233, G: 150, B: 122}
	Salmon      = Color{R: 250, G: 128, B: 114}
	LightSalmon = Color{R: 255, G: 160, B: 122}
	OrangeRed   = Color{R: 255, G: 69, B: 0}

	// Oranges / Yellows
	DarkOrange     = Color{R: 255, G: 140, B: 0}
	Orange         = Color{R: 255, G: 165, B: 0}
	Gold           = Color{R: 255, G: 215, B: 0}
	DarkGoldenRod  = Color{R: 184, G: 134, B: 11}
	GoldenRod      = Color{R: 218, G: 165, B: 32}
	PaleGoldenRod  = Color{R: 238, G: 232, B: 170}
	DarkKhaki      = Color{R: 189, G: 183, B: 107}
	Khaki          = Color{R: 240, G: 230, B: 140}
	Olive          = Color{R: 128, G: 128, B: 0}
	Yellow         = Color{R: 255, G: 255, B: 0}
	YellowGreen    = Color{R: 154, G: 205, B: 50}
	DarkOliveGreen = Color{R: 85, G: 107, B: 47}
	OliveDrab      = Color{R: 107, G: 142, B: 35}
	LawnGreen      = Color{R: 124, G: 252, B: 0}
	Chartreuse     = Color{R: 127, G: 255, B: 0}
	GreenYellow    = Color{R: 173, G: 255, B: 47}

	// Greens
	DarkGreen         = Color{R: 0, G: 100, B: 0}
	Green             = Color{R: 0, G: 128, B: 0}
	ForestGreen       = Color{R: 34, G: 139, B: 34}
	Lime              = Color{R: 0, G: 255, B: 0}
	LimeGreen         = Color{R: 50, G: 205, B: 50}
	LightGreen        = Color{R: 144, G: 238, B: 144}
	PaleGreen         = Color{R: 152, G: 251, B: 152}
	DarkSeaGreen      = Color{R: 143, G: 188, B: 143}
	MediumSpringGreen = Color{R: 0, G: 250, B: 154}
	SpringGreen       = Color{R: 0, G: 255, B: 127}
	SeaGreen          = Color{R: 46, G: 139, B: 87}
	MediumAquaMarine  = Color{R: 102, G: 205, B: 170}
	MediumSeaGreen    = Color{R: 60, G: 179, B: 113}
	LightSeaGreen     = Color{R: 32, G: 178, B: 170}

	// Cyans / Teals
	DarkSlateGray   = Color{R: 47, G: 79, B: 79}
	Teal            = Color{R: 0, G: 128, B: 128}
	DarkCyan        = Color{R: 0, G: 139, B: 139}
	Aqua            = Color{R: 0, G: 255, B: 255}
	Cyan            = Color{R: 0, G: 255, B: 255}
	LightCyan       = Color{R: 224, G: 255, B: 255}
	DarkTurquoise   = Color{R: 0, G: 206, B: 209}
	Turquoise       = Color{R: 64, G: 224, B: 208}
	MediumTurquoise = Color{R: 72, G: 209, B: 204}
	PaleTurquoise   = Color{R: 175, G: 238, B: 238}
	AquaMarine      = Color{R: 127, G: 255, B: 212}

	// Blues
	PowderBlue     = Color{R: 176, G: 224, B: 230}
	CadetBlue      = Color{R: 95, G: 158, B: 160}
	SteelBlue      = Color{R: 70, G: 130, B: 180}
	CornFlowerBlue = Color{R: 100, G: 149, B: 237}
	DeepSkyBlue    = Color{R: 0, G: 191, B: 255}
	DodgerBlue     = Color{R: 30, G: 144, B: 255}
	LightBlue      = Color{R: 173, G: 216, B: 230}
	SkyBlue        = Color{R: 135, G: 206, B: 235}
	LightSkyBlue   = Color{R: 135, G: 206, B: 250}
	MidnightBlue   = Color{R: 25, G: 25, B: 112}
	Navy           = Color{R: 0, G: 0, B: 128}
	DarkBlue       = Color{R: 0, G: 0, B: 139}
	MediumBlue     = Color{R: 0, G: 0, B: 205}
	Blue           = Color{R: 0, G: 0, B: 255}
	RoyalBlue      = Color{R: 65, G: 105, B: 225}

	// Purples / Violets
	BlueViolet      = Color{R: 138, G: 43, B: 226}
	Indigo          = Color{R: 75, G: 0, B: 130}
	DarkSlateBlue   = Color{R: 72, G: 61, B: 139}
	SlateBlue       = Color{R: 106, G: 90, B: 205}
	MediumSlateBlue = Color{R: 123, G: 104, B: 238}
	MediumPurple    = Color{R: 147, G: 112, B: 219}
	DarkMagenta     = Color{R: 139, G: 0, B: 139}
	DarkViolet      = Color{R: 148, G: 0, B: 211}
	DarkOrchid      = Color{R: 153, G: 50, B: 204}
	MediumOrchid    = Color{R: 186, G: 85, B: 211}
	Purple          = Color{R: 128, G: 0, B: 128}
	Thistle         = Color{R: 216, G: 191, B: 216}
	Plum            = Color{R: 221, G: 160, B: 221}
	Violet          = Color{R: 238, G: 130, B: 238}
	Magenta         = Color{R: 255, G: 0, B: 255}
	Orchid          = Color{R: 218, G: 112, B: 214}
	MediumVioletRed = Color{R: 199, G: 21, B: 133}
	PaleVioletRed   = Color{R: 219, G: 112, B: 147}
	DeepPink        = Color{R: 255, G: 20, B: 147}
	HotPink         = Color{R: 255, G: 105, B: 180}
	LightPink       = Color{R: 255, G: 182, B: 193}
	Pink            = Color{R: 255, G: 192, B: 203}

	// Browns / Tans
	AntiqueWhite         = Color{R: 250, G: 235, B: 215}
	Beige                = Color{R: 245, G: 245, B: 220}
	Bisque               = Color{R: 255, G: 228, B: 196}
	BlanchedAlmond       = Color{R: 255, G: 235, B: 205}
	Wheat                = Color{R: 245, G: 222, B: 179}
	CornSilk             = Color{R: 255, G: 248, B: 220}
	LemonChiffon         = Color{R: 255, G: 250, B: 205}
	LightGoldenRodYellow = Color{R: 250, G: 250, B: 210}
	LightYellow          = Color{R: 255, G: 255, B: 224}
	SaddleBrown          = Color{R: 139, G: 69, B: 19}
	Sienna               = Color{R: 160, G: 82, B: 45}
	Chocolate            = Color{R: 210, G: 105, B: 30}
	Peru                 = Color{R: 205, G: 133, B: 63}
	SandyBrown           = Color{R: 244, G: 164, B: 96}
	BurlyWood            = Color{R: 222, G: 184, B: 135}
	Tan                  = Color{R: 210, G: 180, B: 140}
	RosyBrown            = Color{R: 188, G: 143, B: 143}
	Moccasin             = Color{R: 255, G: 228, B: 181}
	NavajoWhite          = Color{R: 255, G: 222, B: 173}
	PeachPuff            = Color{R: 255, G: 218, B: 185}
	MistyRose            = Color{R: 255, G: 228, B: 225}
	LavenderBlush        = Color{R: 255, G: 240, B: 245}
	Linen                = Color{R: 250, G: 240, B: 230}
	OldLace              = Color{R: 253, G: 245, B: 230}
	PapayaWhip           = Color{R: 255, G: 239, B: 213}
	SeaShell             = Color{R: 255, G: 245, B: 238}
	MintCream            = Color{R: 245, G: 255, B: 250}

	// Light / Pale colors
	SlateGray      = Color{R: 112, G: 128, B: 144}
	LightSlateGray = Color{R: 119, G: 136, B: 153}
	LightSteelBlue = Color{R: 176, G: 196, B: 222}
	Lavender       = Color{R: 230, G: 230, B: 250}
	FloralWhite    = Color{R: 255, G: 250, B: 240}
	AliceBlue      = Color{R: 240, G: 248, B: 255}
	GhostWhite     = Color{R: 248, G: 248, B: 255}
	Honeydew       = Color{R: 240, G: 255, B: 240}
	Ivory          = Color{R: 255, G: 255, B: 240}
	Azure          = Color{R: 240, G: 255, B: 255}
	Snow           = Color{R: 255, G: 250, B: 250}

	// Grays
	Black      = Color{R: 0, G: 0, B: 0}
	DimGray    = Color{R: 105, G: 105, B: 105}
	Gray       = Color{R: 128, G: 128, B: 128}
	DarkGray   = Color{R: 169, G: 169, B: 169}
	Silver     = Color{R: 192, G: 192, B: 192}
	LightGray  = Color{R: 211, G: 211, B: 211}
	Gainsboro  = Color{R: 220, G: 220, B: 220}
	WhiteSmoke = Color{R: 245, G: 245, B: 245}
	White      = Color{R: 255, G: 255, B: 255}
)
