package app

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

// TestRunGlobalActionRecoversPanickingHandler confirms a global
// (App-level, BindKey/BindMouse-bound) action that panics doesn't
// propagate out of runGlobalAction, is reported as a Fatal-severity
// log, and -- since redraw is forced to false on a recovered panic --
// never touches a.renderer, so this is safe to call on an App built
// without a real Renderer wired in.
func TestRunGlobalActionRecoversPanickingHandler(t *testing.T) {
	logs := make(chan core.AppLog, 4)
	a := &App{
		logger: framework.NewLogger(logs, core.Debug, "App", ""),
	}

	panicky := AppActionFunc(func(ctx framework.AppContext, ev framework.Event) (bool, error) {
		panic("boom")
	})

	a.runGlobalAction(framework.AppContext{}, framework.Event{Key: framework.KeyRune, Rune: 'x'}, panicky)

	if !drainHasFatal(logs) {
		t.Fatal("expected a Fatal log after recovering a panic in a global action handler")
	}
}
