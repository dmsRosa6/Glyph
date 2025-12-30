package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	shape "github.com/dmsRosa6/glyph/geom"
)

func main() {


	canvas := canvas.NewCanvas(30,30)	

	rect := shape.NewRect(0,0,10,5,'X',true)

	canvas.Init()

	canvas.AddShape(rect)

	var a int = 0
	for a < 10{
		canvas.Draw()
		rect.Translate(shape.Vector{1,1})
		fmt.Scanln()
		a++
	}

	canvas.Restore()
}
