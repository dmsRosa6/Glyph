package widgets

import (
	"testing"

	"github.com/dmsRosa6/glyph/geom"
)

// buildList constructs a List with n single-row-tall items, each
// AddItem-created row, in a viewport viewH rows tall.
func buildList(t *testing.T, viewH, n int, scrollable bool) (*List, []*ListRow) {
	t.Helper()
	l, err := NewList(geom.NewBounds(0, 0, 10, viewH), ListConfig{Scrollable: scrollable})
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]*ListRow, n)
	for i := 0; i < n; i++ {
		row, err := l.AddItem(1)
		if err != nil {
			t.Fatal(err)
		}
		rows[i] = row
	}
	return l, rows
}

func TestListFixedNeverAutoScrolls(t *testing.T) {
	l, rows := buildList(t, 3, 5, false)
	rows[4].Focus()
	l.followFocus()
	if got := l.ScrollY(); got != 0 {
		t.Fatalf("ScrollY() = %d, want 0 for a non-Scrollable List regardless of focus", got)
	}
}

func TestListScrollableFollowsFocusDown(t *testing.T) {
	// 5 rows, 1 cell tall each, in a 3-row viewport. Focusing the last
	// row (occupying [4,5)) should scroll just enough to bring row 5
	// (exclusive) to the bottom of the viewport: scrollY = 5-3 = 2.
	l, rows := buildList(t, 3, 5, true)
	rows[4].Focus()
	l.followFocus()
	if got := l.ScrollY(); got != 2 {
		t.Fatalf("ScrollY() = %d, want 2", got)
	}
}

func TestListScrollableFollowsFocusBackUp(t *testing.T) {
	l, rows := buildList(t, 3, 5, true)
	rows[4].Focus()
	l.followFocus()
	rows[4].Blur()
	rows[0].Focus()
	l.followFocus()
	if got := l.ScrollY(); got != 0 {
		t.Fatalf("ScrollY() = %d, want 0 after focusing the first row again", got)
	}
}

func TestListScrollableDoesNotMoveWhenAlreadyVisible(t *testing.T) {
	// Viewport 3 rows, 5 rows total. Focus row 4 (scrolls to 2), then
	// focus row 3 -- row 3 occupies [3,4), already fully inside
	// [2,5), so the scroll offset should not change at all.
	l, rows := buildList(t, 3, 5, true)
	rows[4].Focus()
	l.followFocus()
	rows[4].Blur()
	rows[3].Focus()
	l.followFocus()
	if got := l.ScrollY(); got != 2 {
		t.Fatalf("ScrollY() = %d, want unchanged 2 -- row 3 was already fully visible", got)
	}
}

func TestListScrollableNoFocusedRowIsNoop(t *testing.T) {
	l, _ := buildList(t, 3, 5, true)
	l.ScrollTo(1)
	l.followFocus() // nothing is focused
	if got := l.ScrollY(); got != 1 {
		t.Fatalf("ScrollY() = %d, want unchanged 1 when nothing is focused", got)
	}
}

func TestListScrollByAndScrollToClamp(t *testing.T) {
	l, _ := buildList(t, 3, 5, true) // total content height 5, viewport 3 -> max scroll 2

	l.ScrollBy(100)
	if got := l.ScrollY(); got != 2 {
		t.Fatalf("ScrollBy(100): ScrollY() = %d, want clamped to 2", got)
	}

	l.ScrollTo(-50)
	if got := l.ScrollY(); got != 0 {
		t.Fatalf("ScrollTo(-50): ScrollY() = %d, want clamped to 0", got)
	}

	l.ScrollTo(1)
	if got := l.ScrollY(); got != 1 {
		t.Fatalf("ScrollTo(1): ScrollY() = %d, want 1", got)
	}
}

func TestListScrollByAndScrollToAreNoopsWhenNotScrollable(t *testing.T) {
	l, _ := buildList(t, 3, 5, false)
	l.ScrollBy(1)
	l.ScrollTo(1)
	if got := l.ScrollY(); got != 0 {
		t.Fatalf("ScrollY() = %d, want 0 -- ScrollBy/ScrollTo should be no-ops on a non-Scrollable List", got)
	}
}

func TestListScrollClampsWhenContentFitsViewport(t *testing.T) {
	// Only 2 rows in a 3-row viewport -- everything already fits, so
	// max scroll is 0 (not negative), and any ScrollBy should clamp
	// straight back to 0.
	l, _ := buildList(t, 3, 2, true)
	l.ScrollBy(5)
	if got := l.ScrollY(); got != 0 {
		t.Fatalf("ScrollY() = %d, want 0 when content fits entirely within the viewport", got)
	}
}
