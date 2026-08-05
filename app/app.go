package app

import (
	"errors"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
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
	go a.Renderer.Run(a.Canvas)
    
	a.Input.Start()
	for ev := range a.Input.Events() {
        //filter out special ones
		if ev.Key == input.KeyCtrlC{
            a.Stop()
            return
        }

        if f, ok := any(a.Canvas).(canvas.Focusable); ok {
            reRender, _ := f.HandleInput(ev)
		
			if a.Renderer.Mode == render.LoopMode(0) && reRender{
				a.Renderer.RequestRedraw()
			}
		}
    }
}

func (a *App) Stop() {
    a.Renderer.Stop()
    a.Input.Stop()
}