package mixin

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// raceDrawable is a minimal framework.Drawable that also implements
// Identifiable, Raisable, and Stoppable -- so a concurrent stress test
// exercises every optional path Track/Untrack take (SetRaiser wiring,
// registry registration, Stop-on-removal), not just the plain Drawable
// contract. Its own layer field uses atomic load/store so any race the
// detector flags is genuinely about Propagator's own synchronization,
// not this stub's.
type raceDrawable struct {
	id    string
	layer int64
}

func (d *raceDrawable) Draw(buf *core.Buffer, vec geom.Vector) {}
func (d *raceDrawable) IsInBounds(parent geom.Bounds) bool     { return true }
func (d *raceDrawable) SetLayer(l int) error {
	atomic.StoreInt64(&d.layer, int64(l))
	return nil
}
func (d *raceDrawable) GetLayer() int                       { return int(atomic.LoadInt64(&d.layer)) }
func (d *raceDrawable) SetParentStyle(s *framework.Style)   {}
func (d *raceDrawable) SetContext(ctx framework.AppContext) {}
func (d *raceDrawable) ID() string                          { return d.id }
func (d *raceDrawable) SetRaiser(raise func())              {}
func (d *raceDrawable) Stop()                               {}

// TestPropagatorConcurrentAccess hammers a single shared Propagator
// from many goroutines doing every operation it exposes at once. Meant
// to be run with -race, which is what actually catches a bug here -- a
// plain pass/fail run proves nothing about the concurrency safety this
// test exists to check.
func TestPropagatorConcurrentAccess(t *testing.T) {
	var p Propagator

	const poolSize = 32
	pool := make([]framework.Drawable, poolSize)
	for i := range pool {
		pool[i] = &raceDrawable{id: string(rune('a' + i%26))}
	}
	pick := func(i int) framework.Drawable { return pool[i%poolSize] }

	ctx := framework.AppContext{}

	const workers = 8
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(workers * 7)

	for w := 0; w < workers; w++ {
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				p.Track(pick(seed + i))
			}
		}(w)
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				p.Untrack(pick(seed + i))
			}
		}(w)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = p.Children()
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = p.Count()
			}
		}()
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				p.BringToFront(pick(seed + i))
			}
		}(w)
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				p.SendToBack(pick(seed + i))
			}
		}(w)
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				if i%2 == 0 {
					p.PropagateStyle(&framework.Style{Bg: core.Red, Fg: core.White})
				} else {
					p.PropagateContext(ctx)
				}
			}
		}(w)
	}

	wg.Wait()
}
