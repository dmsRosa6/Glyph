package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func main() {
	a, err := app.NewApp(app.AppConfig{Bg: &core.White, RenderMode: render.FixedFPSMode(10)})
	if err != nil {
		panic(err)
	}
	cfg := widgets.SpinnerConfig{SpinnerType: *framework.NewDotsSpinnerContext(), Anchor: framework.Anchor{V: framework.Center, H: framework.Center}}

	sp, _ := widgets.NewSpinner(cfg)
	a.Canvas.AddShape(sp)

	a.Run()
}
