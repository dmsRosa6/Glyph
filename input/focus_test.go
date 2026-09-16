package input

import (
	"testing"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// stubFocusable is a minimal framework.Focusable for exercising
// FocusManager without any real widget machinery.
type stubFocusable struct {
	name    string
	focused bool
}

func (s *stubFocusable) Draw(buf *core.Buffer, vec geom.Vector) {}
func (s *stubFocusable) IsInBounds(parent geom.Bounds) bool     { return true }
func (s *stubFocusable) SetLayer(l int) error                   { return nil }
func (s *stubFocusable) GetLayer() int                          { return 0 }
func (s *stubFocusable) SetParentStyle(style *framework.Style)  {}
func (s *stubFocusable) SetContext(ctx framework.AppContext)    {}
func (s *stubFocusable) HandleInput(ev framework.Event) (bool, error) {
	return false, nil
}
func (s *stubFocusable) Focus()          { s.focused = true }
func (s *stubFocusable) Blur()           { s.focused = false }
func (s *stubFocusable) IsFocused() bool { return s.focused }
func (s *stubFocusable) ID() string      { return s.name }

// stubFocusContainer additionally implements framework.FocusContainer,
// so FocusManager.Enter can drill into it.
type stubFocusContainer struct {
	stubFocusable
	children []framework.Focusable
}

func (s *stubFocusContainer) FocusableChildren() []framework.Focusable {
	return s.children
}

func newStubs(names ...string) []framework.Focusable {
	out := make([]framework.Focusable, len(names))
	for i, n := range names {
		out[i] = &stubFocusable{name: n}
	}
	return out
}

func TestNewFocusManagerAutoFocusesFirstChild(t *testing.T) {
	children := newStubs("a", "b", "c")
	NewFocusManager(children, framework.Logger{})

	if !children[0].IsFocused() {
		t.Error("expected the first child to be focused on construction")
	}
	if children[1].IsFocused() || children[2].IsFocused() {
		t.Error("expected only the first child to be focused")
	}
}

func TestFocusManagerNewWithNoChildrenDoesNotPanic(t *testing.T) {
	NewFocusManager(nil, framework.Logger{})
}

func TestFocusManagerNextPrevWrap(t *testing.T) {
	children := newStubs("a", "b", "c")
	m := NewFocusManager(children, framework.Logger{})

	if got := m.Current(); got != children[0] {
		t.Fatalf("Current() = %v, want children[0]", got)
	}

	m.Next()
	if got := m.Current(); got != children[1] {
		t.Fatalf("after Next(): Current() = %v, want children[1]", got)
	}
	if children[0].IsFocused() {
		t.Error("expected children[0] to be blurred after Next()")
	}
	if !children[1].IsFocused() {
		t.Error("expected children[1] to be focused after Next()")
	}

	m.Next()
	m.Next() // wraps: 0 -> 1 -> 2 -> 0
	if got := m.Current(); got != children[0] {
		t.Fatalf("Next() should wrap back to children[0], got %v", got)
	}

	m.Prev()
	if got := m.Current(); got != children[2] {
		t.Fatalf("Prev() from children[0] should wrap to children[2], got %v", got)
	}
}

func TestFocusManagerEnterOnNonFocusContainerIsNoop(t *testing.T) {
	children := newStubs("a", "b")
	m := NewFocusManager(children, framework.Logger{})

	if drilled := m.Enter(); drilled {
		t.Error("Enter() on a plain Focusable (not a FocusContainer) should return false")
	}
	if got := m.Current(); got != children[0] {
		t.Fatalf("Current() should be unchanged after a no-op Enter(), got %v", got)
	}
}

func TestFocusManagerEnterOnEmptyFocusContainerIsNoop(t *testing.T) {
	container := &stubFocusContainer{stubFocusable: stubFocusable{name: "empty"}}
	m := NewFocusManager([]framework.Focusable{container}, framework.Logger{})

	if drilled := m.Enter(); drilled {
		t.Error("Enter() on a FocusContainer with zero children should return false")
	}
}

// TestFocusManagerEnterDrillsWithoutBlurringOwner is the contract the
// original todo item specifically named as worth pinning down: drilling
// into a FocusContainer should NOT blur the container itself -- both
// the owner and its first child should read as focused simultaneously,
// which is what makes "outer box and inner box both light up" work.
func TestFocusManagerEnterDrillsWithoutBlurringOwner(t *testing.T) {
	inner := newStubs("inner-a", "inner-b")
	owner := &stubFocusContainer{
		stubFocusable: stubFocusable{name: "owner"},
		children:      inner,
	}
	m := NewFocusManager([]framework.Focusable{owner}, framework.Logger{})

	if !owner.IsFocused() {
		t.Fatal("expected owner to be focused before Enter()")
	}

	if drilled := m.Enter(); !drilled {
		t.Fatal("Enter() on a non-empty FocusContainer should return true")
	}

	if !owner.IsFocused() {
		t.Error("owner should STILL be focused after Enter() -- drilling in must not blur it")
	}
	if got := m.Current(); got != inner[0] {
		t.Fatalf("Current() after Enter() = %v, want the first inner child", got)
	}
	if !inner[0].IsFocused() {
		t.Error("expected the first inner child to be focused after Enter()")
	}
}

// TestFocusManagerExitPopsWithoutRefocusingOwner confirms Exit doesn't
// need to explicitly re-focus the owner -- it was never blurred by
// Enter in the first place (see the test above), so it's already
// exactly where Exit should leave it.
func TestFocusManagerExitPopsWithoutRefocusingOwner(t *testing.T) {
	inner := newStubs("inner-a", "inner-b")
	owner := &stubFocusContainer{
		stubFocusable: stubFocusable{name: "owner"},
		children:      inner,
	}
	m := NewFocusManager([]framework.Focusable{owner}, framework.Logger{})
	m.Enter()

	m.Exit()

	if inner[0].IsFocused() {
		t.Error("expected the inner child to be blurred after Exit()")
	}
	if !owner.IsFocused() {
		t.Error("expected owner to still be focused after Exit() -- it was never blurred, so nothing should have re-focused it either")
	}
	if got := m.Current(); got != owner {
		t.Fatalf("Current() after Exit() = %v, want owner", got)
	}
}

func TestFocusManagerExitAtRootIsNoop(t *testing.T) {
	children := newStubs("a", "b")
	m := NewFocusManager(children, framework.Logger{})

	m.Exit() // already at root -- nothing to pop

	if got := m.Current(); got != children[0] {
		t.Fatalf("Exit() at root should be a no-op, Current() = %v, want children[0]", got)
	}
}

// TestFocusManagerNestedDrillAndExit covers Enter twice / Exit twice
// unwinding correctly through two levels of nesting.
func TestFocusManagerNestedDrillAndExit(t *testing.T) {
	leaf := newStubs("leaf-a", "leaf-b")
	mid := &stubFocusContainer{
		stubFocusable: stubFocusable{name: "mid"},
		children:      leaf,
	}
	outer := &stubFocusContainer{
		stubFocusable: stubFocusable{name: "outer"},
		children:      []framework.Focusable{mid},
	}
	m := NewFocusManager([]framework.Focusable{outer}, framework.Logger{})

	if !m.Enter() {
		t.Fatal("expected first Enter() (outer -> mid) to succeed")
	}
	if got := m.Current(); got != mid {
		t.Fatalf("after first Enter(): Current() = %v, want mid", got)
	}
	if !outer.IsFocused() {
		t.Error("outer should still be focused after drilling into it")
	}

	if !m.Enter() {
		t.Fatal("expected second Enter() (mid -> leaf-a) to succeed")
	}
	if got := m.Current(); got != leaf[0] {
		t.Fatalf("after second Enter(): Current() = %v, want leaf[0]", got)
	}
	if !mid.IsFocused() {
		t.Error("mid should still be focused after drilling into it")
	}
	if !outer.IsFocused() {
		t.Error("outer should still be focused two levels deep")
	}

	m.Exit()
	if got := m.Current(); got != mid {
		t.Fatalf("after first Exit(): Current() = %v, want mid", got)
	}
	if leaf[0].IsFocused() {
		t.Error("leaf[0] should be blurred after exiting its scope")
	}

	m.Exit()
	if got := m.Current(); got != outer {
		t.Fatalf("after second Exit(): Current() = %v, want outer", got)
	}
	if mid.IsFocused() {
		t.Error("mid should be blurred after exiting its scope")
	}
	if !outer.IsFocused() {
		t.Error("outer should still be focused back at the root")
	}
}
