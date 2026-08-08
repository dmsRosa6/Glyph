package app

import (
	"errors"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/input"
	"github.com/dmsRosa6/glyph/render"
)

type AppConfig struct {
	Bg         *core.Color
	RenderMode render.LoopMode
}

type App struct {
	Canvas   *canvas.Canvas
	Renderer *render.Renderer
	Input    *input.Manager
	Focus    *input.FocusManager
}

// NewApp creates a full-screen App: the Canvas always tracks the
// terminal's current size, growing and shrinking as the window is
// resized.
func NewApp(cfg AppConfig) (*App, error) {
	bg := core.Transparent
	if cfg.Bg != nil {
		bg = *cfg.Bg
	}

	c := canvas.NewCanvas(core.Transparent, bg)
	return newApp(c, cfg.RenderMode)
}

// NewFixedSizeApp creates an App whose Canvas is locked to exactly
// width x height (still capped at the terminal's current size if the
// terminal is smaller, but never growing past it if the terminal is
// larger).
func NewFixedSizeApp(width, height int, cfg AppConfig) (*App, error) {
	if width <= 0 {
		return nil, errors.New("width must be > 0")
	}
	if height <= 0 {
		return nil, errors.New("height must be > 0")
	}

	bg := core.Transparent
	if cfg.Bg != nil {
		bg = *cfg.Bg
	}

	c := canvas.NewFixedSizeCanvas(width, height, core.Transparent, bg)
	return newApp(c, cfg.RenderMode)
}

func newApp(c *canvas.Canvas, mode render.LoopMode) (*App, error) {
	r := render.NewRenderer(mode, 60)

	in, err := input.NewManager()
	if err != nil {
		return nil, err
	}

	return &App{Canvas: c, Renderer: r, Input: in}, nil
}

func (a *App) Run() {
	a.Renderer.Start(a.Canvas)
	a.Focus = input.NewFocusManager(a.Canvas.CollectFocusable())
	a.Input.Start()

	for ev := range a.Input.Events() {
		if ev.Key == framework.KeyCtrlC {
			a.Stop()
			return
		}

		if ev.Key == framework.KeyEnter {
			if !a.Focus.Enter() {
				if f := a.Focus.Current(); f != nil {
					f.HandleInput(ev)
				}
			}
			a.Renderer.RequestRedraw()
			continue
		}
		if ev.Key == framework.KeyEsc {
			a.Focus.Exit()
			a.Renderer.RequestRedraw()
			continue
		}

		if ev.Key == framework.KeyTab {
			a.Focus.Next()
			a.Renderer.RequestRedraw()
			continue
		}

		if f := a.Focus.Current(); f != nil {
			reRender, _ := f.HandleInput(ev)
			if reRender {
				a.Renderer.RequestRedraw()
			}
		}
	}
}

func (a *App) Stop() {
	a.Renderer.Stop()
	a.Input.Stop()
}