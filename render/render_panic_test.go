package render

import (
	"testing"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

// panickyDrawable is a minimal framework.Drawable whose Draw always
// panics -- standing in for a bug in some widget's own rendering code,
// framework-provided or user-defined.
type panickyDrawable struct {
	mixin.Node
}

func newPanickyDrawable(t *testing.T) *panickyDrawable {
	t.Helper()
	bn, err := mixin.NewNode(geom.NewBounds(0, 0, 1, 1), framework.Anchor{}, framework.Style{}, 0, "Panicky")
	if err != nil {
		t.Fatal(err)
	}
	return &panickyDrawable{Node: bn}
}

func (p *panickyDrawable) Draw(buf *core.Buffer, vec geom.Vector) {
	panic("boom: simulated Draw panic")
}

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

// TestRenderRecoversFromPanicInDraw confirms a panicking widget's Draw
// doesn't crash the renderer goroutine (and, per Go's per-process
// panic semantics, therefore the whole app mid-raw-mode) -- it should
// be recovered and reported as a Fatal-severity log instead. What
// fault.FaultManager does with that Fatal log (promote it to a clean
// SIGTERM shutdown via App.Stop()) is that package's own job, already
// covered by its own tests; this test only proves the renderer itself
// survives the panic and reports it correctly.
func TestRenderRecoversFromPanicInDraw(t *testing.T) {
	c, err := canvas.NewCanvas(canvas.CanvasConfig{Width: 5, Height: 5})
	if err != nil {
		t.Fatal(err)
	}
	c.AddShape(newPanickyDrawable(t))

	logs := make(chan core.AppLog, 4)
	logger := framework.NewLogger(logs, core.Debug, "Renderer", "")

	r, err := NewRenderer(OnDemandMode(), false, logger)
	if err != nil {
		t.Fatal(err)
	}

	// Must not panic out of this test.
	r.render(c)

	if !drainHasFatal(logs) {
		t.Fatal("expected a Fatal log after recovering a panic in Draw, got none")
	}
}
