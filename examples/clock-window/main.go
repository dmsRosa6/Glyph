package main

import (
	"time"

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
	})
	if err != nil {
		panic(err)
	}

	win, err := widgets.NewWindow(geom.NewBounds(1, 1, 26, 5), widgets.WindowConfig{
		BoxStyle:     framework.Style{Bg: core.DarkSlateGray, Fg: core.White},
		BorderConfig: primitive.DefaultBorderConfig(),
		Title:        "Clock",
		TitleFg:      core.White,
	})
	if err != nil {
		panic(err)
	}

	clock, err := primitive.NewText(&geom.Point{X: 1, Y: 1}, primitive.TextConfig{
		Value: time.Now().Format("15:04:05"),
		Fg:    core.White,
	})
	if err != nil {
		panic(err)
	}
	win.AddChild(clock)

	a.Canvas.AddShape(win)

	go func() {
		for range time.Tick(time.Second) {
			clock.SetValue(time.Now().Format("15:04:05"))
		}
	}()

	a.Run()
}
