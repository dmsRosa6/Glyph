package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
)

func main() {
	
	c := canvas.NewCanvas(-1,-1,core.NewColor(0,0,0),core.NewColor(200,30,100))	

	t := canvas.NewText(2,3,"12",core.White, core.Transparent)

	border := canvas.NewBorder(0,1,6,5,'X',core.Transparent,core.Transparent)


	c.Init()
	fmt.Scanln()
	c.AddShape(t)
	c.AddShape(border)
	c.Draw()
	
	fmt.Scanln()
	c.Restore()

}
