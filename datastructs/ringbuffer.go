package datastructs

import (
	"errors"
)

type RingBuffer struct {
	buffer     []string
	readIndex  int
	writeIndex int
}

// This will be a panic cause its a framework fuck up and will not be exposed
//
// capacity < 2 is rejected, not just capacity < 0: readIndex ==
// writeIndex means empty (see Read/Size), which is also exactly what
// happens after Add fills the last free slot -- full and empty are
// otherwise indistinguishable with a single read/write index pair.
// capacity 0 hits that ambiguity immediately (the very first Add does
// (writeIndex+1) % len(buffer), a divide-by-zero on an empty backing
// slice) and capacity 1 can never hold anything (every Add would
// immediately look full to Read). This is also why a RingBuffer only
// ever has capacity-1 truly usable slots for any capacity >= 2 --
// fault.FaultManager's NewRingBuffer(100) really gives 99 usable retry
// slots, not 100. That's a deliberate consequence of this same
// full-vs-empty scheme, not a bug -- don't "fix" it without addressing
// full/empty ambiguity everywhere else in this type first.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity < 2 {
		panic("ring buffer size must be at least 2")
	}

	return &RingBuffer{
		buffer: make([]string, capacity),
	}
}

// Add inserts a value. If the buffer is full, the oldest value is discarded.
// I know this way we lose a log but i dont care
func (b *RingBuffer) Add(str string) {
	next := (b.writeIndex + 1) % len(b.buffer)

	// Buffer is full: discard the oldest element.
	if next == b.readIndex {
		b.readIndex = (b.readIndex + 1) % len(b.buffer)
	}

	b.buffer[b.writeIndex] = str
	b.writeIndex = next
}

// Read returns the oldest value.
func (b *RingBuffer) Read() (string, error) {
	if b.readIndex == b.writeIndex {
		return "", errors.New("nothing to read")
	}

	str := b.buffer[b.readIndex]
	b.readIndex = (b.readIndex + 1) % len(b.buffer)

	return str, nil
}

func (b *RingBuffer) Size() int {
	if b.writeIndex >= b.readIndex {
		return b.writeIndex - b.readIndex
	}

	return len(b.buffer) - b.readIndex + b.writeIndex
}
