package mixin

import "sync"

// TextBuffer is a mutex-protected, atomically-replaceable []rune value
// -- the storage behind every mutable, single-line text-holding leaf.
// Pulled out of Text once a second consumer (the eventual TextInput,
// which needs this exact same "one []rune, replace it wholesale under
// a lock" shape plus cursor state on top) was about to duplicate it --
// same reasoning FocusBehavior was extracted for, back when three
// composites needed identical focus handling instead of three
// hand-rolled copies of it.
//
// Deliberately NOT used by primitive.StaticText: StaticText's whole
// reason to exist is a leaf that pays for zero lock, because its value
// genuinely never changes after construction -- see StaticText's own
// doc comment. Embedding TextBuffer there would just be "the same
// lock, now with a promise attached," not an actual cost saving.
type TextBuffer struct {
	mu    sync.RWMutex
	value []rune
}

func (b *TextBuffer) SetValue(v string) {
	b.mu.Lock()
	b.value = []rune(v)
	b.mu.Unlock()
}

func (b *TextBuffer) Value() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return string(b.value)
}

// Runes returns a defensive copy -- same convention as
// Propagator.Children(): a Draw call reading every rune shouldn't ever
// race a concurrent SetValue (e.g. from a background goroutine, the
// Spinner/clock pattern) replacing the backing slice underneath it.
func (b *TextBuffer) Runes() []rune {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]rune, len(b.value))
	copy(out, b.value)
	return out
}

func (b *TextBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.value)
}
