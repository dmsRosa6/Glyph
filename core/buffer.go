package core

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

func (b *Buffer) Get(x, y int) Cell {
	return b.cells[y*b.W+x]
}

func (b *Buffer) Cells() []Cell {
	return b.cells
}
