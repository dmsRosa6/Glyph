package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
)

func main() {

	fmt.Println("Press Enter to start...")
	fmt.Scanln()

	c := canvas.NewCanvas(60, 30, core.Black, core.White)
	r := render.NewRenderer(render.LoopMode(0), 30)

	bgCfg := canvas.WindowConfig{
		BoxConfig: canvas.BoxConfig{
			Padding: 1,
			BorderConfig: canvas.BorderConfig{
				Thickness: 1,
				Style:     canvas.DoubleLine,
				Fg:        core.Gray,
				Bg:        core.Transparent,
			},
		},
		Title:         "Background",
		TitleXOffset:  2,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.DarkGray,
	}

	bgWin, _ := canvas.NewWindow(geom.NewBounds(2, 2, 50, 20), bgCfg)
	
	
	rectCfg := canvas.RectConfig{
		Bg: core.Red,
	}
	

	rect, _ := canvas.NewRect(geom.NewBounds(4, 4, 2, 2), rectCfg)
	
	bgWin.AddChild(rect)
	
	c.AddShape(bgWin)

	c.Compose()

	go r.Run(c)
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	r.Stop()
}
