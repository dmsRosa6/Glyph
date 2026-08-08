// canvas/focusablebox.go
package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type FocusableBox struct {
	FocusableBaseNode
	border  *Border
	content *Container
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
	base, err := newBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	innerW := bounds.W - 2*cfg.Padding
	innerH := bounds.H - 2*cfg.Padding
	content, err := NewContainer(geom.NewBounds(cfg.Padding, cfg.Padding, innerW, innerH), ContainerConfig{Layer: cfg.Layer})
	if err != nil {
		return nil, err
	}

	border, err := NewBorder(geom.NewBounds(0, 0, bounds.W, bounds.H), cfg.BorderConfig)
	if err != nil {
		return nil, err
	}

	fb := &FocusableBox{
		FocusableBaseNode: newFocusableBaseNode(base),
		border:            border,
		content:           content,
	}
	fs := cfg.FocusStyle
	fb.focusStyle = &fs // set directly here; add a setter if you want it public

	border.SetParentStyle(fb.ResolvedStyle())
	content.SetParentStyle(fb.ResolvedStyle())

	return fb, nil
}

func (fb *FocusableBox) Draw(buf *core.Buffer, vec geom.Vector) {
	v := geom.Vector{X: vec.X + fb.computedPos.X, Y: vec.Y + fb.computedPos.Y}
	fb.content.Layout(fb.LocalFrame())
	fb.content.Draw(buf, v)
	fb.border.SetParentStyle(fb.ResolvedStyle()) // picks up focus color each frame
	fb.border.Draw(buf, v)
}

func (fb *FocusableBox) AddChild(child framework.Drawable) { fb.content.AddChild(child) }

// FocusableChildren makes this box drillable by FocusManager.Enter().
func (fb *FocusableBox) FocusableChildren() []framework.Focusable {
	var out []framework.Focusable
	for _, c := range fb.content.children {
		if f, ok := c.(framework.Focusable); ok {
			out = append(out, f)
		}
	}
	return out
}