package base

import (
	"github.com/dmsRosa6/glyph/framework"
)

type ActionFunc func(node *FocusableBaseNode, ev framework.Event) (bool, error)

type FocusableBaseNode struct {
	BaseNode
	actions    map[framework.Key]ActionFunc
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
		actions:  make(map[framework.Key]ActionFunc),
	}
}

func (f *FocusableBaseNode) SetFocusStyle(s framework.Style) {
	f.focusStyle = &s
}

func (f *FocusableBaseNode) BindAction(k framework.Key, fn ActionFunc) {
	f.actions[k] = fn
}

func (f *FocusableBaseNode) HandleInput(ev framework.Event) (bool, error) {
	fn, ok := f.actions[ev.Key]
	if !ok {
		return false, nil
	}
	refresh, err := fn(f, ev)
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
	f.Invalidate()
}

func (f *FocusableBaseNode) Blur() {
	if !f.focused {
		return
	}
	f.focused = false
	f.Invalidate()
}

func (f *FocusableBaseNode) IsFocused() bool {
	return f.focused
}
