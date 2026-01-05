package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	
	c := canvas.NewCanvas(50,30,core.NewColor(0,0,0), core.White)	

	b := canvas.NewBorderWithStyle(0, 1, 24, 12, 3, core.NewColor(200,30,100), core.Transparent, canvas.SingleLine)
	
	r := render.NewRenderer(render.LoopMode(0), 30)
	c.AddShape(b)
	
	go r.Run(c)
	

	fmt.Scanln()

	r.Stop()
}
