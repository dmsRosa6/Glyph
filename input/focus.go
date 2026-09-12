package input

import (
	"fmt"

	"github.com/dmsRosa6/glyph/framework"
)

type FocusScope struct {
	owner    framework.Focusable
	children []framework.Focusable
	index    int
}

type FocusManager struct {
	stack  []*FocusScope
	logger framework.Logger
}

func NewFocusManager(root []framework.Focusable, logger framework.Logger) *FocusManager {
	m := &FocusManager{stack: []*FocusScope{{children: root}}, logger: logger}
	if c := m.Current(); c != nil {
		c.Focus()
		m.log(fmt.Sprintf("focus: start at %s", focusDesc(c)))
	}
	return m
}

func (m *FocusManager) top() *FocusScope {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[len(m.stack)-1]
}

func (m *FocusManager) Current() framework.Focusable {
	s := m.top()
	if s == nil || len(s.children) == 0 {
		return nil
	}
	return s.children[s.index]
}

func (m *FocusManager) Next() {
	s := m.top()
	if s == nil || len(s.children) == 0 {
		return
	}
	old := s.children[s.index]
	old.Blur()
	s.index = (s.index + 1) % len(s.children)
	next := s.children[s.index]
	next.Focus()
	m.log(fmt.Sprintf("focus: %s -> %s", focusDesc(old), focusDesc(next)))
}

func (m *FocusManager) Prev() {
	s := m.top()
	if s == nil || len(s.children) == 0 {
		return
	}
	old := s.children[s.index]
	old.Blur()
	s.index = (s.index - 1 + len(s.children)) % len(s.children)
	prev := s.children[s.index]
	prev.Focus()
	m.log(fmt.Sprintf("focus: %s -> %s", focusDesc(old), focusDesc(prev)))
}

func (m *FocusManager) Enter() bool {
	cur := m.Current()
	fc, ok := cur.(framework.FocusContainer)
	if !ok {
		return false
	}
	children := fc.FocusableChildren()
	if len(children) == 0 {
		return false
	}
	m.stack = append(m.stack, &FocusScope{owner: cur, children: children})
	children[0].Focus()
	m.log(fmt.Sprintf("focus: drilled into %s, now %s", focusDesc(cur), focusDesc(children[0])))
	return true
}

func (m *FocusManager) Exit() {
	if len(m.stack) <= 1 {
		return
	}
	s := m.top()
	var blurred framework.Focusable
	if c := s.children[s.index]; c != nil {
		c.Blur()
		blurred = c
	}
	owner := s.owner
	m.stack = m.stack[:len(m.stack)-1]
	m.log(fmt.Sprintf("focus: exited %s, back to %s", focusDesc(blurred), focusDesc(owner)))
}

func focusDesc(f framework.Focusable) string {
	if f == nil {
		return "<none>"
	}
	if id, ok := f.(framework.Identifiable); ok {
		return id.ID()
	}
	return fmt.Sprintf("%T", f)
}

func (m *FocusManager) log(msg string) {
	m.logger.Debug(msg)
}
