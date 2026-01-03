package main

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	
	c := canvas.NewCanvas(100,100,core.NewColor(0,0,0),core.NewColor(200,30,100))	

	t := canvas.NewText(2,3,"12",core.Black, core.White)

	c.AddShape(t)

	r := render.NewRenderer(render.LoopMode(0), 30)

	r.Run(c)
}
