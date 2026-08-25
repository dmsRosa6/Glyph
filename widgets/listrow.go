package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type ListRow struct {
	base.FocusableBaseNode
	content *canvas.Container
}

func newListRow(bounds *geom.Bounds, anchor framework.Anchor) (*ListRow, error) {
	bn, err := base.NewBaseNode(bounds, anchor, framework.Style{Bg: core.Transparent, Fg: core.Transparent}, 0, "ListRow")
	if err != nil {
		return nil, err
	}

	content, err := canvas.NewContainer(geom.NewBounds(0, 0, bounds.W, bounds.H), canvas.ContainerConfig{
		Style: framework.Style{Bg: core.Transparent, Fg: core.Transparent},
	})
	if err != nil {
		return nil, err
	}

	return &ListRow{
		FocusableBaseNode: base.NewFocusableBaseNode(bn),
		content:           content,
	}, nil
}

func (r *ListRow) Draw(buf *core.Buffer, vec geom.Vector) {
	resolved := r.FocusableBaseNode.Style()
	r.content.SetParentStyle(&resolved)

	pos := r.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	r.content.Draw(buf, v)
}

func (r *ListRow) AddChild(child framework.Drawable) {
	r.content.AddChild(child)
}

func (r *ListRow) RemoveChild(target framework.Drawable) {
	r.content.RemoveChild(target)
}

func (r *ListRow) Children() []framework.Drawable {
	return r.content.Children()
}

func (r *ListRow) SetParentStyle(s *framework.Style) {
	r.FocusableBaseNode.SetParentStyle(s)
	resolved := r.FocusableBaseNode.Style()
	r.content.SetParentStyle(&resolved)
}

func (r *ListRow) SetContext(ctx framework.AppContext) {
	r.FocusableBaseNode.SetContext(ctx)
	r.content.SetContext(ctx)
}

func (r *ListRow) SetLayer(l int) error {
	return r.FocusableBaseNode.SetLayer(l)
}

func (r *ListRow) FocusableChildren() []framework.Focusable {
	var out []framework.Focusable
	for _, c := range r.content.Children() {
		if f, ok := c.(framework.Focusable); ok {
			out = append(out, f)
		}
	}
	return out
}
