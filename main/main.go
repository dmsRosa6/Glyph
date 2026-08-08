package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
)

func statusItem(width int, label string, color core.Color) *canvas.Bordered {
	bounds := geom.NewBounds(0, 0, width, 3)

	box, err := canvas.NewBox(bounds, canvas.BoxConfig{
		Padding: 1,
		Style:   framework.Style{Bg: core.Transparent, Fg: color},
		BorderConfig: canvas.BorderConfig{
			Thickness:   1,
			BorderStyle: canvas.Rounded,
			Style:       framework.Style{Bg: core.Transparent, Fg: color},
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
	a, _ := app.NewApp(app.AppConfig{Bg: &core.Black, RenderMode: render.FixedFPS})

	// Box A: a container that drills into Box B on Enter (no action of
	// its own bound to Enter, and it has a focusable child).
	boxA, _ := canvas.NewFocusableBox(geom.NewBounds(2, 2, 20, 8), canvas.FocusableBoxConfig{
		Padding: 1, BorderConfig: canvas.DefaultBorderConfig(),
		Style:      framework.Style{Bg: core.White, Fg: core.White},
		FocusStyle: framework.Style{Bg: core.Red, Fg: core.Red},
	})

	// Box B: lives inside Box A, has no children, so Enter falls through
	// to its own bound action instead of drilling further.
	boxB, _ := canvas.NewFocusableBox(geom.NewBounds(0, 0, 14, 4), canvas.FocusableBoxConfig{
		Padding: 1, BorderConfig: canvas.DefaultBorderConfig(),
		Style:      framework.Style{Bg: core.Blue, Fg: core.Blue},
		FocusStyle: framework.Style{Bg: core.Red, Fg: core.Red},
	})
	boxB.BindAction(framework.KeyEnter, func(n *canvas.FocusableBaseNode, ev framework.Event) (bool, error) {
		boxA, _ = canvas.NewFocusableBox(geom.NewBounds(0, 0, 14, 4), canvas.FocusableBoxConfig{
		Padding: 1, BorderConfig: canvas.DefaultBorderConfig(),
		Style:      framework.Style{Bg: core.White, Fg: core.White},
		FocusStyle: framework.Style{Bg: core.Red, Fg: core.Red},
	})
		return true, nil // triggers a redraw
	})
	boxA.AddChild(boxB)

	a.Canvas.AddShape(boxA)
	a.Run()
}
