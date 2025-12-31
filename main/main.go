package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

func main() {

	c1 := *core.NewColor(255,100,100)
	c2 := *core.NewColor(100,100,255)
	c := canvas.NewCanvas(60,30,*core.NewColor(0,0,0),*core.NewColor(255,255,255))	

	rect := canvas.NewRect(0,0,10,5,' ',true, c1,c1)

	c.Init()
	fmt.Scanln()
	c.AddShape(rect)

	var a int = 0
	for a < 50{
		c.Draw()
		rect.Translate(geom.Vector{1,0})
		if rect.Fg == c1 {
			rect.Fg = c2
		}else{
			rect.Fg = c1
		}
		fmt.Scanln()
		a++
	}

	c.Restore()
}
