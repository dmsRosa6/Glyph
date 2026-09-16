package app

import (
	"testing"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/input"
	"github.com/dmsRosa6/glyph/render"
)

// dispatchStub is a minimal framework.Focusable whose HandleInput
// behavior is pluggable per test, with a call counter so a test can
// assert whether it was reached at all.
type dispatchStub struct {
	focused  bool
	handleFn func(ev framework.Event) (bool, error)
	calls    int
}

func (s *dispatchStub) Draw(buf *core.Buffer, vec geom.Vector) {}
func (s *dispatchStub) IsInBounds(parent geom.Bounds) bool     { return true }
func (s *dispatchStub) SetLayer(l int) error                   { return nil }
func (s *dispatchStub) GetLayer() int                          { return 0 }
func (s *dispatchStub) SetParentStyle(style *framework.Style)  {}
func (s *dispatchStub) SetContext(ctx framework.AppContext)    {}
func (s *dispatchStub) Focus()                                 { s.focused = true }
func (s *dispatchStub) Blur()                                  { s.focused = false }
func (s *dispatchStub) IsFocused() bool                        { return s.focused }
func (s *dispatchStub) HandleInput(ev framework.Event) (bool, error) {
	s.calls++
	if s.handleFn != nil {
		return s.handleFn(ev)
	}
	return false, nil
}

// newDispatchApp builds a real App (via NewApp, so every field is
// wired the way production code wires it) with file logging disabled,
// then manually attaches a FocusManager around stub -- the same
// "fixed width/height avoids needing a real tty" approach the original
// todo item called for, without ever calling Run() (which would start
// real goroutines and try to engage raw mode).
func newDispatchApp(t *testing.T, stub *dispatchStub) (*App, framework.AppContext) {
	t.Helper()

	a, err := NewApp(AppConfig{
		Width:          10,
		Height:         5,
		RenderMode:     render.OnDemandMode(),
		DisableFileLog: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	a.focus = input.NewFocusManager([]framework.Focusable{stub}, framework.Logger{})

	ctx := framework.AppContext{
		Invalidate: func() {},
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

	return a, ctx
}

// TestHandleEventStructuralKeyGlobalWinsOverWidget: a structural key
// (Tab) with BOTH a global binding and a focused widget that would
// also handle it -- the global binding must run, and the widget must
// NOT be called at all.
func TestHandleEventStructuralKeyGlobalWinsOverWidget(t *testing.T) {
	stub := &dispatchStub{handleFn: func(ev framework.Event) (bool, error) { return true, nil }}
	a, ctx := newDispatchApp(t, stub)

	globalCalls := 0
	a.appEvents[framework.Binding{Key: framework.KeyTab}] = func(ctx framework.AppContext, ev framework.Event) (bool, error) {
		globalCalls++
		return false, nil
	}

	a.handleEvent(ctx, framework.Event{Key: framework.KeyTab})

	if globalCalls != 1 {
		t.Errorf("global handler called %d times, want 1", globalCalls)
	}
	if stub.calls != 0 {
		t.Errorf("widget HandleInput called %d times, want 0 (global should have won)", stub.calls)
	}
}

// TestHandleEventStructuralKeyNoGlobalFallsToWidget: a structural key
// with no global binding falls through to the focused widget.
func TestHandleEventStructuralKeyNoGlobalFallsToWidget(t *testing.T) {
	stub := &dispatchStub{handleFn: func(ev framework.Event) (bool, error) { return true, nil }}
	a, ctx := newDispatchApp(t, stub)

	a.handleEvent(ctx, framework.Event{Key: framework.KeyEnter})

	if stub.calls != 1 {
		t.Errorf("widget HandleInput called %d times, want 1", stub.calls)
	}
}

// TestHandleEventNonStructuralKeyWidgetClaimsOverGlobal: a
// non-structural key (KeyRune) that the focused widget claims
// (returns handled=true) AND has a global binding too -- the widget
// wins, the global handler must NOT fire.
func TestHandleEventNonStructuralKeyWidgetClaimsOverGlobal(t *testing.T) {
	stub := &dispatchStub{handleFn: func(ev framework.Event) (bool, error) { return true, nil }}
	a, ctx := newDispatchApp(t, stub)

	globalCalls := 0
	a.appEvents[framework.Binding{Key: framework.KeyRune}] = func(ctx framework.AppContext, ev framework.Event) (bool, error) {
		globalCalls++
		return false, nil
	}

	a.handleEvent(ctx, framework.Event{Key: framework.KeyRune, Rune: 'x'})

	if stub.calls != 1 {
		t.Errorf("widget HandleInput called %d times, want 1", stub.calls)
	}
	if globalCalls != 0 {
		t.Errorf("global handler called %d times, want 0 (widget should have won)", globalCalls)
	}
}

// TestHandleEventNonStructuralKeyWidgetDeclinesFallsToGlobal: a
// non-structural key the widget declines (handled=false) falls
// through to the global binding.
func TestHandleEventNonStructuralKeyWidgetDeclinesFallsToGlobal(t *testing.T) {
	stub := &dispatchStub{handleFn: func(ev framework.Event) (bool, error) { return false, nil }}
	a, ctx := newDispatchApp(t, stub)

	globalCalls := 0
	a.appEvents[framework.Binding{Key: framework.KeyRune}] = func(ctx framework.AppContext, ev framework.Event) (bool, error) {
		globalCalls++
		return false, nil
	}

	a.handleEvent(ctx, framework.Event{Key: framework.KeyRune, Rune: 'x'})

	if stub.calls != 1 {
		t.Errorf("widget HandleInput called %d times, want 1", stub.calls)
	}
	if globalCalls != 1 {
		t.Errorf("global handler called %d times, want 1 (should fall through after widget declined)", globalCalls)
	}
}

// TestHandleEventNeitherClaimsIsNoop: nothing bound anywhere -- no
// panic, no effect.
func TestHandleEventNeitherClaimsIsNoop(t *testing.T) {
	stub := &dispatchStub{handleFn: func(ev framework.Event) (bool, error) { return false, nil }}
	a, ctx := newDispatchApp(t, stub)

	a.handleEvent(ctx, framework.Event{Key: framework.KeyRune, Rune: 'z'})

	if stub.calls != 1 {
		t.Errorf("widget HandleInput called %d times, want 1", stub.calls)
	}
}

// TestHandleEventMouseRoutesToMouseHandlerNotKeyBindings confirms the
// Kind-vs-Key zero-value invariant at the DISPATCH level (the decoder
// already covers it in input/manager_test.go): a mouse Event's Key
// field sits at its zero value, KeyRune. A global binding on plain
// KeyRune must NOT fire for a mouse event -- only the dedicated mouse
// handler should, and the focused widget's HandleInput must not be
// called at all for a mouse event.
func TestHandleEventMouseRoutesToMouseHandlerNotKeyBindings(t *testing.T) {
	stub := &dispatchStub{handleFn: func(ev framework.Event) (bool, error) { return true, nil }}
	a, ctx := newDispatchApp(t, stub)

	keyRuneGlobalCalls := 0
	a.appEvents[framework.Binding{Key: framework.KeyRune}] = func(ctx framework.AppContext, ev framework.Event) (bool, error) {
		keyRuneGlobalCalls++
		return false, nil
	}

	mouseCalls := 0
	a.mouseHandler = func(ctx framework.AppContext, ev framework.Event) (bool, error) {
		mouseCalls++
		if ev.Kind != framework.EventKindMouse {
			t.Errorf("mouse handler received Kind = %v, want EventKindMouse", ev.Kind)
		}
		return false, nil
	}

	a.handleEvent(ctx, framework.Event{Kind: framework.EventKindMouse, MouseAction: framework.MousePress})

	if mouseCalls != 1 {
		t.Errorf("mouse handler called %d times, want 1", mouseCalls)
	}
	if keyRuneGlobalCalls != 0 {
		t.Errorf("KeyRune global handler called %d times, want 0 -- a mouse event's zero-value Key must not be mistaken for a KeyRune keypress", keyRuneGlobalCalls)
	}
	if stub.calls != 0 {
		t.Errorf("widget HandleInput called %d times, want 0 -- mouse events don't go through the focused-widget path", stub.calls)
	}
}

// TestHandleEventMouseWithNoHandlerIsNoop: a mouse event with no
// mouse handler bound is dropped cleanly, no panic.
func TestHandleEventMouseWithNoHandlerIsNoop(t *testing.T) {
	stub := &dispatchStub{}
	a, ctx := newDispatchApp(t, stub)

	a.handleEvent(ctx, framework.Event{Kind: framework.EventKindMouse, MouseAction: framework.MousePress})

	if stub.calls != 0 {
		t.Errorf("widget HandleInput called %d times, want 0", stub.calls)
	}
}
