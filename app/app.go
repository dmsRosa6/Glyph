package app

import (
	"errors"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/input"
	"github.com/dmsRosa6/glyph/render"
)

type AppSignal int

const (
	NOOP AppSignal = iota
	SIGTERM
)

type AppActionFunc func(app *App, ev framework.Event) (redraw bool, sig AppSignal, err error)

type AppConfig struct {
	Width, Height int
	Fg, Bg        *core.Color
	RenderMode    render.RenderMode
	AppEvents     map[framework.Key]AppActionFunc
}

type App struct {
	Canvas    *canvas.Canvas
	Renderer  *render.Renderer
	Input     *input.Manager
	Focus     *input.FocusManager
	appEvents map[framework.Key]AppActionFunc
}

func NewApp(cfg AppConfig) (*App, error) {
	if cfg.Width < 0 {
		return nil, errors.New("width is less than 0")
	}
	if cfg.Height < 0 {
		return nil, errors.New("height is less than 0")
	}

	bg := core.Transparent
	if cfg.Bg != nil {
		bg = *cfg.Bg
	}
	fg := core.Transparent
	if cfg.Fg != nil {
		fg = *cfg.Fg
	}

	c, err := canvas.NewCanvas(canvas.CanvasConfig{Width: cfg.Width, Height: cfg.Height, Fg: fg, Bg: bg})
	if err != nil {
		return nil, err
	}

	r := render.NewRenderer(cfg.RenderMode.Mode, cfg.RenderMode.Fps)

	in, err := input.NewManager()
	if err != nil {
		return nil, err
	}

	defaultEvents := cfg.AppEvents
	if defaultEvents == nil {
		defaultEvents = defaultGlobalActions()
	}
	return &App{Canvas: c, Renderer: r, Input: in, appEvents: defaultEvents}, nil
}

func defaultGlobalActions() map[framework.Key]AppActionFunc {
	return map[framework.Key]AppActionFunc{
		framework.KeyCtrlC: func(a *App, ev framework.Event) (bool, AppSignal, error) {
			return false, SIGTERM, nil
		},
		framework.KeyEnter: func(a *App, ev framework.Event) (bool, AppSignal, error) {
			if !a.Focus.Enter() {
				if f := a.Focus.Current(); f != nil {
					f.HandleInput(ev)
				}
			}
			return true, NOOP, nil
		},
		framework.KeyEsc: func(a *App, ev framework.Event) (bool, AppSignal, error) {
			a.Focus.Exit()
			return true, NOOP, nil
		},
		framework.KeyTab: func(a *App, ev framework.Event) (bool, AppSignal, error) {
			a.Focus.Next()
			return true, NOOP, nil
		},
	}
}

func (a *App) Run() {
	a.Renderer.Start(a.Canvas)
	a.Focus = input.NewFocusManager(a.Canvas.CollectFocusable())
	a.Input.Start()

	for ev := range a.Input.Events() {
		if f, ok := a.appEvents[ev.Key]; ok {
			reRender, sig, _ := f(a, ev)

			if sig == SIGTERM {
				a.Stop()
				return
			}

			if reRender {
				a.Renderer.RequestRedraw()
			}
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
