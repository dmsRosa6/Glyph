package mixin

import (
	"testing"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
)

func drainHasFatal(logs chan core.AppLog) bool {
	for {
		select {
		case l := <-logs:
			if l.Severity() == core.Fatal {
				return true
			}
		default:
			return false
		}
	}
}

// TestFocusBehaviorHandleInputRecoversPanickingAction confirms a bound
// action that panics doesn't propagate out of HandleInput -- it should
// be recovered, reported as a Fatal-severity log (the same channel a
// normal returned error already goes through), and HandleInput should
// still report "this binding was found and invoked" (true) with a nil
// error, since the panic itself is what got reported, not something
// left for the caller to also handle a second time.
func TestFocusBehaviorHandleInputRecoversPanickingAction(t *testing.T) {
	fb := NewFocusBehavior("Test")
	logs := make(chan core.AppLog, 4)
	fb.SetFocusContext(framework.AppContext{Logs: logs, LogLevel: core.Warning}, "test#1")

	fb.BindAction(framework.KeyRune, func(FocusableActionContext) (bool, error) {
		panic("boom")
	})

	handled, err := fb.HandleInput(framework.Event{Key: framework.KeyRune, Rune: 'x'})
	if !handled {
		t.Fatal("expected handled=true (the binding was found and invoked) even though the action panicked")
	}
	if err != nil {
		t.Fatalf("expected err=nil after a recovered panic (already reported as Fatal), got: %v", err)
	}

	if !drainHasFatal(logs) {
		t.Fatal("expected a Fatal log after recovering the panic, got none")
	}
}

// TestFocusBehaviorHandleInputOrdinaryErrorStillWarns is the control
// case: a bound action that returns an ordinary error (no panic) still
// goes through the existing Warning path, unaffected by the new
// recover wrapping around it.
func TestFocusBehaviorHandleInputOrdinaryErrorStillWarns(t *testing.T) {
	fb := NewFocusBehavior("Test")
	logs := make(chan core.AppLog, 4)
	fb.SetFocusContext(framework.AppContext{Logs: logs, LogLevel: core.Warning}, "test#1")

	sentinel := &testError{"ordinary failure"}
	fb.BindAction(framework.KeyRune, func(FocusableActionContext) (bool, error) {
		return false, sentinel
	})

	handled, err := fb.HandleInput(framework.Event{Key: framework.KeyRune, Rune: 'x'})
	if !handled {
		t.Fatal("expected handled=true")
	}
	if err != sentinel {
		t.Fatalf("expected the ordinary error to be returned unchanged, got: %v", err)
	}

	select {
	case l := <-logs:
		if l.Severity() != core.Warning {
			t.Fatalf("severity = %v, want Warning", l.Severity())
		}
	default:
		t.Fatal("expected a Warning log for the ordinary error, got none")
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
