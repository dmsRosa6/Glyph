package datastructs

import (
	"errors"
)

type RingBuffer struct {
	buffer     []string
	readIndex  int
	writeIndex int
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity < 2 {
		panic("ring buffer size must be at least 2")
	}

	return &RingBuffer{
		buffer: make([]string, capacity),
	}
}

func (b *RingBuffer) Add(str string) {
	next := (b.writeIndex + 1) % len(b.buffer)

	if next == b.readIndex {
		b.readIndex = (b.readIndex + 1) % len(b.buffer)
	}

	b.buffer[b.writeIndex] = str
	b.writeIndex = next
}

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
