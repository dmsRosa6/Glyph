package main

import (
	"time"

	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func clockWindow() *widgets.Window {
	win, err := widgets.NewWindow(geom.NewBounds(0, 0, 30, 10), widgets.WindowConfig{
		Padding:  1,
		BoxStyle: framework.Style{Bg: core.Transparent, Fg: core.Red},
		BorderConfig: widgets.BorderConfig{
			Thickness:   1,
			BorderStyle: widgets.DoubleLine,
			Style:       framework.Style{Bg: core.Transparent, Fg: core.LightBlue},
		},
		Anchor:  framework.Anchor{V: framework.Center, H: framework.Center},
		Title:   " Clock ",
		TitleFg: core.HotPink,
	})
	if err != nil {
		panic(err)
	}

	clock, err := widgets.NewText(&geom.Point{X: 0, Y: 0}, widgets.TextConfig{
		Value:  "00.00.00",
		Anchor: framework.Anchor{V: framework.Center, H: framework.Center},
		Fg:     core.HotPink,
	})
	if err != nil {
		panic(err)
	}
	win.AddChild(clock)

	go func() {
		for range time.Tick(time.Second) {
			clock.SetValue(time.Now().Format("15:04:05"))
		}
	}()

	return win
}

func main() {
	a, err := app.NewApp(app.AppConfig{Bg: &core.Black, RenderMode: render.FixedFPSMode(30)})
	if err != nil {
		panic(err)
	}

	a.Canvas.AddShape(clockWindow())

	a.Run()
}
