package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	
	c := canvas.NewCanvas(60,30,core.NewColor(0,0,0), core.White)	

	r := render.NewRenderer(render.LoopMode(0), 30)
	
	winCfg := canvas.WindowConfig{
		Bounds: geom.NewBounds(5, 3, 40, 10), // x=5,y=3,w=40,h=10
		Box: canvas.BoxConfig{
			Bounds: geom.NewBounds(5, 3, 40, 10),
			Padding: 1,
			Border: canvas.BorderConfig{
				Bounds:    geom.NewBounds(5, 3, 40, 10),
				Thickness: 1,
				Style:     canvas.ThickLine,
				Fg:        core.White,
				Bg:        core.Black,
			},
		},
		Title:         "My Window",
		TitleXOffset:  2,
		TitlePosition: canvas.TitleTop,
		TitleFg:       core.White,
	}


	win := canvas.NewWindow(winCfg)

	c.AddShape(win)
	go r.Run(c)

	fmt.Scanln()

	r.Stop()
}
