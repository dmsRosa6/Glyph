package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	
	c := canvas.NewCanvas(60,30,core.NewColor(0,0,0), core.White)	

	
	box := canvas.NewWindow(20,10, 20, 10, core.Transparent,core.NewColor(200,30,100) ,core.Transparent, core.NewColor(200,30,100), canvas.Rounded, "Jorge Manuel Viana", 0, true, core.Black)

	r := render.NewRenderer(render.LoopMode(0), 30)
	
	c.AddShape(box)
	
	go r.Run(c)
	

	fmt.Scanln()

	r.Stop()
}
