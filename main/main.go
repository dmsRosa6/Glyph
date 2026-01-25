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
				BorderStyle:     canvas.DoubleLine,
				Style: canvas.Style{Fg:core.Gray, Bg: core.Transparent},
			},
		},
		Title:         "Background",
		TitleXOffset:  0,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.DarkGray,
		//Anchor: canvas.Anchor{canvas.Start,canvas.Center},
	}

	winPoint := geom.NewBounds(0, 2, 20, 10)

	bgWin, _ := canvas.NewWindow(winPoint, bgCfg)
	
	rec,_ := canvas.NewRect(
		geom.NewBounds(0,0,2,2),
		canvas.RectConfig{
			Ch: '*',
			Style: canvas.Style{
				Fg: core.Black,
				Bg: core.Red,
			},	
			Anchor: canvas.Anchor{canvas.End,canvas.End},
		},
		
	)

	bgWin.AddChild(rec)

	c.AddShape(bgWin)

	go r.Run(c)

	fmt.Scanln()
	
	r.Stop()
}
