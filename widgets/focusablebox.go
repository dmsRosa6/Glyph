package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// FocusableBox is Bordered plus focus behavior. It embeds base.BaseNode
// and base.FocusBehavior side by side -- unambiguous now, since
// FocusBehavior owns no BaseNode of its own -- and still holds box as a
// plain field rather than embedding *Bordered.
//
// That field is NOT a leftover of the old ambiguity: this outer
// BaseNode's own style is deliberately Transparent, used only to
// receive whatever style FocusableBox's real parent hands it, so the
// focus tint can be blended in before that style reaches box. box's own
// Container has an unrelated BaseNode for its own bounds. Two BaseNodes
// serving two different purposes, not two BaseNodes fighting over one
// purpose -- FocusBehavior only ever fixed the second problem.
type FocusableBox struct {
	base.BaseNode
	base.FocusBehavior
	box *Bordered
}

type FocusableBoxConfig struct {
	Padding      int
	BorderConfig BorderConfig
	Style        framework.Style
	FocusStyle   *framework.Style
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
		BaseNode:      bn,
		FocusBehavior: base.NewFocusBehavior("FocusableBox"),
		box:           box,
	}
	if cfg.FocusStyle != nil {
		fb.SetFocusStyle(*cfg.FocusStyle)
	}

	return fb, nil
}

// Style blends the focus tint over this wrapper's own resolved style --
// what actually gets pushed down into box, see Draw.
func (fb *FocusableBox) Style() framework.Style {
	return fb.FocusBehavior.ResolveFocusStyle(fb.BaseNode.Style())
}

func (fb *FocusableBox) Draw(buf *core.Buffer, vec geom.Vector) {
	// Re-push the focus-resolved style every frame: Focus()/Blur() only
	// flip a bool and call Invalidate(), they never re-call
	// SetParentStyle, so this is where box actually picks up FocusStyle.
	resolved := fb.Style()
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
// FocusManager.Enter() can drill into it.
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
	fb.BaseNode.SetParentStyle(s)
	resolved := fb.Style()
	fb.box.SetParentStyle(&resolved)
}

func (fb *FocusableBox) SetContext(ctx framework.AppContext) {
	fb.BaseNode.SetContext(ctx)
	fb.FocusBehavior.SetFocusContext(ctx, fb.BaseNode.ID())
	fb.box.SetContext(ctx)
}
