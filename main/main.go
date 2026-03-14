package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	c := canvas.NewCanvas(80, 40, core.Black, core.White)
	r := render.NewRenderer(render.LoopMode(1), 60)
	boxCfg := canvas.BoxConfig{
		Padding: 1,
		BorderConfig: canvas.BorderConfig{
			Thickness:   1,
			BorderStyle: canvas.SingleLine,
			Style:       canvas.Style{Fg: core.Gray, Bg: core.Transparent},
		},
		Style: canvas.Style{Fg: core.White, Bg: core.Transparent},
		Anchor: canvas.Anchor{canvas.Start, canvas.Start},
	}

	mainBox, _ := canvas.NewBox(geom.NewBounds(0, 0, 30, 15), boxCfg)

	list, _ := canvas.NewList(canvas.ListConfig{
		Box:     mainBox,
		Style:   canvas.Style{Fg: core.White, Bg: core.Transparent},
		Anchor:  canvas.Anchor{canvas.Start, canvas.Start},
		Padding: 1,
		Layer:   0,
	})

	for i := 0; i < 3; i++ {
		itemBox, _ := canvas.NewBox(geom.NewBounds(0, 4*i, 30, 5), boxCfg)
		list.AddItem(itemBox)
	}
 
	c.AddShape(list)
	go r.Run(c)

	fmt.Scanln()
	r.Stop()

	fmt.Println("Done")
}
