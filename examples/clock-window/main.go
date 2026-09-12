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
	rm, _ := render.FixedFPSMode(30)
	a, err := app.NewApp(app.AppConfig{
		Bg:         core.Transparent,
		RenderMode: rm,
	})
	if err != nil {
		panic(err)
	}

	borderCfg := primitive.BorderConfig{
		BorderStyle: primitive.DoubleLine,
		Thickness:   1,
		Style:       framework.Style{Bg: core.Transparent, Fg: core.DarkSlateGray},
		Layer:       1,
	}
	win, err := widgets.NewWindow(geom.NewBounds(1, 1, 15, 5), widgets.WindowConfig{
		BoxStyle:     framework.Style{Bg: core.DarkSlateGray, Fg: core.White},
		BorderConfig: borderCfg,
		Title:        "Clock",
		TitleFg:      core.DarkSlateGray,
		Anchor:       framework.Anchor{V: framework.Center, H: framework.Center},
	})
	if err != nil {
		panic(err)
	}

	clock, err := primitive.NewText(&geom.Point{X: 1, Y: 1}, primitive.TextConfig{
		Value:  time.Now().Format("15:04:05"),
		Fg:     core.White,
		Anchor: framework.Anchor{V: framework.Center, H: framework.Center},
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
