package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

// Canvas owns the actual pixel buffer and treats the whole screen as a
// single full-size Container -- the root of the tree, not a fourth
// reimplementation of "add a child, check bounds, propagate style,
// propagate invalidator, iterate and draw." Every one of those
// responsibilities is exactly Container's job already; Canvas just
// delegates to it instead of keeping its own copy in sync by hand.
type Canvas struct {
	root *Container
	Buf  *core.Buffer

	RequestedWidth  int
	RequestedHeight int
}

// NewCanvas: default-substitution keys off core.Transparent, not the
// zero-value core.Color{} (which has the same field values as
// core.Black, so an explicitly-passed Black used to silently become
// White).
func NewCanvas(w, h int, fg, bg core.Color) *Canvas {
	if bg == core.Transparent {
		bg = core.White
	}
	if fg == core.Transparent {
		fg = core.Black
	}

	// NewContainer only errors if Layer < 0, and NewCanvas never passes
	// a caller-supplied layer here -- this can't actually fail.
	root, err := NewContainer(geom.NewBounds(0, 0, w, h), ContainerConfig{
		Style: Style{Bg: bg, Fg: fg},
	})
	if err != nil {
		panic(err)
	}

	return &Canvas{
		root:            root,
		Buf:             core.NewBuffer(w, h, fg, bg),
		RequestedWidth:  w,
		RequestedHeight: h,
	}
}

func (c *Canvas) ApplySize(termW, termH int) {
	w := c.RequestedWidth
	h := c.RequestedHeight

	if w <= 0 {
		w = termW
	}
	if h <= 0 {
		h = termH
	}

	actualW := min(termW, w)
	actualH := min(termH, h)

	s := c.root.Style()
	c.Buf = core.NewBuffer(actualW, actualH, s.Fg, s.Bg)
	c.Compose()
}

func (c *Canvas) Restore() {
	s := c.root.Style()
	c.Buf.Clear(s.Fg, s.Bg)
}

// AddShape delegates straight to the root Container's AddChild: same
// bounds check, same style propagation, same invalidator propagation,
// same layer-sorted insertion every nested Container already gives its
// children. Nothing about being "the top" needs its own version of this.
func (c *Canvas) AddShape(s Drawable) {
	c.root.AddChild(s)
}

// SetInvalidator wires up the redraw callback for the whole tree, called
// by Renderer.Run. Container.SetInvalidator already handles reaching
// children added before this was called (the normal order: build the
// tree, then hand it to a Renderer) -- Canvas doesn't need its own copy
// of that logic.
func (c *Canvas) SetInvalidator(fn func()) {
	c.root.SetInvalidator(fn)
}

// Shapes returns the top-level shapes currently on the canvas. This is
// deliberately a read-only accessor, not an exported slice -- the old
// Canvas.Shapes was a public field, which meant anyone could mutate it
// directly (`c.Shapes = append(...)`) and skip AddShape's bounds check,
// style propagation, and invalidator wiring entirely. That's the same
// class of bug as the old List building a Box by hand instead of going
// through NewBox: an escape hatch around the constructor that leaves the
// tree in a half-wired state.
func (c *Canvas) Shapes() []Drawable {
	out := make([]Drawable, len(c.root.children))
	copy(out, c.root.children)
	return out
}

func (c *Canvas) Compose() {
	c.Restore()
	c.root.Draw(c.Buf, geom.Vector{})
}