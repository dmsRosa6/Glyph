package main

import (
	"time"

	canvas "github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/shape"
)

func main() {
    canvas := canvas.NewCanvas(40, 20)
    canvas.Init()
    defer canvas.Restore()

    rect := shape.NewRect(1, 1, 10, 10, 'A', true)
    canvas.AddShape(rect)

    canvas.Draw()

    time.Sleep(3 * time.Second)
}
