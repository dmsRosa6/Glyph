package primitive

import (
	"errors"
	"sync"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

type TextInput struct {
	mixin.FocusableNode

	mu     sync.RWMutex
	value  []rune
	cursor int // 0..len(value)
}

type TextInputConfig struct {
	Value  string
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
	// OnSubmit, if set, is bound to Enter. Unlike Button.OnActivate
	// (bound to Space to avoid this), TextInput takes the shadow-key
	// risk deliberately.
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

// bindKeys wires the fixed set of editing keys. All go through
// BindAction (ModNone only), so a Ctrl+<letter> event is never claimed
// here and falls through to global bindings -- Ctrl+C still quits even
// while a TextInput is focused.
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

// SetValue replaces the entire content and moves the cursor to the
// end -- for loading text into the field or clearing it, not for
// per-keystroke editing (see bindKeys for that).
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

// scrollOffset picks the leftmost visible index so the cursor always
// stays within the width-wide window, recomputed fresh every Draw.
func scrollOffset(cursor, width int) int {
	if width <= 0 || cursor < width {
		return 0
	}
	return cursor - width + 1
}
