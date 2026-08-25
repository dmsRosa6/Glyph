package base

import (
	"github.com/dmsRosa6/glyph/framework"
)

type FocusableActionContext struct {
	node *FocusableBaseNode
	ev   framework.Event
}

type FocusableActionFunc func(action FocusableActionContext) (bool, error)

func (a FocusableActionContext) Node() *FocusableBaseNode {
	return a.node
}

func (a FocusableActionContext) Event() framework.Event {
	return a.ev
}

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

// BoundKeys lists every key this node currently has an action bound
// to. Used by Propagator's shadow-warning check.
func (f *FocusableBaseNode) BoundKeys() []framework.Key {
	keys := make([]framework.Key, 0, len(f.actions))
	for k := range f.actions {
		keys = append(keys, k)
	}
	return keys
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
