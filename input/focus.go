package input

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
)

type FocusScope struct {
	owner    framework.Focusable // nil for the root scope
	children []framework.Focusable
	index    int
}

type FocusManager struct {
	stack []*FocusScope
	logs  chan<- core.AppLog
}

func NewFocusManager(root []framework.Focusable, logs chan<- core.AppLog) *FocusManager {
	m := &FocusManager{stack: []*FocusScope{{children: root}}, logs: logs}
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
	m.log(fmt.Sprintf("focus: drilled into %s, now %s", focusDesc(cur), focusDesc(children[0])))
	return true
}

func (m *FocusManager) Exit() {
	if len(m.stack) <= 1 {
		return // already at root, nothing to pop
	}
	s := m.top()
	var blurred framework.Focusable
	if c := s.children[s.index]; c != nil {
		c.Blur()
		blurred = c
	}
	owner := s.owner
	m.stack = m.stack[:len(m.stack)-1]
	// owner was never blurred on Enter, so no re-focus needed here --
	// it's still exactly where we left it.
	m.log(fmt.Sprintf("focus: exited %s, back to %s", focusDesc(blurred), focusDesc(owner)))
}

// focusDesc identifies a Focusable for a log line: its Registry id if
// it's Identifiable (every BaseNode-derived widget is, auto-generated
// as "<source>#<n>" unless overridden via SetID), or its Go type as a
// fallback for a hand-rolled Focusable that isn't built on BaseNode.
func focusDesc(f framework.Focusable) string {
	if f == nil {
		return "<none>"
	}
	if id, ok := f.(framework.Identifiable); ok {
		return id.ID()
	}
	return fmt.Sprintf("%T", f)
}

// log is a no-op if this FocusManager was built without a log channel
// (logs == nil), same nil-safe contract every other logging path in
// this codebase already follows.
func (m *FocusManager) log(msg string) {
	if m.logs == nil {
		return
	}
	m.logs <- *core.NewInfoAppLog(msg, string(core.InputSource))
}
