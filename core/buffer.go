package core

// Buffer is a W*H grid of terminal cells plus a clip-rect stack. cells
// is flat, row-major storage (index y*W+x) of plain Cell values, not
// [][]*Cell -- see Cell's doc comment for why. Clear/Set write values
// in place here; nothing in this package touches the heap per cell
// touched anymore. At 120x30 and 30 FPS, the old [][]*Cell shape meant
// Clear alone allocated a fresh *Cell for all 3600 cells on every
// single frame (~108k allocations/sec just to reset state) -- this
// shape allocates nothing after construction.
type Buffer struct {
	W, H  int
	cells []Cell

	Bg Color
	Fg Color

	clipStack []bufferClip
}

type bufferClip struct {
	x0, y0, x1, y1 int
}

func NewBuffer(w, h int, fg, bg Color) *Buffer {
	b := &Buffer{
		W:     w,
		H:     h,
		Fg:    fg,
		Bg:    bg,
		cells: make([]Cell, w*h),
	}
	b.ClearUsingDefaults()
	return b
}

func (b *Buffer) ClearUsingDefaults() {
	b.Clear(b.Fg, b.Bg)
}

// Clear overwrites every cell in place with a blank space in fg/bg --
// one Cell value computed once, then copied into every slot, rather
// than a fresh heap allocation per cell (the old NewCell(...) per
// cell, per Clear call).
func (b *Buffer) Clear(fg, bg Color) {
	blank := Cell{Ch: ' ', Fg: fg, Bg: bg}
	for i := range b.cells {
		b.cells[i] = blank
	}
}

func (b *Buffer) PushClip(x, y, w, h int) {
	nx0, ny0, nx1, ny1 := x, y, x+w, y+h

	if len(b.clipStack) > 0 {
		cur := b.clipStack[len(b.clipStack)-1]
		if cur.x0 > nx0 {
			nx0 = cur.x0
		}
		if cur.y0 > ny0 {
			ny0 = cur.y0
		}
		if cur.x1 < nx1 {
			nx1 = cur.x1
		}
		if cur.y1 < ny1 {
			ny1 = cur.y1
		}
	}

	b.clipStack = append(b.clipStack, bufferClip{x0: nx0, y0: ny0, x1: nx1, y1: ny1})
}

func (b *Buffer) PopClip() {
	if len(b.clipStack) == 0 {
		return
	}
	b.clipStack = b.clipStack[:len(b.clipStack)-1]
}

// Set writes one cell in place -- Cell{Ch: ch, Fg: fg, Bg: bg}
// constructed directly, no allocation, and (unlike the old
// NewCell(ch, fg, bg) this used to call) no argument-order swap: bg
// goes into Bg, fg goes into Fg, matching this method's own parameter
// names exactly. See Cell's doc comment for the bug this replaced.
func (b *Buffer) Set(x, y int, ch rune, bg, fg Color) {
	if y >= b.H || x >= b.W || y < 0 || x < 0 {
		return
	}

	if len(b.clipStack) > 0 {
		c := b.clipStack[len(b.clipStack)-1]
		if x < c.x0 || x >= c.x1 || y < c.y0 || y >= c.y1 {
			return
		}
	}

	b.cells[y*b.W+x] = Cell{Ch: ch, Fg: fg, Bg: bg}
}

// Get returns a copy of the cell at (x, y). Cell is a small value type
// (a rune plus two colors), so returning it by value rather than a
// pointer into internal storage is cheap and avoids exposing a pointer
// callers might be tempted to mutate through -- Set is the only
// supported way to write a cell.
func (b *Buffer) Get(x, y int) Cell {
	return b.cells[y*b.W+x]
}

// Cells returns the buffer's live, flat (row-major, index y*W+x) cell
// storage -- NOT a defensive copy, unlike base.Propagator.Children()'s
// copy-on-read convention elsewhere in this codebase. That copy exists
// to protect against concurrent mutation from other goroutines; a
// Buffer is only ever touched from the single render goroutine that
// owns one frame's whole compose-then-flush cycle, so there's no
// concurrent writer here to protect against, and a defensive copy would
// undo exactly the per-frame allocation this type exists to avoid.
// Callers must treat the result as read-only -- Set is still the only
// supported way to write a cell.
func (b *Buffer) Cells() []Cell {
	return b.cells
}
