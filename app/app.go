package app

import (
	"errors"
	"fmt"
	"sync"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/fault"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/input"
	"github.com/dmsRosa6/glyph/render"
)

type AppActionFunc func(ctx framework.AppContext, ev framework.Event) (redraw bool, err error)

type AppConfig struct {
	Width, Height int
	Fg, Bg        *core.Color
	RenderMode    render.RenderMode
	AppEvents     map[framework.Key]AppActionFunc
	LogLevel      core.Severity
}

type App struct {
	Canvas     *canvas.Canvas
	Renderer   *render.Renderer
	Input      *input.Manager
	focus      *input.FocusManager
	appEvents  map[framework.Key]AppActionFunc
	logs       *fault.FaultManager
	appSignals chan core.AppSignal
	nodes      *framework.Registry
	done       chan struct{}
	stopOnce   sync.Once
}

func NewApp(cfg AppConfig) (*App, error) {

	appSignals := make(chan core.AppSignal, 10)

	logs, error := fault.NewFaultManager(cfg.LogLevel, appSignals)

	if error != nil {
		return nil, fmt.Errorf("failed to create fault manager: %v", error)
	}

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
		return nil, fmt.Errorf("failed to create canvas: %v", err)
	}

	r := render.NewRenderer(cfg.RenderMode.Mode, cfg.RenderMode.Fps, logs.Logs())

	in, err := input.NewManager(logs.Logs())
	if err != nil {
		return nil, fmt.Errorf("failed to create input manager: %v", err)
	}

	// Ctrl+C is the one binding every app gets unconditionally -- raw
	// mode (term.SafeRawMode, via Input.Start) disables ISIG, so the
	// terminal's own SIGINT never fires; without this, a keyboard-only
	// user would have zero way to exit. It's seeded here, then anything
	// in cfg.AppEvents is layered on top -- so a caller who explicitly
	// wants to redefine Ctrl+C still can, but simply not mentioning it
	// (the common case: animations, on-demand widgets, anything that
	// isn't building nav) costs nothing and needs no boilerplate.
	// Tab/Enter/Esc are NOT included here -- those stay fully opt-in via
	// NavActions(), merged in only by apps that actually want them.
	appEvents := map[framework.Key]AppActionFunc{
		framework.KeyCtrlC: QuitAction(),
	}
	for k, fn := range cfg.AppEvents {
		appEvents[k] = fn
	}

	return &App{
		Canvas:     c,
		Renderer:   r,
		Input:      in,
		appEvents:  appEvents,
		appSignals: appSignals,
		logs:       logs,
		nodes:      framework.NewRegistry(),
		done:       make(chan struct{}),
	}, nil
}

func (a *App) signal(sig core.AppSignal) {
	a.appSignals <- sig
}

func QuitAction() AppActionFunc {
	return func(ctx framework.AppContext, ev framework.Event) (bool, error) {
		ctx.SignalApp(core.SIGTERM)
		return false, nil
	}
}

func (a *App) BindKey(k framework.Key, fn AppActionFunc) {
	if a.appEvents == nil {
		a.appEvents = make(map[framework.Key]AppActionFunc)
	}
	a.appEvents[k] = fn
}

func (a *App) UnbindKey(k framework.Key) {
	delete(a.appEvents, k)
}

func (a *App) Run() {
	a.logs.Start()

	a.focus = input.NewFocusManager(a.Canvas.CollectFocusable(), a.logs.Logs())

	ctx := framework.AppContext{
		Logs:       a.logs.Logs(),
		Invalidate: a.Renderer.RequestRedraw,
		Focus:      a.focus,
		Signal:     a.signal,
		Registry:   a.nodes,
		Done:       a.done,
		IsGlobalKey: func(k framework.Key) bool {
			_, ok := a.appEvents[k]
			return ok
		},
	}
	a.Canvas.SetContext(ctx)

	a.Renderer.Start(a.Canvas)

	err := a.Input.Start()
	if err != nil {
		a.logs.Logs() <- *core.NewInfoAppLog("Failed to start input Manager", string(core.AppSource))
		a.Renderer.Stop() // otherwise the terminal stays corrupted and the render goroutine leaks
		return
	}

	a.logs.Logs() <- *core.NewInfoAppLog("App Started", string(core.AppSource))

	for {
		select {
		case sig, ok := <-a.appSignals:
			if !ok {
				return
			}
			a.logs.Logs() <- *core.NewInfoAppLog(fmt.Sprintf("App signal '%s' received", sig.String()), string(core.AppSource))
			if sig == core.SIGTERM {
				a.Stop()
				return
			}

		case ev, ok := <-a.Input.Events():
			if !ok {
				return
			}
			a.logs.Logs() <- *core.NewInfoAppLog(fmt.Sprintf("Key '%s' pressed", ev.Key.String()), string(core.AppSource))

			if framework.IsStructuralKey(ev.Key) {
				// Global-first, unconditionally, for Ctrl+C/Enter/Tab/Esc.
				if fn, bound := a.appEvents[ev.Key]; bound {
					a.runGlobalAction(ctx, ev, fn)
					continue
				}
				if f := a.focus.Current(); f != nil {
					f.HandleInput(ev)
				}
				continue
			}

			// Every other key: widget-first, global as fallback.
			if f := a.focus.Current(); f != nil {
				if handled, _ := f.HandleInput(ev); handled {
					continue
				}
			}
			if fn, bound := a.appEvents[ev.Key]; bound {
				a.runGlobalAction(ctx, ev, fn)
			}
		}
	}
}

func (a *App) runGlobalAction(ctx framework.AppContext, ev framework.Event, fn AppActionFunc) {
	reRender, err := fn(ctx, ev)
	a.logs.Logs() <- *core.NewInfoAppLog(fmt.Sprintf("App event of key '%s' triggered. Re-render is '%t'", ev.Key.String(), reRender), string(core.AppSource))
	if err != nil {
		a.logs.Logs() <- *core.NewWarningAppLog(err, string(core.AppSource))
	}
	if reRender {
		a.Renderer.RequestRedraw()
	}
}

func (a *App) Stop() {
	a.stopOnce.Do(func() {
		close(a.done)
	})
	a.logs.Logs() <- *core.NewInfoAppLog("App Stopped", string(core.AppSource))
	a.Renderer.Stop()
	a.Input.Stop()
	a.logs.Stop()
}
