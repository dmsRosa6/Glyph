package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/primitive"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/widgets"
)

const (
	gridOriginX = 2
	gridOriginY = 2
	gridW       = 120
	gridH       = 30
)

type paletteState struct {
	originX, originY int
	gridW, gridH     int
	palette          []core.Color
	paletteNames     []string
	colorIndex       int
}

func newPaletteState(originX, originY, gridW, gridH int) *paletteState {
	return &paletteState{
		originX:      originX,
		originY:      originY,
		gridW:        gridW,
		gridH:        gridH,
		palette:      []core.Color{core.Red, core.Orange, core.Yellow, core.Green, core.DodgerBlue, core.Purple, core.White},
		paletteNames: []string{"Red", "Orange", "Yellow", "Green", "Blue", "Purple", "White"},
	}
}

func (s *paletteState) currentColor() core.Color {
	return s.palette[s.colorIndex]
}

func (s *paletteState) currentColorName() string {
	return s.paletteNames[s.colorIndex]
}

func (s *paletteState) cycleColor(up bool) {
	if up {
		s.colorIndex = (s.colorIndex + 1) % len(s.palette)
	} else {
		s.colorIndex = (s.colorIndex - 1 + len(s.palette)) % len(s.palette)
	}
}

func (s *paletteState) gridCoords(mouseX, mouseY int) (x, y int, ok bool) {
	x = mouseX - s.originX
	y = mouseY - s.originY
	if x < 0 || y < 0 || x >= s.gridW || y >= s.gridH {
		return 0, 0, false
	}
	return x, y, true
}

func blankMatrix(w, h int, c core.Color) [][]core.Color {
	m := make([][]core.Color, h)
	for y := range m {
		m[y] = make([]core.Color, w)
		for x := range m[y] {
			m[y][x] = c
		}
	}
	return m
}

func main() {
	grid, err := widgets.NewTileGrid(&geom.Point{X: gridOriginX, Y: gridOriginY}, widgets.TileGridConfig{
		ColorMatrix: blankMatrix(gridW, gridH, core.DarkSlateGray),
	})
	if err != nil {
		panic(err)
	}

	state := newPaletteState(gridOriginX, gridOriginY, gridW, gridH)

	status, err := primitive.NewText(&geom.Point{X: gridOriginX, Y: gridOriginY + gridH + 1}, primitive.TextConfig{
		Value: fmt.Sprintf("Color: %s", state.currentColorName()),
		Fg:    core.White,
	})
	if err != nil {
		panic(err)
	}

	help, err := primitive.NewText(&geom.Point{X: gridOriginX, Y: gridOriginY + gridH + 2}, primitive.TextConfig{
		Value: "Left: paint   Right: erase   Scroll: change color   Ctrl+C: quit",
		Fg:    core.LightGray,
	})
	if err != nil {
		panic(err)
	}

	a, err := app.NewApp(app.AppConfig{
		Width:        gridOriginX*2 + gridW,
		Height:       gridOriginY + gridH + 4,
		Bg:           core.Black,
		RenderMode:   render.OnDemandMode(),
		MouseEnabled: true,
	})
	if err != nil {
		panic(err)
	}

	a.BindMouse(func(ctx framework.AppContext, ev framework.Event) (bool, error) {
		switch ev.MouseAction {
		case framework.MouseWheelUp, framework.MouseWheelDown:
			state.cycleColor(ev.MouseAction == framework.MouseWheelUp)
			status.SetValue(fmt.Sprintf("Color: %s", state.currentColorName()))
			return true, nil

		case framework.MousePress, framework.MouseDrag:
			x, y, ok := state.gridCoords(ev.MouseX, ev.MouseY)
			if !ok {
				return false, nil
			}
			switch ev.MouseButton {
			case framework.MouseButtonLeft:
				if err := grid.SetCell(x, y, state.currentColor()); err != nil {
					return false, err
				}
				return true, nil
			case framework.MouseButtonRight:
				if err := grid.SetCell(x, y, core.DarkSlateGray); err != nil {
					return false, err
				}
				return true, nil
			}
		}
		return false, nil
	})

	a.Canvas.AddShape(grid)
	a.Canvas.AddShape(status)
	a.Canvas.AddShape(help)

	a.Run()
}
