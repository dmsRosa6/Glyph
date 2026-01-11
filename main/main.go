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
				Style:     canvas.DoubleLine,
				Fg:        core.Gray,
				Bg:        core.Transparent,
			},
		},
		Title:         "Background",
		TitleXOffset:  0,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.DarkGray,
		Anchor: canvas.AnchorBottom,
	}

	winPoint := geom.NewBounds(2, 2, 50, 20)

	bgWin, _ := canvas.NewWindow(winPoint, bgCfg)
	
	
	rectCfg := canvas.RectConfig{
		Bg: core.Red,
	}

	rect, _ := canvas.NewRect(geom.NewBounds(0, 0, 2, 2), rectCfg)
	
	bgWin.AddChild(rect)
	
	c.AddShape(bgWin)

	go r.Run(c)
	fmt.Scanln()

	winPoint.Pos.AddVector(geom.Vector{5,0})

	r.RequestRedraw()

	fmt.Scanln()
	
	r.Stop()
}
