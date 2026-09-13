package main

import (
	"fmt"

	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/primitive"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	a, err := app.NewApp(app.AppConfig{
		Width:      40,
		Height:     6,
		Bg:         core.Black,
		RenderMode: render.OnDemandMode(),
		// Deliberately NOT merging app.NavActions() here -- see
		// TextInputConfig.OnSubmit's own doc comment: NavActions binds
		// Enter globally, which would shadow this input's OnSubmit
		// entirely (and mixin.Propagator would log exactly that
		// warning the moment the input is attached, if you add it
		// back in and want to see it fire).
	})
	if err != nil {
		panic(err)
	}

	prompt, err := primitive.NewStaticText(&geom.Point{X: 1, Y: 1}, primitive.StaticTextConfig{
		Value: "Type something, Enter to submit:",
		Fg:    core.White,
	})
	if err != nil {
		panic(err)
	}

	status, err := primitive.NewText(&geom.Point{X: 1, Y: 3}, primitive.TextConfig{
		Value: "",
		Fg:    core.LightGray,
	})
	if err != nil {
		panic(err)
	}

	input, err := primitive.NewTextInput(&geom.Point{X: 1, Y: 2}, 30, primitive.TextInputConfig{
		Style: framework.Style{Bg: core.DarkSlateGray, Fg: core.White},
		OnSubmit: func(value string) (bool, error) {
			status.SetValue(fmt.Sprintf("You submitted: %q", value))
			return true, nil
		},
	})
	if err != nil {
		panic(err)
	}

	a.Canvas.AddShape(prompt)
	a.Canvas.AddShape(input)
	a.Canvas.AddShape(status)

	a.Run()
}
