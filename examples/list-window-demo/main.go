package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func newCard(bounds *geom.Bounds, title string, bodyColor core.Color) (*widgets.Window, error) {
	focus := framework.StyleBg(core.SteelBlue) // Bg: SteelBlue, Fg: Transparent -- explicit, no accidental black
	card, err := widgets.NewWindow(bounds, widgets.WindowConfig{
		BoxStyle:     framework.Style{Bg: bodyColor, Fg: core.White},
		FocusStyle:   &focus,
		BorderConfig: widgets.DefaultBorderConfig(),
		Title:        title,
		TitleFg:      core.White,
	})
	if err != nil {
		return nil, err
	}

	label, err := widgets.NewText(&geom.Point{X: 1, Y: 1}, widgets.TextConfig{
		Value: "Tab to switch",
		Fg:    core.White,
	})
	if err != nil {
		return nil, err
	}
	card.AddChild(label)

	return card, nil
}

func main() {
	cardA, err := newCard(geom.NewBounds(2, 1, 22, 7), "Card A", core.DarkSlateGray)
	if err != nil {
		panic(err)
	}

	cardB, err := newCard(geom.NewBounds(10, 4, 22, 7), "Card B", core.DarkOliveGreen)
	if err != nil {
		panic(err)
	}

	a, err := app.NewApp(app.AppConfig{
		Width:      36,
		Height:     12,
		Bg:         &core.Black,
		RenderMode: render.FixedFPSMode(30),
		AppEvents: map[framework.Key]app.AppActionFunc{
			// Ctrl+C is already seeded in by NewApp -- this app only
			// needs to add the one binding it actually cares about.
			framework.KeyTab: func(ctx framework.AppContext, ev framework.Event) (bool, error) {
				ctx.Nav().Next()
				return true, nil
			},
		},
	})
	if err != nil {
		panic(err)
	}

	a.Canvas.AddShape(cardA)
	a.Canvas.AddShape(cardB)

	a.Run()
}
