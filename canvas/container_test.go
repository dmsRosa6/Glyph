package canvas

import (
	"testing"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/primitive"
)

// TestContainerZeroScrollYShowsFirstChild is the baseline: with the
// default ScrollY (0), a stack's first child is what shows at the
// container's own top row.
func TestContainerZeroScrollYShowsFirstChild(t *testing.T) {
	c, err := NewContainer(geom.NewBounds(0, 0, 3, 1), ContainerConfig{Layout: framework.StackLayout{}})
	if err != nil {
		t.Fatal(err)
	}
	top, _ := primitive.NewRect(geom.NewBounds(0, 0, 3, 1), primitive.RectConfig{Style: framework.Style{Bg: core.Red}})
	bottom, _ := primitive.NewRect(geom.NewBounds(0, 0, 3, 1), primitive.RectConfig{Style: framework.Style{Bg: core.Blue}})
	c.AddChild(top)
	c.AddChild(bottom)

	buf := core.NewBuffer(3, 3, core.White, core.Black)
	c.Draw(buf, geom.Vector{})

	if got := buf.Get(0, 0).Bg; got != core.Red {
		t.Fatalf("row 0 Bg = %+v, want core.Red (first child, no scroll)", got)
	}
}

// TestContainerScrollYShiftsContentNotClip is the actual mechanism
// under test: a container with two stacked 1-row children and
// ScrollY=1 should show the SECOND child's color at its own top row
// (scrolled into view), while rows outside this container's own
// declared 1-row-tall frame stay at the buffer's untouched background
// regardless of ScrollY -- confirming the clip rectangle itself never
// moved, only the content drawn inside it did.
func TestContainerScrollYShiftsContentNotClip(t *testing.T) {
	c, err := NewContainer(geom.NewBounds(0, 0, 3, 1), ContainerConfig{Layout: framework.StackLayout{}})
	if err != nil {
		t.Fatal(err)
	}
	top, _ := primitive.NewRect(geom.NewBounds(0, 0, 3, 1), primitive.RectConfig{Style: framework.Style{Bg: core.Red}})
	bottom, _ := primitive.NewRect(geom.NewBounds(0, 0, 3, 1), primitive.RectConfig{Style: framework.Style{Bg: core.Blue}})
	c.AddChild(top)
	c.AddChild(bottom)

	c.SetScrollY(1)

	buf := core.NewBuffer(3, 3, core.White, core.Black)
	c.Draw(buf, geom.Vector{})

	if got := buf.Get(0, 0).Bg; got != core.Blue {
		t.Fatalf("row 0 (inside the container) Bg = %+v, want core.Blue (scrolled into view)", got)
	}
	for y := 1; y < 3; y++ {
		if got := buf.Get(0, y).Bg; got != core.Black {
			t.Fatalf("row %d Bg = %+v, want the untouched background core.Black -- the container's clip should not have moved", y, got)
		}
	}
}

func TestContainerScrollYDefaultsToZero(t *testing.T) {
	c, err := NewContainer(geom.NewBounds(0, 0, 3, 1), ContainerConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.ScrollY(); got != 0 {
		t.Fatalf("ScrollY() = %d, want 0 for a freshly-constructed Container", got)
	}
}
