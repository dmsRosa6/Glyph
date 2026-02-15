package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	
	c := canvas.NewCanvas(120, 60, core.Black, core.White)
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
		Title:         "Tomas",
		TitleXOffset:  0,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.DarkGray,
		Anchor: canvas.Anchor{canvas.Center,canvas.Center},
	}

	bgCfg2 := canvas.WindowConfig{
		BoxConfig: canvas.BoxConfig{
			Padding: 1,
			BorderConfig: canvas.BorderConfig{
				Thickness: 1,
				BorderStyle:     canvas.DoubleLine,
				Style: canvas.Style{Fg:core.Gray, Bg: core.Transparent},
			},
		},
		Title:         "Diogo",
		TitleXOffset:  0,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.DarkGray,
		//Anchor: canvas.Anchor{canvas.Center,canvas.Center},
	}

	winPoint := geom.NewBounds(0, 0, 20, 10)

	bgWin, _ := canvas.NewWindow(winPoint, bgCfg)
	
	winPoint2 := geom.NewBounds(0, 0, 20, 10)

	bgWin2, _ := canvas.NewWindow(winPoint2, bgCfg2)

	rec,_ := canvas.NewRect(
		geom.NewBounds(0,0,2,1),
		canvas.RectConfig{
			Style: canvas.Style{
				Fg: core.Black,
				Bg: core.Red,
			},	
			Anchor: canvas.Anchor{canvas.Start,canvas.End},
		},
		
	)

	
	bgWin2.AddChild(rec)

	c.AddShape(bgWin)
	c.AddShape(bgWin2)

	go r.Run(c)

	fmt.Scanln()

	r.Stop()
}
