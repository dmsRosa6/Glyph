package mixin

import "github.com/dmsRosa6/glyph/framework"

type FocusableNode struct {
	Node
	FocusBehavior
}

func NewFocusableNode(base Node) FocusableNode {
	return FocusableNode{
		Node:          base,
		FocusBehavior: NewFocusBehavior(base.Source()),
	}
}

func (f *FocusableNode) Style() framework.Style {
	return f.FocusBehavior.ResolveFocusStyle(f.Node.Style())
}

func (f *FocusableNode) SetContext(ctx framework.AppContext) {
	f.Node.SetContext(ctx)
	f.FocusBehavior.SetFocusContext(ctx, f.Node.ID())
}
