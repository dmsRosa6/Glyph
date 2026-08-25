package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func main() {
	a, err := app.NewApp(app.AppConfig{
		Width:      40,
		Height:     10,
		Bg:         &core.Black,
		RenderMode: render.FixedFPSMode(30),
	})
	if err != nil {
		panic(err)
	}

	tooBig, err := widgets.NewWindow(geom.NewBounds(1, 1, 16, 6), widgets.WindowConfig{
		BoxStyle:     framework.Style{Bg: core.DarkSlateGray, Fg: core.White},
		BorderConfig: widgets.DefaultBorderConfig(),
		Title:        "Too Big",
		TitleFg:      core.White,
	})
	if err != nil {
		panic(err)
	}

	overflow, err := widgets.NewRect(geom.NewBounds(0, 0, 40, 20), widgets.RectConfig{
		Style: framework.Style{Bg: core.Firebrick},
	})
	if err != nil {
		panic(err)
	}
	tooBig.AddChild(overflow)

	fits, err := widgets.NewWindow(geom.NewBounds(20, 1, 16, 6), widgets.WindowConfig{
		BoxStyle:     framework.Style{Bg: core.DarkSlateGray, Fg: core.White},
		BorderConfig: widgets.DefaultBorderConfig(),
		Title:        "Fits",
		TitleFg:      core.White,
	})
	if err != nil {
		panic(err)
	}

	label, err := widgets.NewText(&geom.Point{X: 1, Y: 1}, widgets.TextConfig{
		Value: "inside",
		Fg:    core.White,
	})
	if err != nil {
		panic(err)
	}
	fits.AddChild(label)

	a.Canvas.AddShape(tooBig)
	a.Canvas.AddShape(fits)

	a.Run()
}
