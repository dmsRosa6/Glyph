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
			Padding: 3,
			BorderConfig: canvas.BorderConfig{
				Thickness: 1,
				BorderStyle:     canvas.DoubleLine,
				Style: canvas.Style{Fg:core.Gray, Bg: core.Transparent},
			},
		},
		Title:         "Background",
		TitleXOffset:  0,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.DarkGray,
		//Anchor: canvas.Anchor{canvas.Start,canvas.Start},
	}

	winPoint := geom.NewBounds(0, 0, 20, 10)

	bgWin, _ := canvas.NewWindow(winPoint, bgCfg)
	
	c.AddShape(bgWin)

	rec,_ := canvas.NewRect(
		geom.NewBounds(1,1,2,2),
		canvas.RectConfig{
			Ch: '*',
			Style: canvas.Style{
				Fg: core.Transparent,
				Bg: core.Red,
			},
		},
	)

	bgWin.AddChild(rec)

	go r.Run(c)

	fmt.Scanln()
	
	r.Stop()
}
