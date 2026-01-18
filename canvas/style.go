package canvas

import "github.com/dmsRosa6/glyph/core"

type Style struct {
	Bg core.Color
	Fg core.Color
}

func NewTransparentStyle() *Style{
	return &Style{
		Bg: core.Transparent,
		Fg: core.Transparent,
	}
}

func ResolveStyle(style, parent Style) *Style{

	newStyle := Style{Fg: style.Fg}

	if style.Fg.IsTransparent{
		newStyle.Bg = core.Transparent
		return &newStyle
	}

	if parent.Bg == (core.Color{}) {
		parent.Bg = core.Transparent
	}

    if style.Bg == (core.Color{}) || style.Bg == core.Transparent{
        newStyle.Bg = parent.Bg
    }else{
        newStyle.Bg = style.Bg
    }

	return &newStyle
}