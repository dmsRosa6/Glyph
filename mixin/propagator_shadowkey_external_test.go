package mixin_test

import (
	"testing"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
	"github.com/dmsRosa6/glyph/primitive"
)

// TestTextInputOnSubmitShadowedByGlobalEnterIsWarned confirms the risk
// TextInputConfig.OnSubmit's own doc comment calls out is actually
// caught by mixin.Propagator's shadow-key check, not just documented:
// attaching a TextInput with OnSubmit set, under an AppContext that
// already has Enter bound globally (the app.NavActions() case), should
// log a Warning the moment it's tracked -- exactly what would tell a
// developer their OnSubmit is never going to fire, before they spend
// time wondering why pressing Enter does nothing.
func TestTextInputOnSubmitShadowedByGlobalEnterIsWarned(t *testing.T) {
	ti, err := primitive.NewTextInput(&geom.Point{}, 10, primitive.TextInputConfig{
		OnSubmit: func(string) (bool, error) { return false, nil },
	})
	if err != nil {
		t.Fatal(err)
	}

	logs := make(chan core.AppLog, 4)
	ctx := framework.AppContext{
		Logs:     logs,
		LogLevel: core.Warning,
		IsGlobalKey: func(b framework.Binding) bool {
			return b.Key == framework.KeyEnter // simulates app.NavActions()
		},
	}

	var p mixin.Propagator
	p.PropagateContext(ctx)
	p.Track(ti)

	select {
	case l := <-logs:
		if l.Severity() != core.Warning {
			t.Fatalf("severity = %v, want Warning", l.Severity())
		}
	default:
		t.Fatal("expected a shadow-key Warning when attaching a TextInput with OnSubmit under a global Enter binding, got none")
	}
}

// TestTextInputWithoutOnSubmitIsNotWarned is the control case: a
// TextInput built with no OnSubmit binds no action to Enter at all
// (see bindKeys), so it has nothing for the shadow-key check to warn
// about even under the exact same globally-bound-Enter AppContext.
func TestTextInputWithoutOnSubmitIsNotWarned(t *testing.T) {
	ti, err := primitive.NewTextInput(&geom.Point{}, 10, primitive.TextInputConfig{})
	if err != nil {
		t.Fatal(err)
	}

	logs := make(chan core.AppLog, 4)
	ctx := framework.AppContext{
		Logs:     logs,
		LogLevel: core.Warning,
		IsGlobalKey: func(b framework.Binding) bool {
			return b.Key == framework.KeyEnter
		},
	}

	var p mixin.Propagator
	p.PropagateContext(ctx)
	p.Track(ti)

	select {
	case l := <-logs:
		t.Fatalf("expected no warning with OnSubmit unset, got: %s", l.Reason())
	default:
	}
}
