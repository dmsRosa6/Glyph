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
		BoxConfig: canvas.BoxConfig{
			Padding: 1,
			BorderConfig: canvas.BorderConfig{
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

	bgWin, _ := canvas.NewWindow(geom.NewBounds(2, 2, 50, 20), bgCfg)
	bgWin.SetLayer(10)

	mainCfg := canvas.WindowConfig{
		BoxConfig: canvas.BoxConfig{
			Padding: 1,
			BorderConfig: canvas.BorderConfig{
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

	mainWin, _ := canvas.NewWindow(geom.NewBounds(8, 6, 35, 12), mainCfg)
	mainWin.SetLayer(20)

	c.AddShape(bgWin)
	c.AddShape(mainWin)

	go r.Run(c)
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	r.Stop()
}
