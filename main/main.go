package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
)

func main() {

	c := canvas.NewCanvas(60, 30, core.Black, core.White)
	r := render.NewRenderer(render.LoopMode(0), 30)

	bgCfg := canvas.WindowConfig{
		Bounds: geom.NewBounds(2, 2, 50, 20),
		Box: canvas.BoxConfig{
			Bounds:  geom.NewBounds(2, 2, 50, 20),
			Padding: 1,
			Border: canvas.BorderConfig{
				Bounds:    geom.NewBounds(2, 2, 50, 20),
				Thickness: 1,
				Style:     canvas.SingleLine,
				Fg:        core.Gray,
				Bg:        core.Transparent,
			},
		},
		Title:         "Background",
		TitleXOffset:  2,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.Gray,
	}

	bgWin, _ := canvas.NewWindow(bgCfg)
	bgWin.SetLayer(10)

	mainCfg := canvas.WindowConfig{
		Bounds: geom.NewBounds(8, 6, 35, 12),
		Box: canvas.BoxConfig{
			Bounds:  geom.NewBounds(8, 6, 35, 12),
			Padding: 1,
			Border: canvas.BorderConfig{
				Bounds:    geom.NewBounds(8, 6, 35, 12),
				Thickness: 1,
				Style:     canvas.ThickLine,
				Fg:        core.White,
				Bg:        core.Black,
			},
		},
		Title:         "Main Window",
		TitleXOffset:  2,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.White,
	}

	mainWin, _ := canvas.NewWindow(mainCfg)
	mainWin.SetLayer(20)

	popupCfg := canvas.WindowConfig{
		Bounds: geom.NewBounds(10, 12, 28, 8),
		Box: canvas.BoxConfig{
			Bounds:  geom.NewBounds(10, 12, 28, 8),
			Padding: 1,
			Border: canvas.BorderConfig{
				Bounds:    geom.NewBounds(10, 12, 28, 8),
				Thickness: 1,
				Style:     canvas.DoubleLine,
				Fg:        core.Yellow,
				Bg:        core.Black,
			},
		},
		Title:         "Modal Dialog",
		TitleXOffset:  2,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.Yellow,
		Layer: 100,
	}

	popupWin, _ := canvas.NewWindow(popupCfg)

	c.AddShape(popupWin)
	c.AddShape(bgWin)
	c.AddShape(mainWin)

	go r.Run(c)
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	r.Stop()
}
