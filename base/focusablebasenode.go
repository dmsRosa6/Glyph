package base

import "github.com/dmsRosa6/glyph/framework"

// FocusableBaseNode is the common case: a leaf widget that needs both
// ordinary node behavior (BaseNode) and focus behavior (FocusBehavior)
// and holds no children of its own -- Button is the example. Because
// FocusBehavior owns no BaseNode (see focusbehavior.go), embedding both
// here is unambiguous: they promote disjoint method sets, merged into
// one convenience type.
//
// Composites that ALSO need to hold children (Window, FocusableBox,
// ListRow) don't use this combo -- they embed BaseNode and
// FocusBehavior separately, because they need a third piece (a
// Container/Bordered field) wired in too, and that wiring is
// composite-specific. See those types.
type FocusableBaseNode struct {
	BaseNode
	FocusBehavior
}

func NewFocusableBaseNode(base BaseNode) FocusableBaseNode {
	return FocusableBaseNode{
		BaseNode:      base,
		FocusBehavior: NewFocusBehavior(base.Source()),
	}
}

// Style blends the focus tint over the plain BaseNode-resolved style.
// The one method this combo can't get for free -- see
// FocusBehavior.ResolveFocusStyle for why.
func (f *FocusableBaseNode) Style() framework.Style {
	return f.FocusBehavior.ResolveFocusStyle(f.BaseNode.Style())
}

// SetContext reaches both halves. This isn't resolving an ambiguity --
// BaseNode.SetContext and FocusBehavior.SetFocusContext are different
// names, so promotion alone would just silently pick BaseNode.SetContext
// and never call SetFocusContext at all (FocusBehavior.ctx would sit at
// its zero value forever, and every Focus()/Blur()/HandleInput() log or
// redraw would silently no-op). This override exists to make sure both
// get the call, not to pick a winner.
func (f *FocusableBaseNode) SetContext(ctx framework.AppContext) {
	f.BaseNode.SetContext(ctx)
	f.FocusBehavior.SetFocusContext(ctx, f.BaseNode.ID())
}
