package primitive

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

// MultilineText is mutable, wrapping, fixed-size text. A '\n' in the
// value is a hard break; each resulting line is greedily word-wrapped
// to fit the declared width. Content that still doesn't fit the
// declared height is TRUNCATED, not scrolled, with a Warning logged
// once per SetValue/Resize call that overflows -- including the
// constructor's own initial value, via a check in SetContext (a
// Warning fired before ctx exists would otherwise silently no-op).
type MultilineText struct {
	mixin.Node

	mu       sync.RWMutex
	raw      string
	lines    [][]rune
	overflow bool
}

type MultilineTextConfig struct {
	Value  string
	Fg     core.Color
	Anchor framework.Anchor
	Layer  int
}

func NewMultilineText(bounds *geom.Bounds, cfg MultilineTextConfig) (*MultilineText, error) {
	style := framework.Style{Bg: core.Transparent, Fg: cfg.Fg}

	bn, err := mixin.NewNode(bounds, cfg.Anchor, style, cfg.Layer, "MultilineText")
	if err != nil {
		return nil, err
	}

	t := &MultilineText{Node: bn}
	t.recompute(cfg.Value)
	return t, nil
}

func (t *MultilineText) Draw(buf *core.Buffer, vec geom.Vector) {
	t.mu.RLock()
	lines := t.lines
	t.mu.RUnlock()

	s := t.Style()
	pos := t.ComputedPos()

	for row, line := range lines {
		for col, r := range line {
			buf.Set(vec.X+pos.X+col, vec.Y+pos.Y+row, r, s.Bg, s.Fg)
		}
	}
}

func (t *MultilineText) SetValue(v string) {
	if t.recompute(v) {
		t.warnOverflow()
	}
	t.Invalidate()

	t.mu.RLock()
	n := len(t.lines)
	t.mu.RUnlock()
	t.Logger().Debug(fmt.Sprintf("value set (%d line(s))", n))
}

// Resize re-wraps the original raw value against the new bounds.
func (t *MultilineText) Resize(w, h int) {
	t.Node.Resize(w, h)

	t.mu.RLock()
	raw := t.raw
	t.mu.RUnlock()

	if t.recompute(raw) {
		t.warnOverflow()
	}
	t.Invalidate()
}

// SetContext re-checks overflow on attach -- see the type doc comment.
func (t *MultilineText) SetContext(ctx framework.AppContext) {
	t.Node.SetContext(ctx)

	t.mu.RLock()
	overflow := t.overflow
	t.mu.RUnlock()

	if overflow {
		t.warnOverflow()
	}
}

// Value returns the last raw string passed to SetValue, unwrapped.
// Use Lines() for what's actually on screen.
func (t *MultilineText) Value() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.raw
}

func (t *MultilineText) Lines() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]string, len(t.lines))
	for i, l := range t.lines {
		out[i] = string(l)
	}
	return out
}

func (t *MultilineText) Overflowed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.overflow
}

func (t *MultilineText) warnOverflow() {
	w, h := t.Size()
	t.Warn(fmt.Errorf("MultilineText: content does not fit in %dx%d and was truncated -- shrink the text, grow the widget, or split it across more than one MultilineText", w, h))
}

func (t *MultilineText) recompute(v string) (overflow bool) {
	w, h := t.Size()
	wrapped, truncated := wrapText(v, w, h)

	t.mu.Lock()
	t.raw = v
	t.lines = wrapped
	t.overflow = truncated
	t.mu.Unlock()

	return truncated
}

// wrapText honors '\n' as a hard break, word-wraps each resulting line
// to width, then truncates the result to height.
func wrapText(v string, width, height int) (lines [][]rune, truncated bool) {
	if width <= 0 || height <= 0 {
		return nil, len(v) > 0
	}

	for _, para := range strings.Split(v, "\n") {
		lines = append(lines, wrapLine(para, width)...)
	}

	if len(lines) > height {
		return lines[:height], true
	}
	return lines, false
}

// wrapLine greedily word-wraps one line to width, always returning at
// least one line. A word longer than width is hard-broken at width.
func wrapLine(s string, width int) [][]rune {
	words := strings.Fields(s)
	if len(words) == 0 {
		return [][]rune{{}}
	}

	var out [][]rune
	cur := make([]rune, 0, width)

	flush := func() {
		out = append(out, cur)
		cur = make([]rune, 0, width)
	}

	for _, w := range words {
		word := []rune(w)

		for len(word) > width {
			if len(cur) > 0 {
				flush()
			}
			out = append(out, word[:width])
			word = word[width:]
		}
		if len(word) == 0 {
			continue
		}

		need := len(word)
		if len(cur) > 0 {
			need++
		}
		if len(cur)+need > width {
			flush()
		}
		if len(cur) > 0 {
			cur = append(cur, ' ')
		}
		cur = append(cur, word...)
	}
	flush()

	return out
}
