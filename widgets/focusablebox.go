// widgets/focusablebox.go
package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type FocusableBox struct {
	base.FocusableBaseNode
	border  *Border
	content *canvas.Container
}

type FocusableBoxConfig struct {
	BorderConfig BorderConfig
	Padding      int
	Style        framework.Style
	FocusStyle   framework.Style
	Anchor       framework.Anchor
	Layer        int
}

func NewFocusableBox(bounds *geom.Bounds, cfg FocusableBoxConfig) (*FocusableBox, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	innerW := bounds.W - 2*cfg.Padding
	innerH := bounds.H - 2*cfg.Padding
	content, err := canvas.NewContainer(geom.NewBounds(cfg.Padding, cfg.Padding, innerW, innerH), canvas.ContainerConfig{Layer: cfg.Layer})
	if err != nil {
		return nil, err
	}

	border, err := NewBorder(geom.NewBounds(0, 0, bounds.W, bounds.H), cfg.BorderConfig)
	if err != nil {
		return nil, err
	}

	fb := &FocusableBox{
		FocusableBaseNode: base.NewFocusableBaseNode(bn),
		border:            border,
		content:           content,
	}
	fb.SetFocusStyle(cfg.FocusStyle)

	border.SetParentStyle(fb.ResolvedStyle())
	content.SetParentStyle(fb.ResolvedStyle())

	return fb, nil
}

func (fb *FocusableBox) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := fb.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	fb.content.Layout(fb.LocalFrame())
	fb.content.Draw(buf, v)
	fb.border.SetParentStyle(fb.ResolvedStyle()) // picks up focus color each frame
	fb.border.Draw(buf, v)
}

func (fb *FocusableBox) AddChild(child framework.Drawable) { fb.content.AddChild(child) }

// FocusableChildren makes this box drillable by FocusManager.Enter().
func (fb *FocusableBox) FocusableChildren() []framework.Focusable {
	var out []framework.Focusable
	for _, c := range fb.content.Children() {
		if f, ok := c.(framework.Focusable); ok {
			out = append(out, f)
		}
	}
	return out
}

func (fb *FocusableBox) SetParentStyle(s *framework.Style) {
	fb.BaseNode.SetParentStyle(s)
	fb.border.SetParentStyle(fb.ResolvedStyle())
	fb.content.SetParentStyle(fb.ResolvedStyle())
}

func (fb *FocusableBox) SetInvalidator(fn func()) {
	fb.BaseNode.SetInvalidator(fn)
	fb.border.SetInvalidator(fn)
	fb.content.SetInvalidator(fn)
}

func (fb *FocusableBox) SetLogChannel(ch chan<- core.AppLog) {
	fb.BaseNode.SetLogChannel(ch)
	fb.border.SetLogChannel(ch)
	fb.content.SetLogChannel(ch)
}
