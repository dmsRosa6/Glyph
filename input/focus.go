package input

import "github.com/dmsRosa6/glyph/framework"

type FocusScope struct {
	owner    framework.Focusable // nil for the root scope
	children []framework.Focusable
	index    int
}

type FocusManager struct {
    stack []*FocusScope
}

func NewFocusManager(root []framework.Focusable) *FocusManager {
	m := &FocusManager{stack: []*FocusScope{{children: root}}}
	if c := m.Current(); c != nil {
		c.Focus()
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
	s.children[s.index].Blur()
	s.index = (s.index + 1) % len(s.children)
	s.children[s.index].Focus()
}

func (m *FocusManager) Prev() {
	s := m.top()
	if s == nil || len(s.children) == 0 {
		return
	}
	s.children[s.index].Blur()
	s.index = (s.index - 1 + len(s.children)) % len(s.children)
	s.children[s.index].Focus()
}

// Enter drills into the current focused widget's children, if it has
// any. Returns false (no-op) if the current widget isn't a
// FocusContainer or has nothing to drill into -- callers should treat
// that as "let the widget's own bound action handle it instead."
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

	// Deliberately NOT blurring cur -- it stays visually focused as
	// "the container you're inside," while its first child also lights
	// up. That's what makes "outer box AND inner box both recolor" work.
	m.stack = append(m.stack, &FocusScope{owner: cur, children: children})
	children[0].Focus()
	return true
}

func (m *FocusManager) Exit() {
	if len(m.stack) <= 1 {
		return // already at root, nothing to pop
	}
	s := m.top()
	if c := s.children[s.index]; c != nil {
		c.Blur()
	}
	m.stack = m.stack[:len(m.stack)-1]
	// owner was never blurred on Enter, so no re-focus needed here --
	// it's still exactly where we left it.
}