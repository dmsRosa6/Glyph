package mixin

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type ContainerLike interface {
	framework.Drawable
	framework.Composable
	framework.ChildrenLister
	Resize(w, h int)
}

type FocusableWrapper struct {
	Node
	FocusBehavior
	inner ContainerLike
}

func NewFocusableWrapper(bn Node, inner ContainerLike) FocusableWrapper {
	return FocusableWrapper{
		Node:          bn,
		FocusBehavior: NewFocusBehavior(bn.Source()),
		inner:         inner,
	}
}

func (w *FocusableWrapper) Style() framework.Style {
	return w.FocusBehavior.ResolveFocusStyle(w.Node.Style())
}

func (w *FocusableWrapper) Draw(buf *core.Buffer, vec geom.Vector) {
	resolved := w.Style()
	w.inner.SetParentStyle(&resolved)
	pos := w.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}
	w.inner.Draw(buf, v)
}

func (w *FocusableWrapper) AddChild(child framework.Drawable) {
	w.inner.AddChild(child)
}

func (w *FocusableWrapper) RemoveChild(target framework.Drawable) {
	w.inner.RemoveChild(target)
}

func (w *FocusableWrapper) Children() []framework.Drawable {
	return w.inner.Children()
}

func (w *FocusableWrapper) FocusableChildren() []framework.Focusable {
	var out []framework.Focusable
	for _, c := range w.inner.Children() {
		if f, ok := c.(framework.Focusable); ok {
			out = append(out, f)
		}
	}
	return out
}

func (w *FocusableWrapper) SetParentStyle(s *framework.Style) {
	w.Node.SetParentStyle(s)
	resolved := w.Style()
	w.inner.SetParentStyle(&resolved)
}

func (w *FocusableWrapper) SetContext(ctx framework.AppContext) {
	w.Node.SetContext(ctx)
	w.FocusBehavior.SetFocusContext(ctx, w.Node.ID())
	w.inner.SetContext(ctx)
}

func (w *FocusableWrapper) Resize(width, height int) {
	w.Node.Resize(width, height)
	w.inner.Resize(width, height)
}

func (w *FocusableWrapper) Inner() ContainerLike {
	return w.inner
}
