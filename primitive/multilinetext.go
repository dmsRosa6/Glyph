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

// MultilineText is mutable, wrapping, fixed-size text -- a size is
// given up front (bounds.W x bounds.H), and content is greedily
// word-wrapped to fit that width. An explicit '\n' in the value is
// always honored as a hard line break FIRST (so callers get real
// paragraphs, not just an accident of wrapping); each resulting
// logical line is then word-wrapped independently.
//
// If the wrapped content doesn't fit in bounds.H rows, it is
// TRUNCATED, not scrolled -- there is no scroll state here (see
// todo.md's scrollable-container item for where that eventually goes)
// -- and a Warning is logged once per SetValue/Resize call that
// actually overflows, the same "warn once per real cause, not once per
// frame" shape Container's own fit-warning already uses (see
// canvas.Container.refreshFitWarnings). Truncation never panics or
// grows past the declared bounds no matter what's passed to SetValue --
// on the same "log it, then it's on the caller" terms this codebase
// already uses elsewhere (core/ringbuffer.go's own doc comment is the
// clearest example): shrink the text, grow the box, or split it across
// more than one MultilineText. This widget will not go out of its way
// to protect a caller who keeps shipping more text than the box they
// built can hold.
//
// Concurrency note: recompute reads this node's current Size() outside
// any lock, consistent with the rest of this framework's assumption
// that Resize is called from whichever goroutine owns structural/
// terminal-size changes (see mixin.Node's own doc comment on ctx), not
// concurrently with arbitrary other goroutines. If you're calling
// SetValue on a MultilineText from a background goroutine (the
// Spinner/clock pattern) AND that same widget can also be resized (a
// terminal resize cascading down, or a manual Resize call) at the same
// moment, there's a narrow read/write race on the underlying bounds --
// same class of known, documented gap as Container.fitMu's own "Known
// gap" note, not something this type goes further to solve.
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
	// Overflow (if any) from THIS initial value can't be logged yet --
	// Logger.Warning is a nil-safe no-op before ctx is set (see
	// framework.Logger's own doc comment), and ctx isn't set until this
	// node is attached to a tree. recompute still runs so Draw has
	// something correct to show even before that happens; SetContext
	// below is what actually reports the overflow once there's
	// somewhere for the warning to go.
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

// SetValue re-wraps v against the current bounds and replaces the
// displayed content. Logs a Warning if the new content overflows --
// see the type's own doc comment for what "overflow" means here and
// why it's a warning, not a rejected call.
func (t *MultilineText) SetValue(v string) {
	overflow := t.recompute(v)
	if overflow {
		t.warnOverflow()
	}
	t.Invalidate()

	t.mu.RLock()
	n := len(t.lines)
	t.mu.RUnlock()
	t.Logger().Debug(fmt.Sprintf("value set (%d line(s))", n))
}

// Resize shadows the promoted mixin.Node.Resize: a size change
// invalidates the previous wrap -- the same content can newly overflow
// a shrunk box, or newly fit a grown one -- so the raw value is
// re-wrapped against the new bounds rather than leaving stale,
// wrong-width lines on screen until the next SetValue.
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

// SetContext additionally re-checks overflow the moment this node is
// attached to a tree -- see NewMultilineText's own comment for why
// this is what actually makes the constructor's initial value get
// warned about in the common "build, then immediately AddChild" case.
func (t *MultilineText) SetContext(ctx framework.AppContext) {
	t.Node.SetContext(ctx)

	t.mu.RLock()
	overflow := t.overflow
	t.mu.RUnlock()

	if overflow {
		t.warnOverflow()
	}
}

// Value returns the last raw string passed to SetValue (or the
// constructor), unwrapped -- not the wrapped/truncated lines actually
// on screen. Use Lines() for what's actually being drawn.
func (t *MultilineText) Value() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.raw
}

// Lines returns a defensive copy of the wrapped, possibly-truncated
// lines currently being drawn -- same copy-on-read convention as
// mixin.Propagator.Children() and mixin.TextBuffer.Runes().
func (t *MultilineText) Lines() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]string, len(t.lines))
	for i, l := range t.lines {
		out[i] = string(l)
	}
	return out
}

// Overflowed reports whether the current content had to be truncated
// to fit -- the same condition that drove the last Warning, exposed so
// a caller can react to it programmatically (e.g. show a "..." marker
// itself) without needing to parse log output.
func (t *MultilineText) Overflowed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.overflow
}

func (t *MultilineText) warnOverflow() {
	w, h := t.Size()
	t.Warn(fmt.Errorf("MultilineText: content does not fit in %dx%d and was truncated -- shrink the text, grow the widget, or split it across more than one MultilineText", w, h))
}

// recompute re-wraps v against this node's current bounds, stores the
// result, and reports whether it had to truncate. Guarded by mu for
// the whole read-modify-write, not just the final assignment -- Draw/
// Value/Lines could otherwise observe a half-updated (lines replaced,
// overflow not yet updated, or vice versa) state from a concurrent
// SetValue.
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

// wrapText honors '\n' in v as a hard break first, word-wraps each
// resulting logical line independently to fit width, then truncates
// the combined result to height. width <= 0 or height <= 0 can't wrap
// anything meaningful -- treated as an immediate overflow (unless v is
// empty) rather than a divide-by-zero or an infinite loop.
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

// wrapLine greedily word-wraps one '\n'-free line to fit width,
// always returning at least one (possibly empty) line so a blank
// paragraph still consumes a row, the same as a text editor would. A
// single word longer than width is hard-broken at width rather than
// left to overflow sideways, or stall wrapping forever waiting for a
// fit that can never happen.
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
			continue // the hard-break loop consumed it exactly
		}

		need := len(word)
		if len(cur) > 0 {
			need++ // +1 for the separating space
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
