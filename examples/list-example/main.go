package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func listDemo(width int) *widgets.List {
	list, err := widgets.NewList(geom.NewBounds(2, 2, width, 12), widgets.ListConfig{
		Style:       framework.Style{Bg: core.Transparent, Fg: core.Green},
		ItemPadding: 1,
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 3; i++ {
		row, err := list.AddItem(4)
		if err != nil {
			panic(err)
		}
		text, err := widgets.NewText(&geom.Point{X: 0, Y: 0}, widgets.TextConfig{
			Value: "row",
			Fg:    core.Green,
		})
		if err != nil {
			panic(err)
		}
		row.AddChild(text)
	}

	return list
}

func main() {
	a, err := app.NewApp(app.AppConfig{Bg: &core.Black, RenderMode: render.FixedFPSMode(30)})
	if err != nil {
		panic(err)
	}

	a.Canvas.AddShape(listDemo(24))

	a.Run()
}
