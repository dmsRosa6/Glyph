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

func statusItem(width int, label string, color core.Color) *widgets.Bordered {
	bounds := geom.NewBounds(0, 0, width, 3)

	box, err := widgets.NewBox(bounds, widgets.BoxConfig{
		Padding: 1,
		Style:   framework.Style{Bg: core.Transparent, Fg: color},
		BorderConfig: widgets.BorderConfig{
			Thickness:   1,
			BorderStyle: widgets.Rounded,
			Style:       framework.Style{Bg: core.Transparent, Fg: color},
		},
	})
	if err != nil {
		panic(err)
	}

	text, err := widgets.NewText(&geom.Point{X: 0, Y: 0}, widgets.TextConfig{
		Value: label,
		Fg:    color,
	})
	if err != nil {
		panic(err)
	}
	box.AddChild(text)

	return box
}

func focusDrillDemo() *widgets.FocusableBox {
	outer, err := widgets.NewFocusableBox(geom.NewBounds(2, 2, 24, 8), widgets.FocusableBoxConfig{
		Padding:      1,
		BorderConfig: widgets.DefaultBorderConfig(),
		Style:        framework.Style{Bg: core.Transparent, Fg: core.White},
		FocusStyle:   framework.Style{Bg: core.Transparent, Fg: core.Red},
	})
	if err != nil {
		panic(err)
	}

	inner, err := widgets.NewFocusableBox(geom.NewBounds(0, 0, 14, 4), widgets.FocusableBoxConfig{
		Padding:      1,
		BorderConfig: widgets.DefaultBorderConfig(),
		Style:        framework.Style{Bg: core.Transparent, Fg: core.Blue},
		FocusStyle:   framework.Style{Bg: core.Transparent, Fg: core.Red},
	})
	if err != nil {
		panic(err)
	}

	label, err := widgets.NewText(&geom.Point{X: 0, Y: 0}, widgets.TextConfig{
		Value: "Enter->drill",
		Fg:    core.White,
	})
	if err != nil {
		panic(err)
	}
	inner.AddChild(label)

	outer.AddChild(inner)
	return outer
}

func listDemo(width int) *widgets.List {
	list, err := widgets.NewList(geom.NewBounds(28, 2, width, 12), widgets.ListConfig{
		Style:       framework.Style{Bg: core.Transparent, Fg: core.Green},
		ItemPadding: 1,
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 3; i++ {
		row, err := list.AddItem(4)
		if err != nil {
			panic(err)
		}
		text, err := widgets.NewText(&geom.Point{X: 0, Y: 0}, widgets.TextConfig{
			Value: "row",
			Fg:    core.Green,
		})
		if err != nil {
			panic(err)
		}
		row.AddChild(text)
	}

	return list
}

func clockWindow() *widgets.Window {
	win, err := widgets.NewWindow(geom.NewBounds(0, 0, 30, 10), widgets.WindowConfig{
		Padding:  1,
		BoxStyle: framework.Style{Bg: core.Transparent, Fg: core.LightBlue},
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
