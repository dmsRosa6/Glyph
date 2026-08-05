package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/input"
	"github.com/dmsRosa6/glyph/term"
)


type Canvas struct {
	root *Container
	Buf  *core.Buffer

	RequestedWidth  int
	RequestedHeight int
}

func NewMaxSizeCanvas(fg, bg core.Color) *Canvas {
	size, err := term.TermSize()
	if err != nil {
		panic("Could not retrieve terminal size")
	}

	return NewCanvas(size.Cols-1, size.Rows-1, fg, bg)
}

func NewCanvas(w, h int, fg, bg core.Color) *Canvas {
	if bg == core.Transparent {
		bg = core.White
	}
	if fg == core.Transparent {
		fg = core.Black
	}

	// NewContainer only errors if Layer < 0, and NewCanvas never passes
	root, err := NewContainer(geom.NewBounds(0, 0, w, h), ContainerConfig{
		Style: Style{Bg: bg, Fg: fg},
	})
	if err != nil {
		//TODO we dont need to panic move this
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

func (c *Canvas) HandleInput(e input.Event) (bool, error) {
	var m = map[input.Key] func() (bool, error){
    input.KeyEnter: func() (bool, error) {
		c.root.style = &Style{Fg: core.Red, Bg: core.DarkGray}
		
		return true, nil
	},
	}

	f := m[e.Key];

	if f != nil {
		
		reRender, err := f()
		if err != nil {
			return false, err
		}


		return reRender, nil
	}

	return false, nil;
}

