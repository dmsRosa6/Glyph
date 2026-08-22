package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func main() {
	a, err := app.NewApp(app.AppConfig{Bg: &core.White, RenderMode: render.FixedFPSMode(30)})
	if err != nil {
		panic(err)
	}
	cfg := widgets.SpinnerConfig{
		SpinnerType:    *framework.NewBrailleSpinnerContext(),
		Anchor:         framework.Anchor{V: framework.Center, H: framework.Center},
		Style:          framework.Style{Fg: core.Black, Bg: core.Transparent},
		TicksPerSecond: 10,
	}

	sp, _ := widgets.NewSpinner(cfg)
	a.Canvas.AddShape(sp)

	a.Run()
}
