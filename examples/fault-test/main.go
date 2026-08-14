package main

import (
	"errors"

	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func faultTestPanel() *widgets.FocusableBox {
	panel, err := widgets.NewFocusableBox(geom.NewBounds(2, 2, 26, 5), widgets.FocusableBoxConfig{
		Padding:      1,
		BorderConfig: widgets.DefaultBorderConfig(),
		Style:        framework.Style{Bg: core.Transparent, Fg: core.White},
		FocusStyle:   framework.Style{Bg: core.Transparent, Fg: core.Yellow},
	})
	if err != nil {
		panic(err)
	}

	label, err := widgets.NewText(&geom.Point{X: 0, Y: 0}, widgets.TextConfig{
		Value: "Up=warn Down=fatal",
		Fg:    core.White,
	})
	if err != nil {
		panic(err)
	}
	panel.AddChild(label)

	panel.BindAction(framework.KeyUp, func(base.FocusableActionContext) (bool, error) {
		panel.Warn("FaultTestPanel", errors.New("manually triggered warning"))
		return false, nil
	})

	panel.BindAction(framework.KeyDown, func(base.FocusableActionContext) (bool, error) {
		panel.Fault("FaultTestPanel", errors.New("manually triggered fatal"))
		return false, nil
	})

	return panel
}

func main() {
	a, err := app.NewApp(app.AppConfig{Bg: &core.Black, RenderMode: render.FixedFPSMode(30)})
	if err != nil {
		panic(err)
	}

	a.Canvas.AddShape(faultTestPanel())

	a.Run()
}
