package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/primitive"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func main() {
	a, err := app.NewApp(app.AppConfig{
		Width:      40,
		Height:     14,
		Bg:         core.Black,
		RenderMode: render.OnDemandMode(),
		AppEvents:  app.NavActions(),
	})
	if err != nil {
		panic(err)
	}

	focus := framework.StyleBg(core.SteelBlue)

	outer, err := widgets.NewFocusableBox(geom.NewBounds(1, 1, 36, 10), widgets.FocusableBoxConfig{
		Padding:      1,
		BorderConfig: primitive.DefaultBorderConfig(),
		Style:        framework.Style{Bg: core.DarkSlateGray, Fg: core.White},
		FocusStyle:   &focus,
	})
	if err != nil {
		panic(err)
	}

	inner, err := widgets.NewFocusableBox(geom.NewBounds(1, 1, 20, 6), widgets.FocusableBoxConfig{
		Padding:      1,
		BorderConfig: primitive.DefaultBorderConfig(),
		Style:        framework.Style{Bg: core.DarkOliveGreen, Fg: core.White},
		FocusStyle:   &focus,
	})
	if err != nil {
		panic(err)
	}

	label, err := primitive.NewText(&geom.Point{X: 0, Y: 0}, primitive.TextConfig{
		Value: "Enter to drill, Esc to exit",
		Fg:    core.White,
	})
	if err != nil {
		panic(err)
	}
	inner.AddChild(label)
	outer.AddChild(inner)

	a.Canvas.AddShape(outer)

	a.Run()
}
