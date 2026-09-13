package primitive

import (
	"reflect"
	"testing"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

func linesOf(ss ...string) [][]rune {
	out := make([][]rune, len(ss))
	for i, s := range ss {
		out[i] = []rune(s)
	}
	return out
}

func TestWrapText(t *testing.T) {
	cases := []struct {
		name          string
		value         string
		width, height int
		wantLines     [][]rune
		wantTruncated bool
	}{
		{
			name:  "fits on one line, no wrap needed",
			value: "hello",
			width: 10, height: 3,
			wantLines: linesOf("hello"),
		},
		{
			name:  "explicit newline is a hard break",
			value: "line one\nline two",
			width: 20, height: 3,
			wantLines: linesOf("line one", "line two"),
		},
		{
			name:  "word wrap packs greedily",
			value: "the quick brown fox",
			width: 10, height: 5,
			wantLines: linesOf("the quick", "brown fox"),
		},
		{
			name:  "a single word longer than width is hard-broken",
			value: "supercalifragilistic",
			width: 10, height: 5,
			wantLines: linesOf("supercalif", "ragilistic"),
		},
		{
			name:  "blank paragraph still consumes a row",
			value: "a\n\nb",
			width: 10, height: 5,
			wantLines: linesOf("a", "", "b"),
		},
		{
			name:  "too many wrapped rows truncates and reports it",
			value: "one\ntwo\nthree\nfour",
			width: 10, height: 2,
			wantLines:     linesOf("one", "two"),
			wantTruncated: true,
		},
		{
			name:  "zero width with non-empty value overflows immediately",
			value: "x",
			width: 0, height: 3,
			wantLines:     nil,
			wantTruncated: true,
		},
		{
			name:  "zero width with empty value does not overflow",
			value: "",
			width: 0, height: 3,
			wantLines:     nil,
			wantTruncated: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotLines, gotTruncated := wrapText(tc.value, tc.width, tc.height)
			if !reflect.DeepEqual(gotLines, tc.wantLines) {
				t.Errorf("lines = %q, want %q", runeLinesToStrings(gotLines), runeLinesToStrings(tc.wantLines))
			}
			if gotTruncated != tc.wantTruncated {
				t.Errorf("truncated = %v, want %v", gotTruncated, tc.wantTruncated)
			}
		})
	}
}

func runeLinesToStrings(lines [][]rune) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = string(l)
	}
	return out
}

// TestMultilineTextOverflowWarnedOnAttach covers the timing gap the
// type's own doc comment calls out: a Warning issued from inside the
// constructor would silently no-op (no ctx yet), so the overflow from
// the CONSTRUCTOR's own value has to be re-checked and actually
// reported once SetContext gives it somewhere to go.
func TestMultilineTextOverflowWarnedOnAttach(t *testing.T) {
	mt, err := NewMultilineText(geom.NewBounds(0, 0, 5, 1), MultilineTextConfig{
		Value: "this is way too much text",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !mt.Overflowed() {
		t.Fatal("expected Overflowed() to be true after construction")
	}

	logs := make(chan core.AppLog, 4)
	ctx := framework.AppContext{Logs: logs, LogLevel: core.Warning}
	mt.SetContext(ctx)

	select {
	case l := <-logs:
		if l.Severity() != core.Warning {
			t.Fatalf("severity = %v, want Warning", l.Severity())
		}
	default:
		t.Fatal("expected a Warning to be logged once ctx was attached, got none")
	}
}

// TestMultilineTextResizeRewraps confirms Resize re-wraps the ORIGINAL
// raw value against the new bounds, rather than leaving stale lines
// from the previous size on screen.
func TestMultilineTextResizeRewraps(t *testing.T) {
	mt, err := NewMultilineText(geom.NewBounds(0, 0, 20, 3), MultilineTextConfig{
		Value: "the quick brown fox",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := mt.Lines(); !reflect.DeepEqual(got, []string{"the quick brown fox"}) {
		t.Fatalf("initial Lines() = %q", got)
	}

	mt.Resize(10, 3)

	want := []string{"the quick", "brown fox"}
	if got := mt.Lines(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Lines() after Resize = %q, want %q", got, want)
	}
	if mt.Value() != "the quick brown fox" {
		t.Fatalf("Value() = %q, want the original raw string unchanged", mt.Value())
	}
}
