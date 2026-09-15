package app

import (
	"errors"
	"fmt"
	"runtime/debug"
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
	// Fg, Bg: core.Transparent means "let NewCanvas pick its own default
	// (white bg / black fg)". The zero value core.Color{} is opaque
	// black, not transparent -- see framework.StyleBg.
	Fg, Bg     core.Color
	RenderMode render.RenderMode
	// AppEvents is keyed on framework.Binding (Key + Modifiers), so e.g.
	// Tab and Shift+Tab can have different handlers. Ctrl+C is always
	// seeded separately below regardless of what's passed here.
	// NavActions() returns a ready-to-merge Tab/Shift+Tab/Enter/Esc set;
	// nothing here is included automatically.
	AppEvents map[framework.Binding]AppActionFunc
	// LogLevel defaults to core.Warning (its zero value), not Debug.
	LogLevel core.Severity
	// LogDir overrides the log directory (default "logs"). Ignored if
	// DisableFileLog is true.
	LogDir string
	// DisableFileLog skips file logging entirely.
	DisableFileLog bool
	// MouseEnabled turns on terminal mouse reporting. Off by default.
	MouseEnabled bool
	// InputBufferSize overrides the input event channel's capacity
	// (default input.DefaultEventBufferSize).
	InputBufferSize int
}

type App struct {
	Canvas *canvas.Canvas

	renderer *render.Renderer
	input    *input.Manager
	focus    *input.FocusManager

	// bindingsMu guards appEvents and mouseHandler -- both are
	// rebindable at runtime from any goroutine via Bind*/Unbind*.
	bindingsMu   sync.RWMutex
	appEvents    map[framework.Binding]AppActionFunc
	mouseHandler AppActionFunc

	logs       *fault.FaultManager
	logger     framework.Logger
	logLevel   core.Severity
	appSignals chan core.AppSignal
	nodes      *framework.Registry
	done       chan struct{}
	stopOnce   sync.Once
}

func NewApp(cfg AppConfig) (*App, error) {
	appSignals := make(chan core.AppSignal, 10)

	logs, err := fault.NewFaultManager(fault.Config{
		LogLevel:       cfg.LogLevel,
		LogDir:         cfg.LogDir,
		DisableFileLog: cfg.DisableFileLog,
	}, appSignals)
	if err != nil {
		return nil, fmt.Errorf("failed to create fault manager: %v", err)
	}

	if cfg.Width < 0 {
		return nil, errors.New("width is less than 0")
	}
	if cfg.Height < 0 {
		return nil, errors.New("height is less than 0")
	}

	c, err := canvas.NewCanvas(canvas.CanvasConfig{Width: cfg.Width, Height: cfg.Height, Fg: cfg.Fg, Bg: cfg.Bg})
	if err != nil {
		return nil, fmt.Errorf("failed to create canvas: %v", err)
	}

	r, err := render.NewRenderer(cfg.RenderMode, cfg.MouseEnabled, framework.NewLogger(logs.Logs(), cfg.LogLevel, string(core.RendererSource), ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %v", err)
	}

	in, err := input.NewManager(framework.NewLogger(logs.Logs(), cfg.LogLevel, string(core.InputSource), ""), cfg.InputBufferSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create input manager: %v", err)
	}

	// Ctrl+C always quits -- raw mode disables ISIG, so without this a
	// keyboard-only user has no way to exit. cfg.AppEvents layers on
	// top and can still redefine it.
	appEvents := map[framework.Binding]AppActionFunc{
		{Key: framework.KeyCtrlC}: QuitAction(),
	}
	for b, fn := range cfg.AppEvents {
		appEvents[b] = fn
	}

	return &App{
		Canvas:     c,
		renderer:   r,
		input:      in,
		appEvents:  appEvents,
		appSignals: appSignals,
		logs:       logs,
		logger:     framework.NewLogger(logs.Logs(), cfg.LogLevel, string(core.AppSource), ""),
		logLevel:   cfg.LogLevel,
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

// BindKey binds fn to k with no modifiers. Use BindKeyMod for a
// specific modifier combination.
func (a *App) BindKey(k framework.Key, fn AppActionFunc) {
	a.BindKeyMod(k, framework.ModNone, fn)
}

func (a *App) BindKeyMod(k framework.Key, mods framework.Modifier, fn AppActionFunc) {
	a.bindingsMu.Lock()
	defer a.bindingsMu.Unlock()
	if a.appEvents == nil {
		a.appEvents = make(map[framework.Binding]AppActionFunc)
	}
	a.appEvents[framework.Binding{Key: k, Modifiers: mods}] = fn
}

func (a *App) UnbindKey(k framework.Key) {
	a.UnbindKeyMod(k, framework.ModNone)
}

// BindMouse sets the single handler for every decoded mouse event.
// There's no per-widget mouse routing (needs hit-testing, not built
// yet) -- this is the only way application code receives mouse input.
func (a *App) BindMouse(fn AppActionFunc) {
	a.bindingsMu.Lock()
	defer a.bindingsMu.Unlock()
	a.mouseHandler = fn
}

func (a *App) UnbindMouse() {
	a.bindingsMu.Lock()
	defer a.bindingsMu.Unlock()
	a.mouseHandler = nil
}

func (a *App) UnbindKeyMod(k framework.Key, mods framework.Modifier) {
	a.bindingsMu.Lock()
	defer a.bindingsMu.Unlock()
	delete(a.appEvents, framework.Binding{Key: k, Modifiers: mods})
}

func (a *App) globalAction(b framework.Binding) (AppActionFunc, bool) {
	a.bindingsMu.RLock()
	defer a.bindingsMu.RUnlock()
	fn, ok := a.appEvents[b]
	return fn, ok
}

func (a *App) currentMouseHandler() AppActionFunc {
	a.bindingsMu.RLock()
	defer a.bindingsMu.RUnlock()
	return a.mouseHandler
}

// NavActions returns standard focus-navigation bindings: Tab/Shift+Tab
// cycle focus, Enter drills into a focused FocusContainer, Esc drills
// back out. Merge into AppConfig.AppEvents explicitly to use it.
func NavActions() map[framework.Binding]AppActionFunc {
	return map[framework.Binding]AppActionFunc{
		{Key: framework.KeyTab}: func(ctx framework.AppContext, ev framework.Event) (bool, error) {
			ctx.Nav().Next()
			return true, nil
		},
		{Key: framework.KeyTab, Modifiers: framework.ModShift}: func(ctx framework.AppContext, ev framework.Event) (bool, error) {
			ctx.Nav().Prev()
			return true, nil
		},
		{Key: framework.KeyEnter}: func(ctx framework.AppContext, ev framework.Event) (bool, error) {
			drilled := ctx.Nav().Enter()
			return drilled, nil
		},
		{Key: framework.KeyEsc}: func(ctx framework.AppContext, ev framework.Event) (bool, error) {
			ctx.Nav().Exit()
			return true, nil
		},
	}
}

func (a *App) Run() {
	a.logs.Start()

	a.focus = input.NewFocusManager(a.Canvas.CollectFocusable(), framework.NewLogger(a.logs.Logs(), a.logLevel, string(core.InputSource), ""))

	ctx := framework.AppContext{
		Logs:       a.logs.Logs(),
		Invalidate: a.renderer.RequestRedraw,
		Focus:      a.focus,
		Signal:     a.signal,
		Registry:   a.nodes,
		Done:       a.done,
		LogLevel:   a.logLevel,
		IsGlobalKey: func(b framework.Binding) bool {
			_, ok := a.globalAction(b)
			return ok
		},
	}
	a.Canvas.SetContext(ctx)

	a.renderer.Start(a.Canvas)

	if err := a.input.Start(); err != nil {
		a.logger.Warning(fmt.Errorf("failed to start input manager: %w", err))
		a.renderer.Stop()
		a.logs.Stop()
		return
	}

	a.logger.Info("App Started")

	for {
		select {
		case sig, ok := <-a.appSignals:
			if !ok {
				return
			}
			a.logger.Info(fmt.Sprintf("App signal '%s' received", sig.String()))
			if sig == core.SIGTERM {
				a.Stop()
				return
			}

		case ev, ok := <-a.input.Events():
			if !ok {
				return
			}
			a.handleEvent(ctx, ev)
		}
	}
}

// handleEvent dispatches one input event. Mouse events go to the
// global mouse handler if bound. Key events split on
// framework.IsStructuralKey: structural keys (Ctrl+C/Enter/Tab/Esc)
// dispatch global-first, everything else dispatches widget-first with
// global fallback -- see devguide.md's "app" section for why.
func (a *App) handleEvent(ctx framework.AppContext, ev framework.Event) {
	if ev.Kind == framework.EventKindMouse {
		if a.logger.Enabled(core.Debug) {
			a.logger.Debug(fmt.Sprintf("Mouse %s at (%d,%d)", ev.MouseAction.String(), ev.MouseX, ev.MouseY))
		}
		if handler := a.currentMouseHandler(); handler != nil {
			a.runGlobalAction(ctx, ev, handler)
		}
		return
	}

	if a.logger.Enabled(core.Debug) {
		a.logger.Debug(fmt.Sprintf("Key '%s' pressed", ev.Key.String()))
	}

	binding := framework.Binding{Key: ev.Key, Modifiers: ev.Modifiers}

	if framework.IsStructuralKey(ev.Key) {
		if fn, bound := a.globalAction(binding); bound {
			a.runGlobalAction(ctx, ev, fn)
			return
		}
		if f := a.focus.Current(); f != nil {
			f.HandleInput(ev)
		}
		return
	}

	if f := a.focus.Current(); f != nil {
		if handled, _ := f.HandleInput(ev); handled {
			return
		}
	}
	if fn, bound := a.globalAction(binding); bound {
		a.runGlobalAction(ctx, ev, fn)
	}
}

func (a *App) runGlobalAction(ctx framework.AppContext, ev framework.Event, fn AppActionFunc) {
	reRender, err := a.invokeAction(ctx, ev, fn)
	a.logger.Info(fmt.Sprintf("App event of key '%s' triggered. Re-render is '%t'", ev.Key.String(), reRender))
	if err != nil {
		a.logger.Warning(err)
	}
	if reRender {
		a.renderer.RequestRedraw()
	}
}

// invokeAction calls fn with a recover in place: a panic in a
// caller-supplied handler would otherwise crash the whole process
// mid-raw-mode. Recovered panics are logged as Fatal, which
// fault.FaultManager promotes to a clean shutdown.
func (a *App) invokeAction(ctx framework.AppContext, ev framework.Event, fn AppActionFunc) (redraw bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Fatal(fmt.Errorf("recovered panic in bound action for key %q: %v\n%s", ev.Key.String(), r, debug.Stack()))
			redraw, err = false, nil
		}
	}()
	return fn(ctx, ev)
}

// Stop can take over a second in the worst case (input.Manager.Stop
// alone can take ~100ms; FaultManager.Stop can wait on a 1s retry
// ticker). Not instant, but nothing here is wrong.
func (a *App) Stop() {
	a.stopOnce.Do(func() {
		close(a.done)
	})
	a.logger.Info("App Stopped")
	a.renderer.Stop()
	a.input.Stop()
	a.logs.Stop()
}
