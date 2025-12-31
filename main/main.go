package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

func main() {
	c := canvas.NewCanvas(60,30,*core.NewColor(0,0,0),*core.NewColor(255,255,255))	

	rect := canvas.NewRect(0,0,10,5,'X',true,*core.NewColor(100,100,100),*core.NewColor(20,50,100))

	c.Init()
	fmt.Scanln()
	c.AddShape(rect)

	var a int = 0
	for a < 10{
		c.Draw()
		rect.Translate(geom.Vector{1,1})
		fmt.Scanln()
		a++
	}

	c.Restore()
}
