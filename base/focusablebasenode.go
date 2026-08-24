package base

import (
	"github.com/dmsRosa6/glyph/framework"
)

type FocusableActionContext struct {
	node *FocusableBaseNode
	ev   framework.Event
}

type FocusableActionFunc func(action FocusableActionContext) (bool, error)

// Node returns the widget this action fired on. Exported so action
// functions defined outside package base -- the normal case; BindAction
// is how widgets/user code wires these up -- can actually reach it.
func (a FocusableActionContext) Node() *FocusableBaseNode {
	return a.node
}

// Event returns the input event that triggered this action.
func (a FocusableActionContext) Event() framework.Event {
	return a.ev
}

// Nodes gives an action function reach into the rest of the tree by ID,
// e.g. a button's action looking up and updating an unrelated Text
// widget elsewhere:
//
//	if d, ok := action.Nodes().Find("scoreLabel"); ok {
//	    if t, ok := d.(*widgets.Text); ok { t.SetValue("42") }
//	}
//
// Safe to call even if this node isn't attached to a running App yet.
func (a FocusableActionContext) Nodes() *framework.Registry {
	return a.node.Context().Nodes()
}

type FocusableBaseNode struct {
	BaseNode
	actions    map[framework.Key]FocusableActionFunc
	focused    bool
	focusStyle *framework.Style
}

func (f *FocusableBaseNode) Style() framework.Style {
	if f.focused && f.focusStyle != nil {
		return *framework.ResolveStyle(*f.focusStyle, f.BaseNode.Style())
	}
	return f.BaseNode.Style()
}

func NewFocusableBaseNode(base BaseNode) FocusableBaseNode {
	return FocusableBaseNode{
		BaseNode: base,
		actions:  make(map[framework.Key]FocusableActionFunc),
	}
}

func (f *FocusableBaseNode) SetFocusStyle(s framework.Style) {
	f.focusStyle = &s
}

func (f *FocusableBaseNode) BindAction(k framework.Key, fn FocusableActionFunc) {
	f.actions[k] = fn
}

func (f *FocusableBaseNode) HandleInput(ev framework.Event) (bool, error) {
	fn, ok := f.actions[ev.Key]
	if !ok {
		return false, nil
	}
	refresh, err := fn(FocusableActionContext{node: f, ev: ev})
	if err != nil {
		f.Logger().Warning(err)
	}
	if refresh {
		f.Invalidate()
	}
	return true, err
}

func (f *FocusableBaseNode) Focus() {
	if f.focused {
		return
	}
	f.focused = true
	f.Logger().Debug("focused")
	f.Invalidate()
}

func (f *FocusableBaseNode) Blur() {
	if !f.focused {
		return
	}
	f.focused = false
	f.Logger().Debug("blurred")
	f.Invalidate()
}

func (f *FocusableBaseNode) IsFocused() bool {
	return f.focused
}
