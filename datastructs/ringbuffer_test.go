package datastructs

import (
	"testing"
)

func TestNewRingBufferPanicsBelowCapacityTwo(t *testing.T) {
	for _, capacity := range []int{-1, 0, 1} {
		t.Run("", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("NewRingBuffer(%d) did not panic, want it to", capacity)
				}
			}()
			NewRingBuffer(capacity)
		})
	}
}

func TestRingBufferCapacityTwoMinimum(t *testing.T) {
	b := NewRingBuffer(2)
	b.Add("a")
	got, err := b.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "a" {
		t.Fatalf("Read() = %q, want %q", got, "a")
	}
}

func TestRingBufferReadOnEmptyErrors(t *testing.T) {
	b := NewRingBuffer(3)
	if _, err := b.Read(); err == nil {
		t.Fatal("expected an error reading an empty buffer, got nil")
	}
}

func TestRingBufferOverflowEvictsOldest(t *testing.T) {
	// Capacity 3 only ever holds 2 usable items (see the
	// capacity-N-holds-N-1 test below) -- adding a 3rd should evict
	// the oldest ("a"), leaving "b" then "c".
	b := NewRingBuffer(3)
	b.Add("a")
	b.Add("b")
	b.Add("c")

	got, err := b.Read()
	if err != nil {
		t.Fatal(err)
	}
	if got != "b" {
		t.Fatalf("first Read() = %q, want %q (oldest \"a\" should have been evicted)", got, "b")
	}

	got, err = b.Read()
	if err != nil {
		t.Fatal(err)
	}
	if got != "c" {
		t.Fatalf("second Read() = %q, want %q", got, "c")
	}

	if _, err := b.Read(); err == nil {
		t.Fatal("expected the buffer to be empty after draining both remaining items")
	}
}

// TestRingBufferCapacityNHoldsNMinusOne pins down the documented
// full-vs-empty ambiguity: a single read/write index pair can't tell
// "full" apart from "empty" without sacrificing one slot, so capacity
// N only ever holds N-1 items usable at once -- checked across a few
// different capacities, not just the one fault.FaultManager happens
// to use (100).
func TestRingBufferCapacityNHoldsNMinusOne(t *testing.T) {
	for _, capacity := range []int{2, 3, 5, 10} {
		t.Run("", func(t *testing.T) {
			b := NewRingBuffer(capacity)
			for i := 0; i < capacity; i++ {
				b.Add("x")
			}
			if got := b.Size(); got != capacity-1 {
				t.Errorf("capacity %d: Size() after filling = %d, want %d", capacity, got, capacity-1)
			}
		})
	}
}

// TestRingBufferSizeAcrossWraparound covers the writeIndex < readIndex
// case Size()'s own implementation branches on: fill the buffer,
// drain some, add more so writeIndex wraps around past 0 while
// readIndex is still ahead of it, and confirm Size() stays correct
// throughout rather than only being tested in the "hasn't wrapped yet"
// case.
func TestRingBufferSizeAcrossWraparound(t *testing.T) {
	b := NewRingBuffer(5) // 4 usable slots

	b.Add("1")
	b.Add("2")
	b.Add("3")
	if got := b.Size(); got != 3 {
		t.Fatalf("Size() after 3 adds = %d, want 3", got)
	}

	if _, err := b.Read(); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Read(); err != nil {
		t.Fatal(err)
	}
	if got := b.Size(); got != 1 {
		t.Fatalf("Size() after draining 2 = %d, want 1", got)
	}

	// writeIndex is now ahead of readIndex but close to the end of the
	// backing array -- these two Adds push writeIndex past the array's
	// end and back around to the front, landing BEFORE readIndex in
	// raw index terms even though there are more items now than before.
	b.Add("4")
	b.Add("5")
	b.Add("6")
	if got := b.Size(); got != 4 {
		t.Fatalf("Size() after wraparound = %d, want 4 (all 4 usable slots full)", got)
	}

	var got []string
	for {
		v, err := b.Read()
		if err != nil {
			break
		}
		got = append(got, v)
	}
	want := []string{"3", "4", "5", "6"}
	if len(got) != len(want) {
		t.Fatalf("drained %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("drained %v, want %v", got, want)
		}
	}
}
