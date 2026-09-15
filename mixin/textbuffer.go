package mixin

import "sync"

// TextBuffer is a mutex-protected, atomically-replaceable []rune
// value -- shared storage behind Text and (eventually) TextInput. Not
// used by primitive.StaticText, which never changes after construction
// and so needs no lock at all.
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

// Runes returns a defensive copy.
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
