package datastructs

import "errors"

type RingBuffer struct {
	buffer     []string
	readIndex  int
	writeIndex int
}

func NewRingBuffer(size int) *RingBuffer {
	if size < 2 {
		panic("ring buffer size must be at least 2")
	}

	return &RingBuffer{
		buffer: make([]string, size),
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
