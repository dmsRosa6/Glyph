package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/primitive"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	rm, err := render.FixedFPSMode(30)
	if err != nil {
		panic(err)
	}
	a, err := app.NewApp(app.AppConfig{Bg: core.White, RenderMode: rm})
	if err != nil {
		panic(err)
	}
	cfg := primitive.SpinnerConfig{
		SpinnerType:    *framework.NewPulseSpinnerContext(),
		Anchor:         framework.Anchor{V: framework.Center, H: framework.Center},
		Style:          framework.Style{Fg: core.Black, Bg: core.Transparent},
		TicksPerSecond: 10,
	}

	sp, _ := primitive.NewSpinner(cfg)
	a.Canvas.AddShape(sp)

	a.Run()
}
