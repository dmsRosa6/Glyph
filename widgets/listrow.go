package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type ListRow struct {
	base.BaseNode
	base.FocusBehavior
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
		BaseNode:      bn,
		FocusBehavior: base.NewFocusBehavior("ListRow"),
		content:       content,
	}, nil
}

func (r *ListRow) Style() framework.Style {
	return r.FocusBehavior.ResolveFocusStyle(r.BaseNode.Style())
}

func (r *ListRow) Draw(buf *core.Buffer, vec geom.Vector) {
	resolved := r.Style()
	r.content.SetParentStyle(&resolved)

	pos := r.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	r.content.Draw(buf, v)
}

// Resize resizes both halves ListRow actually owns: its own BaseNode
// and content, the inner Container doing the real drawing -- same
// reason and same shape as Window.Resize/FocusableBox.Resize. Not
// called anywhere internally today (List.AddItem sizes a row once, at
// construction), but left unfixed here it's the exact same trap the
// moment anything resizes a row later.
func (r *ListRow) Resize(w, h int) {
	r.BaseNode.Resize(w, h)
	r.content.Resize(w, h)
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
	r.BaseNode.SetParentStyle(s)
	resolved := r.Style()
	r.content.SetParentStyle(&resolved)
}

func (r *ListRow) SetContext(ctx framework.AppContext) {
	r.BaseNode.SetContext(ctx)
	r.FocusBehavior.SetFocusContext(ctx, r.BaseNode.ID())
	r.content.SetContext(ctx)
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
