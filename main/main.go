package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
)

func statusItem(width int, label string, color core.Color) *canvas.Bordered {
	bounds := geom.NewBounds(0, 0, width, 3)

	box, err := canvas.NewBox(bounds, canvas.BoxConfig{
		Padding: 1,
		Style:   canvas.Style{Bg: core.Transparent, Fg: color},
		BorderConfig: canvas.BorderConfig{
			Thickness:   1,
			BorderStyle: canvas.Rounded,
			Style:       canvas.Style{Bg: core.Transparent, Fg: color},
		},
	})
	if err != nil {
		panic(err)
	}

	text, err := canvas.NewText(&geom.Point{X: 0, Y: 0}, canvas.TextConfig{
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

	appCfg := app.AppConfig{
		Bg: &core.Black,
		RenderMode: render.OnDemand,
	}

	a, _ := app.NewApp(appCfg)

	// The whole UI lives in one Window: a double-line border, a title on
	// the frame itself, padded content.
	win, err := canvas.NewWindow(geom.NewBounds(1, 1, 48, 18), canvas.WindowConfig{
		BoxConfig: canvas.BoxConfig{
			Padding: 1,
			BorderConfig: canvas.BorderConfig{
				Thickness:   1,
				BorderStyle: canvas.DoubleLine,
				Style:       canvas.Style{Bg: core.Transparent, Fg: core.White},
			},
		},
		Title:         "SYSTEM STATUS",
		TitleXOffset:  1,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.Yellow,
	})
	if err != nil {
		panic(err)
	}

	innerWidth := 46 // window content width: 48 - 2*padding

	// A plain Rect used as a colored divider line, not a text label.
	divider, err := canvas.NewRect(geom.NewBounds(0, 0, innerWidth, 1), canvas.RectConfig{
		Ch:    '─',
		Style: canvas.Style{Bg: core.Transparent, Fg: core.White},
	})
	if err != nil {
		panic(err)
	}
	win.AddChild(divider)

	statusList, err := canvas.NewList(geom.NewBounds(0, 2, innerWidth, 9), canvas.ListConfig{})
	if err != nil {
		panic(err)
	}
	statusList.AddChild(statusItem(innerWidth, "> CPU      OK", core.Green))
	statusList.AddChild(statusItem(innerWidth, "> MEMORY   WARN", core.Yellow))
	statusList.AddChild(statusItem(innerWidth, "> DISK     ERROR", core.Red))
	win.AddChild(statusList)

	footer, err := canvas.NewText(&geom.Point{X: 0, Y: 15}, canvas.TextConfig{
		Value:  "v1.0.0",
		Fg:     core.White,
		Anchor: canvas.Anchor{H: canvas.End},
	})
	if err != nil {
		panic(err)
	}
	win.AddChild(footer)

	a.Canvas.AddShape(win)

	a.Run()
}

