package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

func checkerboard(w, h int) [][]core.Color {
	matrix := make([][]core.Color, h)
	for y := 0; y < h; y++ {
		matrix[y] = make([]core.Color, w)
		for x := 0; x < w; x++ {
			if (x+y)%2 == 0 {
				matrix[y][x] = core.DarkGray
			} else {
				matrix[y][x] = core.LightGray
			}
		}
	}
	return matrix
}

func main() {
	a, err := app.NewApp(app.AppConfig{
		Width:      40,
		Height:     20,
		RenderMode: render.RenderMode{Mode: render.OnDemand},
	})
	if err != nil {
		panic(err)
	}

	grid, err := widgets.NewTileGrid(&geom.Point{X: 2, Y: 2}, widgets.TileGridConfig{
		ColorMatrix: checkerboard(16, 8),
	})
	if err != nil {
		panic(err)
	}

	a.Canvas.AddShape(grid)

	a.Run()
}
