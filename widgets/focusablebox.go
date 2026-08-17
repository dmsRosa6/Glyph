package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// FocusableBox is Bordered plus focus behavior. Unlike Bordered/Window,
// it can't just embed *canvas.Container: it needs base.FocusableBaseNode
// for Focus/Blur/BindAction/HandleInput, and embedding BOTH
// FocusableBaseNode (which itself embeds BaseNode) and *canvas.Container
// (which also embeds BaseNode+Propagator) at the same depth would make
// every BaseNode-derived method ambiguous -- Go embedding isn't virtual
// dispatch, and two same-depth sources of the same method name is a
// compile error, not a merge.
//
// So FocusableBox holds its box as a plain field instead, and its own
// SetParentStyle/SetInvalidator/SetLogChannel/SetLayer forward to that
// ONE field directly. That's also why no base.Propagator is needed here
// either, on top of the embedding conflict: Propagator's whole job is
// fanning out to N owned sub-drawables, and there's only one (box) to
// forward to -- a direct call already does that job.
//
// RECONSTRUCTED -- see the note atop bordered.go.
type FocusableBox struct {
	base.FocusableBaseNode
	box *Bordered
}

type FocusableBoxConfig struct {
	Padding      int
	BorderConfig BorderConfig
	Style        framework.Style
	FocusStyle   framework.Style
	Layer        int
	Anchor       framework.Anchor
}

func NewFocusableBox(bounds *geom.Bounds, cfg FocusableBoxConfig) (*FocusableBox, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, framework.Style{Bg: core.Transparent, Fg: core.Transparent}, cfg.Layer, "FocusableBox")
	if err != nil {
		return nil, err
	}

	box, err := NewBox(geom.NewBounds(0, 0, bounds.W, bounds.H), BoxConfig{
		Padding:      cfg.Padding,
		Style:        cfg.Style,
		BorderConfig: cfg.BorderConfig,
	})
	if err != nil {
		return nil, err
	}

	fb := &FocusableBox{
		FocusableBaseNode: base.NewFocusableBaseNode(bn),
		box:               box,
	}
	fb.FocusableBaseNode.SetFocusStyle(cfg.FocusStyle)

	return fb, nil
}

func (fb *FocusableBox) Draw(buf *core.Buffer, vec geom.Vector) {
	// Re-push the focus-resolved style every frame: Focus()/Blur() only
	// flip a bool and call Invalidate(), they never re-call
	// SetParentStyle, so this is where box actually picks up FocusStyle.
	resolved := fb.FocusableBaseNode.Style()
	fb.box.SetParentStyle(&resolved)

	pos := fb.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	fb.box.Draw(buf, v)
}

func (fb *FocusableBox) AddChild(child framework.Drawable) {
	fb.box.AddChild(child)
}

func (fb *FocusableBox) RemoveChild(target framework.Drawable) {
	fb.box.RemoveChild(target)
}

func (fb *FocusableBox) Children() []framework.Drawable {
	return fb.box.Children()
}

// FocusableChildren makes FocusableBox a framework.FocusContainer, so
// FocusManager.Enter() can drill into it (main.go's focusDrillDemo).
func (fb *FocusableBox) FocusableChildren() []framework.Focusable {
	var out []framework.Focusable
	for _, c := range fb.box.Children() {
		if f, ok := c.(framework.Focusable); ok {
			out = append(out, f)
		}
	}
	return out
}

func (fb *FocusableBox) SetParentStyle(s *framework.Style) {
	fb.FocusableBaseNode.SetParentStyle(s)
	resolved := fb.FocusableBaseNode.Style()
	fb.box.SetParentStyle(&resolved)
}

func (fb *FocusableBox) SetContext(ctx framework.AppContext) {
	fb.FocusableBaseNode.SetContext(ctx)
	fb.box.SetContext(ctx)
}

func (fb *FocusableBox) SetLayer(l int) error {
	return fb.FocusableBaseNode.SetLayer(l)
}
