package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	
	c := canvas.NewCanvas(1000,1000,core.NewColor(0,0,0),core.NewColor(200,30,100))	

	t := canvas.NewText(2,3,"12",core.Black, core.White)
	
	r := render.NewRenderer(render.LoopMode(0), 30)
	
	go r.Run(c)
	
	c.AddShape(t)

	fmt.Scanln()

	r.Stop()
}
