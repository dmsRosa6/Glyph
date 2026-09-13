package primitive

import (
	"testing"

	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

func TestTextInputConstructorRejectsZeroWidth(t *testing.T) {
	if _, err := NewTextInput(&geom.Point{}, 0, TextInputConfig{}); err == nil {
		t.Fatal("expected an error for width < 1, got nil")
	}
}

func TestTextInputInitialCursorAtEnd(t *testing.T) {
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{Value: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if got := ti.Cursor(); got != 5 {
		t.Fatalf("Cursor() = %d, want 5", got)
	}
	if got := ti.Value(); got != "hello" {
		t.Fatalf("Value() = %q, want %q", got, "hello")
	}
}

func TestTextInputTypingInsertsAtCursor(t *testing.T) {
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{Value: "hello"})
	if err != nil {
		t.Fatal(err)
	}

	handled, err := ti.HandleInput(framework.Event{Key: framework.KeyRune, Rune: '!'})
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected a plain rune to be handled")
	}
	if got := ti.Value(); got != "hello!" {
		t.Fatalf("Value() = %q, want %q", got, "hello!")
	}
	if got := ti.Cursor(); got != 6 {
		t.Fatalf("Cursor() = %d, want 6", got)
	}
}

func TestTextInputCtrlModifiedRuneFallsThrough(t *testing.T) {
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{Value: "hello"})
	if err != nil {
		t.Fatal(err)
	}

	// A Ctrl+H event decodes as KeyRune{Rune: 'h', Modifiers: ModCtrl}
	// (see input.Manager's decoder). TextInput only binds the plain
	// (ModNone) rune case, so this must NOT be claimed -- otherwise
	// Ctrl+C (and every other global Ctrl+<key> binding) would stop
	// working the moment a TextInput has focus.
	handled, err := ti.HandleInput(framework.Event{Key: framework.KeyRune, Rune: 'h', Modifiers: framework.ModCtrl})
	if err != nil {
		t.Fatal(err)
	}
	if handled {
		t.Fatal("expected a Ctrl+<rune> event to fall through unclaimed")
	}
	if got := ti.Value(); got != "hello" {
		t.Fatalf("Value() = %q, want unchanged %q", got, "hello")
	}
}

func TestTextInputBackspaceAndDelete(t *testing.T) {
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{Value: "hello"})
	if err != nil {
		t.Fatal(err)
	}

	// Cursor starts at the end (5). Backspace removes 'o'.
	ti.HandleInput(framework.Event{Key: framework.KeyBackspace})
	if got := ti.Value(); got != "hell" {
		t.Fatalf("after Backspace: Value() = %q, want %q", got, "hell")
	}
	if got := ti.Cursor(); got != 4 {
		t.Fatalf("after Backspace: Cursor() = %d, want 4", got)
	}

	// Home, then Delete removes the leading 'h' without moving the
	// cursor.
	ti.HandleInput(framework.Event{Key: framework.KeyHome})
	ti.HandleInput(framework.Event{Key: framework.KeyDelete})
	if got := ti.Value(); got != "ell" {
		t.Fatalf("after Home+Delete: Value() = %q, want %q", got, "ell")
	}
	if got := ti.Cursor(); got != 0 {
		t.Fatalf("after Home+Delete: Cursor() = %d, want 0", got)
	}
}

func TestTextInputEdgeBackspaceAndDeleteAreNoops(t *testing.T) {
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{Value: "hi"})
	if err != nil {
		t.Fatal(err)
	}

	ti.HandleInput(framework.Event{Key: framework.KeyHome})
	ti.HandleInput(framework.Event{Key: framework.KeyBackspace}) // nothing before cursor 0
	if got := ti.Value(); got != "hi" {
		t.Fatalf("Backspace at start: Value() = %q, want unchanged %q", got, "hi")
	}

	ti.HandleInput(framework.Event{Key: framework.KeyEnd})
	ti.HandleInput(framework.Event{Key: framework.KeyDelete}) // nothing after cursor at end
	if got := ti.Value(); got != "hi" {
		t.Fatalf("Delete at end: Value() = %q, want unchanged %q", got, "hi")
	}
}

func TestTextInputLeftRightHomeEnd(t *testing.T) {
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{Value: "abc"})
	if err != nil {
		t.Fatal(err)
	}

	ti.HandleInput(framework.Event{Key: framework.KeyHome})
	if got := ti.Cursor(); got != 0 {
		t.Fatalf("Home: Cursor() = %d, want 0", got)
	}

	ti.HandleInput(framework.Event{Key: framework.KeyRight})
	ti.HandleInput(framework.Event{Key: framework.KeyRight})
	if got := ti.Cursor(); got != 2 {
		t.Fatalf("Right x2: Cursor() = %d, want 2", got)
	}

	ti.HandleInput(framework.Event{Key: framework.KeyLeft})
	if got := ti.Cursor(); got != 1 {
		t.Fatalf("Left: Cursor() = %d, want 1", got)
	}

	ti.HandleInput(framework.Event{Key: framework.KeyEnd})
	if got := ti.Cursor(); got != 3 {
		t.Fatalf("End: Cursor() = %d, want 3", got)
	}
}

func TestTextInputSetValueMovesCursorToEnd(t *testing.T) {
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{Value: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	ti.HandleInput(framework.Event{Key: framework.KeyHome})

	ti.SetValue("hello world")

	if got := ti.Value(); got != "hello world" {
		t.Fatalf("Value() = %q, want %q", got, "hello world")
	}
	if got := ti.Cursor(); got != len("hello world") {
		t.Fatalf("Cursor() = %d, want %d (end of new value)", got, len("hello world"))
	}
}

func TestTextInputOnSubmit(t *testing.T) {
	var got string
	var calls int
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{
		Value: "submit me",
		OnSubmit: func(value string) (bool, error) {
			got = value
			calls++
			return true, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	handled, err := ti.HandleInput(framework.Event{Key: framework.KeyEnter})
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected Enter to be handled when OnSubmit is set")
	}
	if calls != 1 {
		t.Fatalf("OnSubmit called %d times, want 1", calls)
	}
	if got != "submit me" {
		t.Fatalf("OnSubmit received %q, want %q", got, "submit me")
	}
}

func TestTextInputNoOnSubmitLeavesEnterUnbound(t *testing.T) {
	ti, err := NewTextInput(&geom.Point{}, 10, TextInputConfig{Value: "x"})
	if err != nil {
		t.Fatal(err)
	}

	handled, err := ti.HandleInput(framework.Event{Key: framework.KeyEnter})
	if err != nil {
		t.Fatal(err)
	}
	if handled {
		t.Fatal("expected Enter to fall through unclaimed when OnSubmit is nil")
	}
}

func TestScrollOffset(t *testing.T) {
	cases := []struct {
		name          string
		cursor, width int
		want          int
	}{
		{"cursor within the visible window", 3, 10, 0},
		{"cursor exactly at the last visible column", 9, 10, 0},
		{"cursor one past the window follows it", 10, 10, 1},
		{"cursor far to the right keeps the window pinned to it", 25, 10, 16},
		{"zero width never panics, always 0", 5, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := scrollOffset(tc.cursor, tc.width); got != tc.want {
				t.Errorf("scrollOffset(%d, %d) = %d, want %d", tc.cursor, tc.width, got, tc.want)
			}
		})
	}
}
