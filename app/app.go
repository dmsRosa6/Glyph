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
	// AppEvents is keyed on framework.Binding (Key + Modifiers), not a
	// bare Key -- this is what lets Tab and Shift+Tab (or Ctrl+Right and
	// plain Right, etc.) carry different handlers instead of one
	// handler having to inspect ev.Modifiers itself. NavActions()
	// returns a ready-to-merge set covering standard Tab/Shift+Tab/
	// Enter/Esc navigation; nothing here is auto-included by NewApp
	// (except Ctrl+C, always seeded separately below) -- merge
	// NavActions() in explicitly if you want it.
	AppEvents map[framework.Binding]AppActionFunc
	LogLevel  core.Severity
	// MouseEnabled opts into terminal mouse-click/drag reporting (see
	// input.Manager's decoder and render.Renderer.Init). Off by default
	// -- enabling it changes what the terminal does with the mouse
	// system-wide for the duration of the app (e.g. ordinary text
	// selection by dragging stops working), which isn't something every
	// app wants turned on unconditionally.
	MouseEnabled bool
}

type App struct {
	Canvas *canvas.Canvas
	// renderer and input are unexported deliberately: both own a Stop()
	// that App.Stop() calls in a specific order (close(done), log, THEN
	// renderer.Stop(), THEN input.Stop()) -- a caller reaching
	// a.Renderer.Stop() or a.Input.Stop() directly could skip that
	// ordering entirely, or stop only one of the two and leave the app
	// half-shut-down with nothing to catch it. Canvas has no such
	// lifecycle method (see canvas.go -- AddShape/Shapes/SetContext/etc,
	// no Stop), so it stays exported: every example's
	// a.Canvas.AddShape(...) is the normal, everyday way to build a UI,
	// not a hazard.
	renderer  *render.Renderer
	input     *input.Manager
	focus     *input.FocusManager
	appEvents map[framework.Binding]AppActionFunc
	// mouseHandler is the one place a mouse event can go today -- see
	// BindMouse. There's no per-widget or per-Binding routing for mouse
	// (that needs hit-testing, deliberately not built -- see
	// framework.MouseHandler's doc comment), so this is a single global
	// callback, same shape as AppActionFunc, called for every decoded
	// mouse event if set.
	mouseHandler AppActionFunc
	logs         *fault.FaultManager
	appSignals   chan core.AppSignal
	nodes        *framework.Registry
	done         chan struct{}
	stopOnce     sync.Once
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

	r := render.NewRenderer(cfg.RenderMode, cfg.MouseEnabled, logs.Logs())

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

// BindKey binds fn to k with no modifiers (Binding{Key: k, Modifiers:
// framework.ModNone}). Use BindKeyMod for a specific modifier
// combination -- e.g. BindKeyMod(KeyTab, framework.ModShift, prevFn)
// alongside a plain BindKey(KeyTab, nextFn) for independent Tab /
// Shift+Tab behavior.
func (a *App) BindKey(k framework.Key, fn AppActionFunc) {
	a.BindKeyMod(k, framework.ModNone, fn)
}

// BindKeyMod binds fn to an exact (k, mods) combination. Exact match
// only -- see framework.Binding's doc comment.
func (a *App) BindKeyMod(k framework.Key, mods framework.Modifier, fn AppActionFunc) {
	if a.appEvents == nil {
		a.appEvents = make(map[framework.Binding]AppActionFunc)
	}
	a.appEvents[framework.Binding{Key: k, Modifiers: mods}] = fn
}

func (a *App) UnbindKey(k framework.Key) {
	a.UnbindKeyMod(k, framework.ModNone)
}

// BindMouse sets the single handler that receives every decoded mouse
// event (press/release/drag/wheel -- see framework.MouseAction). There's
// no per-widget or per-button routing here, only one global handler at
// a time -- calling BindMouse again replaces the previous one. This
// exists because there was otherwise NO way for application code to
// receive a mouse event at all: App.Run's dispatch loop only logs
// mouse events, deliberately never routing them through the Key-indexed
// appEvents/per-widget maps (see the dispatch loop's own comment for
// why). Doing real per-widget mouse dispatch needs hit-testing, which
// remains an open, separate design decision -- see
// framework.MouseHandler.
func (a *App) BindMouse(fn AppActionFunc) {
	a.mouseHandler = fn
}

func (a *App) UnbindMouse() {
	a.mouseHandler = nil
}

func (a *App) UnbindKeyMod(k framework.Key, mods framework.Modifier) {
	delete(a.appEvents, framework.Binding{Key: k, Modifiers: mods})
}

// NavActions returns a ready-to-merge set of standard focus-navigation
// bindings: Tab/Shift+Tab cycle focus forward/backward, Enter drills
// into a focused FocusContainer, Esc drills back out. None of this is
// auto-included by NewApp -- an app that wants it merges it in
// explicitly:
//
//	app.NewApp(app.AppConfig{AppEvents: app.NavActions()})
//
// or alongside other bindings:
//
//	events := app.NavActions()
//	events[framework.Binding{Key: framework.KeyCtrlC}] = myQuitHandler
//	app.NewApp(app.AppConfig{AppEvents: events})
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

	a.focus = input.NewFocusManager(a.Canvas.CollectFocusable(), a.logs.Logs())

	ctx := framework.AppContext{
		Logs:       a.logs.Logs(),
		Invalidate: a.renderer.RequestRedraw,
		Focus:      a.focus,
		Signal:     a.signal,
		Registry:   a.nodes,
		Done:       a.done,
		IsGlobalKey: func(b framework.Binding) bool {
			_, ok := a.appEvents[b]
			return ok
		},
	}
	a.Canvas.SetContext(ctx)

	a.renderer.Start(a.Canvas)

	err := a.input.Start()
	if err != nil {
		a.logs.Logs() <- *core.NewInfoAppLog("Failed to start input Manager", string(core.AppSource))
		a.renderer.Stop() // otherwise the terminal stays corrupted and the render goroutine leaks
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

		case ev, ok := <-a.input.Events():
			if !ok {
				return
			}

			if ev.Kind == framework.EventKindMouse {
				// Mouse decoding is real (input.Manager parses actual
				// SGR mouse escape sequences into MouseButton/
				// MouseAction/MouseX/MouseY). What's still NOT built is
				// per-widget dispatch (hit-testing: mapping MouseX/
				// MouseY to whichever Drawable's ABSOLUTE screen bounds
				// contain it) -- no part of this tree currently tracks
				// that (BaseNode only knows its position relative to
				// its own parent), and it's left as an open, separate
				// design decision rather than guessed at here -- see
				// framework.MouseHandler's doc comment.
				//
				// mouseHandler (see BindMouse) is the one place a
				// mouse event can go in the meantime: a single global
				// callback, not per-widget or per-Binding routing.
				//
				// Deliberately checked and handled BEFORE any of the
				// Key-indexed branches below: a mouse Event's Key field
				// sits at its zero value (KeyRune), and both the
				// per-widget actions map (FocusableBaseNode/
				// FocusBehavior) and a.appEvents are indexed by Key --
				// without this early return, every mouse click would be
				// silently indistinguishable from a KeyRune keypress
				// with Rune 0 to any widget or global binding on
				// KeyRune.
				a.logs.Logs() <- *core.NewInfoAppLog(fmt.Sprintf("Mouse %s at (%d,%d)", ev.MouseAction.String(), ev.MouseX, ev.MouseY), string(core.AppSource))
				if a.mouseHandler != nil {
					a.runGlobalAction(ctx, ev, a.mouseHandler)
				}
				continue
			}

			a.logs.Logs() <- *core.NewInfoAppLog(fmt.Sprintf("Key '%s' pressed", ev.Key.String()), string(core.AppSource))

			binding := framework.Binding{Key: ev.Key, Modifiers: ev.Modifiers}

			if framework.IsStructuralKey(ev.Key) {
				// Global-first, unconditionally, for Ctrl+C/Enter/Tab/Esc.
				if fn, bound := a.appEvents[binding]; bound {
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
			if fn, bound := a.appEvents[binding]; bound {
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
		a.renderer.RequestRedraw()
	}
}

func (a *App) Stop() {
	a.stopOnce.Do(func() {
		close(a.done)
	})
	a.logs.Logs() <- *core.NewInfoAppLog("App Stopped", string(core.AppSource))
	a.renderer.Stop()
	a.input.Stop()
	a.logs.Stop()
}
