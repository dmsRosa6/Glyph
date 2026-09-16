package mixin_test

import (
	"sync"
	"testing"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
	"github.com/dmsRosa6/glyph/primitive"
)

// TestFocusableWrapperConcurrentAddChild confirms a real
// FocusableWrapper-based composite (built the same way widgets.Window/
// FocusableBox/ListRow are) survives concurrent AddChild/Children from
// multiple goroutines under -race -- FocusableWrapper is newer code
// than the bare Propagator it forwards to, worth covering directly
// rather than only through the lower-level stress test in
// propagator_concurrent_test.go.
func TestFocusableWrapperConcurrentAddChild(t *testing.T) {
	bn, err := mixin.NewNode(geom.NewBounds(0, 0, 20, 10), framework.Anchor{}, framework.Style{}, 0, "Test")
	if err != nil {
		t.Fatal(err)
	}
	inner, err := canvas.NewContainer(geom.NewBounds(0, 0, 20, 10), canvas.ContainerConfig{})
	if err != nil {
		t.Fatal(err)
	}
	wrapper := mixin.NewFocusableWrapper(bn, inner)

	const workers = 8
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(workers * 2)

	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				r, err := primitive.NewRect(geom.NewBounds(0, 0, 1, 1), primitive.RectConfig{})
				if err != nil {
					t.Error(err)
					return
				}
				wrapper.AddChild(r)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = wrapper.Children()
			}
		}()
	}

	wg.Wait()
}
