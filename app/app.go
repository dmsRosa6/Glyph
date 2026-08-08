package app

import (
	"errors"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/input"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/term"
)

type AppConfig struct {
    Width, Height int
    Bg        *core.Color
    RenderMode    render.LoopMode
}

type App struct {
    Canvas  *canvas.Canvas
    Renderer *render.Renderer
    Input    *input.Manager
	Focus    *input.FocusManager
}

func NewApp(cfg AppConfig) (*App, error) {
	size, err := term.TermSize()
	
	if err != nil {
		return nil, errors.New("could not retrieve terminal size")
	}


	if cfg.Width < 0 {
		return nil, errors.New("width is less than 0")
	}

	if cfg.Height < 0 {
		return nil, errors.New("height is less than 0")
	}


	
	w := size.Cols - 1
	h := size.Rows - 1
	
	if cfg.Width > 0 {
		w = cfg.Width
	}

	if cfg.Height > 0 {
		h  = cfg.Height
	}

	bg := core.Transparent

	if cfg.Bg != nil {
		bg = *cfg.Bg
	}


    c := canvas.NewCanvas(w, h, core.Transparent, bg)
    
	r := render.NewRenderer(cfg.RenderMode, 60)
    
	in, err := input.NewManager()
    if err != nil { return nil, err }
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
				// nothing to drill into -- let the current widget's own
				// bound action (if any) handle Enter itself
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