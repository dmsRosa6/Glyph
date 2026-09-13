package main

import (
	"fmt"

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
		Width:      30,
		Height:     8,
		Bg:         core.Black,
		RenderMode: render.OnDemandMode(),
		AppEvents:  app.NavActions(), // Tab/Shift+Tab to move focus between rows
	})
	if err != nil {
		panic(err)
	}

	list, err := widgets.NewList(geom.NewBounds(1, 1, 28, 6), widgets.ListConfig{
		Style:      framework.Style{Bg: core.DarkSlateGray, Fg: core.White},
		Scrollable: true, // more rows than fit -- the list auto-scrolls to keep the focused one visible
	})
	if err != nil {
		panic(err)
	}

	for i := 1; i <= 20; i++ {
		row, err := list.AddItem(1)
		if err != nil {
			panic(err)
		}
		label, err := primitive.NewText(&geom.Point{X: 0, Y: 0}, primitive.TextConfig{
			Value: fmt.Sprintf("Row %d", i),
			Fg:    core.White,
		})
		if err != nil {
			panic(err)
		}
		row.AddChild(label)
	}

	a.Canvas.AddShape(list)

	a.Run()
}
