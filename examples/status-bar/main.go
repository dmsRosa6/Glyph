package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func statusItem(width int, label string, color core.Color) *widgets.Bordered {
	bounds := geom.NewBounds(0, 0, width, 5)

	box, err := widgets.NewBox(bounds, widgets.BoxConfig{
		Padding: 1,
		Style:   framework.Style{Bg: core.Transparent, Fg: color},
		BorderConfig: widgets.BorderConfig{
			Thickness:   1,
			BorderStyle: widgets.Rounded,
			Style:       framework.Style{Bg: core.Transparent, Fg: color},
		},
	})
	if err != nil {
		panic(err)
	}

	text, err := widgets.NewText(&geom.Point{X: 0, Y: 0}, widgets.TextConfig{
		Value: label,
		Fg:    color,
	})
	if err != nil {
		panic(err)
	}

	box.AddChild(text)

	return box
}

func main() {
	a, err := app.NewApp(app.AppConfig{Bg: &core.Black, RenderMode: render.FixedFPSMode(30)})
	if err != nil {
		panic(err)
	}

	items := []struct {
		label string
		color core.Color
	}{
		{"API", core.Green},
		{"Queue", core.Yellow},
		{"DB", core.Red},
	}

	for i, it := range items {
		item := statusItem(20, it.label, it.color)
		item.Bounds().Pos.X = 2
		item.Bounds().Pos.Y = 2 + i*6
		a.Canvas.AddShape(item)
	}

	a.Run()
}
