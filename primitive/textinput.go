package primitive

import (
	"errors"
	"sync"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

// TextInput is a focusable, editable, single-line text field --
// composition shape 1's focusable variant (mixin.FocusableNode), same
// shape as Button. It deliberately does NOT embed mixin.TextBuffer:
// TextBuffer's SetValue replaces the whole value at once, which is the
// right shape for Text (a background goroutine pushing a brand-new
// string) but the wrong one for a field being edited one keystroke at
// a time -- a raw SetValue exposed here would leave the cursor
// pointing at a stale, possibly out-of-range index into content it no
// longer describes. Instead this holds its own private buffer plus a
// cursor index, mutated only through cursor-aware methods (insert,
// deleteBefore, deleteAfter, moveCursor, setCursor) -- the same reason
// MultilineText doesn't just expose its raw internal []rune either.
//
// Single-line only, deliberately, same scope limit as Text -- a value
// containing '\n' is accepted (nothing filters it out of what a caller
// can pass to SetValue) but drawn as a literal glyph, not a line
// break, exactly like Text does. Width is fixed at construction and
// the visible content SCROLLS horizontally to keep the cursor in view
// once typed content exceeds it -- there's no wrapping and no growing
// the field, matching MultilineText's "declared size, dev's problem if
// they overflow it" philosophy, just via scrolling instead of
// truncation, since dropping characters mid-typing (as truncation
// would) is actively hostile to someone who's still typing.
type TextInput struct {
	mixin.FocusableNode

	mu     sync.RWMutex
	value  []rune
	cursor int // index into value, 0..len(value) inclusive
}

type TextInputConfig struct {
	Value  string
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
	// OnSubmit, if set, is bound to Enter and receives the current
	// value. Unlike Button.OnActivate (deliberately bound to Space,
	// not Enter, specifically to dodge this) -- TextInput binds Enter
	// directly, since "Enter submits" is close to a universal
	// convention for a text field the way it isn't for a generic
	// button. That means the SAME shadow-key tradeoff Button's own doc
	// comment avoids applies here by choice, not by oversight: if an
	// app also merges app.NavActions() (which binds Enter globally to
	// drill into FocusContainers) or otherwise binds Enter itself,
	// this OnSubmit will never fire -- mixin.Propagator's own
	// warnShadowedKeys logs exactly that conflict the moment this
	// TextInput is attached to a tree, so it won't fail silently. nil
	// (the default) leaves Enter completely unbound, same
	// nothing-auto-included philosophy every other FocusBehavior-based
	// widget in this framework already follows.
	OnSubmit func(value string) (redraw bool, err error)
}

func NewTextInput(pos *geom.Point, width int, cfg TextInputConfig) (*TextInput, error) {
	if width < 1 {
		return nil, errors.New("text input width must be >= 1")
	}

	bounds := geom.NewBounds(pos.X, pos.Y, width, 1)
	bn, err := mixin.NewNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "TextInput")
	if err != nil {
		return nil, err
	}

	runes := []rune(cfg.Value)
	ti := &TextInput{
		FocusableNode: mixin.NewFocusableNode(bn),
		value:         runes,
		cursor:        len(runes),
	}

	ti.bindKeys(cfg.OnSubmit)
	return ti, nil
}

// bindKeys wires the fixed set of editing keys this widget understands
// -- called once, from the constructor, not exposed for a caller to
// rebind individually. Every binding here goes through BindAction
// (Binding{Key, ModNone}), never BindActionMod with a modifier: KeyRune
// bound this way will NOT match a Ctrl+<letter> event (see
// framework.Binding's own doc comment) -- that's deliberate, not a
// gap. It's what lets Ctrl+C keep quitting the app (or any other
// global Ctrl+<key> binding keep working) even while a TextInput is
// focused: a Ctrl-modified rune simply isn't claimed here, so it falls
// through to the global binding exactly like an unclaimed key always
// does in this framework's widget-first/global-fallback dispatch.
func (ti *TextInput) bindKeys(onSubmit func(string) (bool, error)) {
	ti.BindAction(framework.KeyRune, func(a mixin.FocusableActionContext) (bool, error) {
		ti.insert(a.Event().Rune)
		return true, nil
	})
	ti.BindAction(framework.KeyBackspace, func(a mixin.FocusableActionContext) (bool, error) {
		return ti.deleteBefore(), nil
	})
	ti.BindAction(framework.KeyDelete, func(a mixin.FocusableActionContext) (bool, error) {
		return ti.deleteAfter(), nil
	})
	ti.BindAction(framework.KeyLeft, func(a mixin.FocusableActionContext) (bool, error) {
		return ti.moveCursor(-1), nil
	})
	ti.BindAction(framework.KeyRight, func(a mixin.FocusableActionContext) (bool, error) {
		return ti.moveCursor(1), nil
	})
	ti.BindAction(framework.KeyHome, func(a mixin.FocusableActionContext) (bool, error) {
		return ti.setCursor(0), nil
	})
	ti.BindAction(framework.KeyEnd, func(a mixin.FocusableActionContext) (bool, error) {
		return ti.setCursor(ti.Len()), nil
	})
	if onSubmit != nil {
		ti.BindAction(framework.KeyEnter, func(a mixin.FocusableActionContext) (bool, error) {
			return onSubmit(ti.Value())
		})
	}
}

func (ti *TextInput) insert(r rune) {
	ti.mu.Lock()
	next := make([]rune, 0, len(ti.value)+1)
	next = append(next, ti.value[:ti.cursor]...)
	next = append(next, r)
	next = append(next, ti.value[ti.cursor:]...)
	ti.value = next
	ti.cursor++
	ti.mu.Unlock()
}

// deleteBefore is Backspace: removes the rune immediately before the
// cursor and moves the cursor back one. Reports false (no-op) at the
// start of the field -- FocusBehavior.HandleInput only redraws when
// the bound action reports true, so an edge-of-field Backspace with
// nothing to delete costs nothing beyond the lookup.
func (ti *TextInput) deleteBefore() bool {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	if ti.cursor == 0 {
		return false
	}
	ti.value = append(ti.value[:ti.cursor-1], ti.value[ti.cursor:]...)
	ti.cursor--
	return true
}

// deleteAfter is Delete: removes the rune at the cursor without moving
// it. Reports false at the end of the field, same reasoning as
// deleteBefore above.
func (ti *TextInput) deleteAfter() bool {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	if ti.cursor >= len(ti.value) {
		return false
	}
	ti.value = append(ti.value[:ti.cursor], ti.value[ti.cursor+1:]...)
	return true
}

func (ti *TextInput) moveCursor(delta int) bool {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	next := clamp(ti.cursor+delta, 0, len(ti.value))
	if next == ti.cursor {
		return false
	}
	ti.cursor = next
	return true
}

func (ti *TextInput) setCursor(pos int) bool {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	pos = clamp(pos, 0, len(ti.value))
	if pos == ti.cursor {
		return false
	}
	ti.cursor = pos
	return true
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// SetValue replaces the entire content programmatically -- loading
// existing text into the field, or clearing it from inside OnSubmit --
// and moves the cursor to the end, the same "cursor follows a full
// replacement" convention most editors use rather than leaving it at a
// now-arbitrary index into content it no longer describes. This is the
// one place outside the fixed key bindings that mutates the buffer;
// unlike those, it isn't routed through FocusBehavior's own redraw
// hook (there's no Event here to hand a bound action), so it calls
// Invalidate itself.
func (ti *TextInput) SetValue(v string) {
	ti.mu.Lock()
	ti.value = []rune(v)
	ti.cursor = len(ti.value)
	ti.mu.Unlock()
	ti.Invalidate()
}

func (ti *TextInput) Value() string {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	return string(ti.value)
}

func (ti *TextInput) Len() int {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	return len(ti.value)
}

func (ti *TextInput) Cursor() int {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	return ti.cursor
}

func (ti *TextInput) Draw(buf *core.Buffer, vec geom.Vector) {
	ti.mu.RLock()
	value := append([]rune(nil), ti.value...)
	cursor := ti.cursor
	ti.mu.RUnlock()

	w, _ := ti.Size()
	scroll := scrollOffset(cursor, w)

	s := ti.Style()
	pos := ti.ComputedPos()
	// IsFocused (from the embedded FocusBehavior) gates the cursor
	// glyph specifically -- an unfocused TextInput shows its content
	// with no cursor at all, the same way a plain text field looks
	// once you tab away from it.
	focused := ti.IsFocused()

	for col := 0; col < w; col++ {
		idx := scroll + col
		ch := ' '
		if idx < len(value) {
			ch = value[idx]
		}
		bg, fg := s.Bg, s.Fg
		if focused && idx == cursor {
			bg, fg = fg, bg // reverse-video cursor cell
		}
		buf.Set(vec.X+pos.X+col, vec.Y+pos.Y, ch, bg, fg)
	}
}

// scrollOffset picks the leftmost visible buffer index so the cursor
// always stays within the width-wide visible window, recomputed fresh
// on every Draw from (cursor, width) alone rather than persisted as
// its own field -- there is no "the user scrolled independently of the
// cursor" state in a plain text field, so nothing needs remembering
// between frames. cursor is allowed to equal the buffer's length (the
// "about to type past the last character" position); the window
// follows it there the same as any other position.
func scrollOffset(cursor, width int) int {
	if width <= 0 || cursor < width {
		return 0
	}
	return cursor - width + 1
}
