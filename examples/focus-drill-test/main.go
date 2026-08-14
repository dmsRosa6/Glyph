package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func focusDrillDemo() *widgets.FocusableBox {
	outer, err := widgets.NewFocusableBox(geom.NewBounds(2, 2, 24, 8), widgets.FocusableBoxConfig{
		Padding:      1,
		BorderConfig: widgets.DefaultBorderConfig(),
		Style:        framework.Style{Bg: core.Transparent, Fg: core.White},
		FocusStyle:   framework.Style{Bg: core.Transparent, Fg: core.Red},
	})
	if err != nil {
		panic(err)
	}

	inner, err := widgets.NewFocusableBox(geom.NewBounds(0, 0, 14, 4), widgets.FocusableBoxConfig{
		Padding:      1,
		BorderConfig: widgets.DefaultBorderConfig(),
		Style:        framework.Style{Bg: core.Transparent, Fg: core.Blue},
		FocusStyle:   framework.Style{Bg: core.Transparent, Fg: core.Red},
	})
	if err != nil {
		panic(err)
	}

	label, err := widgets.NewText(&geom.Point{X: 0, Y: 0}, widgets.TextConfig{
		Value: "Enter->drill",
		Fg:    core.White,
	})
	if err != nil {
		panic(err)
	}
	inner.AddChild(label)

	outer.AddChild(inner)
	return outer
}

func main() {
	a, err := app.NewApp(app.AppConfig{Bg: &core.Black, RenderMode: render.FixedFPSMode(30)})
	if err != nil {
		panic(err)
	}

	a.Canvas.AddShape(focusDrillDemo())

	a.Run()
}
