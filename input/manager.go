package input

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/term"
)

// decodeState tracks how far into a multi-byte escape sequence we are.
type decodeState int

const (
	stateNormal    decodeState = iota
	stateEsc                   // just saw 0x1b, waiting to see if more follows
	stateCSI                   // saw ESC [, waiting to see what kind of sequence this is
	stateCSIParams             // accumulating CSI parameter bytes until a terminator
)

// DefaultEventBufferSize is Events()'s buffer capacity when
// NewManager's bufferSize is left at 0 (or negative). Events beyond
// this, arriving faster than the consumer drains them, are dropped
// (see send below) rather than blocking the decode loop -- 16 is
// plenty of headroom for ordinary typing and clicking, but a
// drag-heavy app (see examples/mouse-paint-demo, which can easily
// emit more than 16 MouseDrag events between renderer ticks under
// OnDemand mode) may want to pass a larger bufferSize explicitly.
const DefaultEventBufferSize = 32

// byteSource is the seam between Manager's decode loop and the actual
// byte stream it reads from. Production code always gets stdinSource
// (below), a thin wrapper over term.ReadStdin against the real tty;
// tests substitute a scripted sequence of bytes/timeouts/errors so
// every branch of the escape-sequence decoder -- plain keys,
// Ctrl+letter, arrows, modified arrows, SGR mouse press/drag/wheel,
// malformed/truncated sequences -- can be exercised without a real tty
// at all. Before this seam existed, the decoder was fused directly to
// term.ReadStdin and had zero test coverage despite being genuinely
// intricate state-machine logic.
type byteSource interface {
	Read(buf []byte) (int, error)
}

type stdinSource struct{}

func (stdinSource) Read(buf []byte) (int, error) {
	return term.ReadStdin(buf)
}

type Manager struct {
	events chan framework.Event
	logger framework.Logger
	source byteSource

	ctx     context.Context
	cancel  context.CancelFunc
	restore func()
	stopped chan struct{}
}

// NewManager builds a Manager reading from the real terminal.
// bufferSize sets Events()'s channel capacity; <= 0 uses
// DefaultEventBufferSize.
func NewManager(logger framework.Logger, bufferSize int) (*Manager, error) {
	if bufferSize <= 0 {
		bufferSize = DefaultEventBufferSize
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		events:  make(chan framework.Event, bufferSize),
		logger:  logger,
		source:  stdinSource{},
		ctx:     ctx,
		cancel:  cancel,
		stopped: make(chan struct{}),
	}, nil
}

func (m *Manager) Events() <-chan framework.Event { return m.events }

func (m *Manager) Start() error {

	restore, err := term.SafeRawMode()
	if err != nil {
		return err
	}

	m.restore = restore

	go m.run()

	return nil
}

// Stop can take up to ~100ms: the read loop only checks m.ctx.Done()
// between blocking term.ReadStdin calls, and those time out at ~100ms
// each (VMIN=0/VTIME=1 -- see term.EnableRawMode's doc comment), so in
// the worst case Stop blocks on <-m.stopped for nearly a full read
// timeout before the loop notices it should exit. Not a bug -- just
// worth knowing if you're calling this from a signal handler expecting
// a near-instant return. (App.Stop layers FaultManager.Stop's own
// potential wait, on its 1s retry ticker, on top of this.)
func (m *Manager) Stop() {
	if m.restore == nil {
		return
	}
	m.logger.Info("Input manager stopping")
	m.cancel()
	<-m.stopped
	m.restore()
}

func (m *Manager) run() {
	defer close(m.stopped)
	defer close(m.events)

	state := stateNormal
	var csiBuf []byte
	var buf [1]byte

	m.logger.Info("Input manager started")

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		n, err := m.source.Read(buf[:])
		if err != nil {
			return
		}

		if n == 0 {
			if state == stateEsc {
				m.send(framework.Event{Key: framework.KeyEsc})
			}
			// A sequence that stalls mid-flight on a read timeout --
			// whether a lone ESC (state == stateEsc, handled above) or
			// a CSI sequence that got partway through its parameters
			// (state == stateCSI or stateCSIParams) -- is abandoned
			// here: state resets to stateNormal UNCONDITIONALLY, not
			// only for the stateEsc case. An incomplete escape
			// sequence has no safe single-key interpretation, so it's
			// dropped rather than replayed as raw KeyRune events, same
			// convention as an unrecognized sequence below.
			//
			// This used to only reset state inside the `if state ==
			// stateEsc` branch above, leaving state stuck at stateCSI/
			// stateCSIParams across the timeout despite this comment
			// already claiming the sequence was abandoned. Concretely:
			// type ESC [ 1, then pause long enough to time out, then
			// type A (a plain Up-arrow key) -- without this reset, that
			// stray 'A' gets fed into handleCSIParams as if it were
			// still completing the abandoned "ESC [ 1..." sequence,
			// misreading an ordinary later keypress using stale csiBuf
			// bytes from a sequence that had already timed out. Caught
			// by TestTimeoutMidCSIParamsDoesNotLeakIntoNextByte once
			// the byteSource seam made this decoder testable at all.
			state = stateNormal
			continue
		}

		ch := buf[0]

		switch state {
		case stateNormal:
			state = m.handleNormal(ch)

		case stateEsc:
			if ch == '[' {
				state = stateCSI
				csiBuf = csiBuf[:0]
			} else {
				m.send(framework.Event{Key: framework.KeyEsc})
				state = m.handleNormal(ch)
			}

		case stateCSI:
			state = m.handleCSI(ch, &csiBuf)

		case stateCSIParams:
			state = m.handleCSIParams(ch, &csiBuf)
		}
	}
}

// handleCSI decides what KIND of CSI sequence this is from the first
// byte right after ESC [. Plain, parameter-less sequences (arrow keys,
// Shift+Tab) resolve immediately. Anything that carries parameters -- a
// mouse report (starts with '<') or a modified key like Ctrl+Left
// ("1;5D") -- switches to stateCSIParams to accumulate the
// variable-length parameter bytes that follow.
func (m *Manager) handleCSI(ch byte, csiBuf *[]byte) decodeState {
	switch ch {
	case 'A':
		m.send(framework.Event{Key: framework.KeyUp})
		return stateNormal
	case 'B':
		m.send(framework.Event{Key: framework.KeyDown})
		return stateNormal
	case 'C':
		m.send(framework.Event{Key: framework.KeyRight})
		return stateNormal
	case 'D':
		m.send(framework.Event{Key: framework.KeyLeft})
		return stateNormal
	case 'Z':
		// CSI Z is the fixed, parameter-less xterm sequence for
		// Shift+Tab -- unlike Ctrl+Left etc. it does NOT go through the
		// "1;N<letter>" modifier-parameter form below.
		m.send(framework.Event{Key: framework.KeyTab, Modifiers: framework.ModShift})
		return stateNormal
	case '<':
		// Start of an SGR mouse report: ESC [ < Cb ; Cx ; Cy M/m.
		*csiBuf = append(*csiBuf, ch)
		return stateCSIParams
	default:
		if ch >= '0' && ch <= '9' {
			// Start of a parameterized sequence, e.g. "1;5D" for
			// Ctrl+Left.
			*csiBuf = append(*csiBuf, ch)
			return stateCSIParams
		}
		// unrecognized escape sequence -- drop it silently, same as the
		// original decoder's default case.
		return stateNormal
	}
}

// handleCSIParams accumulates parameter bytes (digits, ';') until a
// terminator. Mouse sequences (buffer starts with '<') terminate on 'M'
// (press/drag) or 'm' (release); modified-key sequences terminate on
// the direction letter itself (A/B/C/D), the same alphabet as the
// unmodified arrow keys in handleCSI.
func (m *Manager) handleCSIParams(ch byte, csiBuf *[]byte) decodeState {
	isMouse := len(*csiBuf) > 0 && (*csiBuf)[0] == '<'

	if isMouse && (ch == 'M' || ch == 'm') {
		m.decodeMouseSequence(string((*csiBuf)[1:]), ch == 'M')
		return stateNormal
	}

	if !isMouse {
		switch ch {
		case 'A', 'B', 'C', 'D':
			m.decodeModifiedArrow(string(*csiBuf), ch)
			return stateNormal
		}
	}

	if (ch >= '0' && ch <= '9') || ch == ';' {
		*csiBuf = append(*csiBuf, ch)
		return stateCSIParams
	}

	// Any other byte mid-sequence (an unrecognized terminator, garbage,
	// a second ESC) abandons this sequence -- dropped silently, same
	// convention as everywhere else in this decoder.
	return stateNormal
}

// decodeMouseSequence parses the "Cb;Cx;Cy" body of an SGR mouse report
// (the ESC [ < prefix and the M/m terminator are already stripped by
// the caller) and sends the resulting Event. press is true for a
// trailing 'M' (button pressed, or held-drag), false for a trailing 'm'
// (button released).
//
// Cb encodes, per the xterm SGR mouse protocol:
//
//	bits 0-1: button number (0=left, 1=middle, 2=right)
//	bit 2 (4):  Shift held
//	bit 3 (8):  Alt/Meta held
//	bit 4 (16): Ctrl held
//	bit 5 (32): motion flag -- set for a drag report (button held while
//	            moving), NOT set for a plain press/release
//	bit 6 (64): wheel flag -- when set, bits 0-1 distinguish wheel up
//	            (0) from wheel down (1) instead of a button number
//
// Cx/Cy are 1-indexed terminal columns/rows in the wire protocol; this
// converts to 0-indexed to match every other coordinate in this
// codebase (geom.Point, core.Buffer, BaseNode.ComputedPos are all
// 0-indexed from the top-left).
func (m *Manager) decodeMouseSequence(body string, press bool) {
	parts := strings.Split(body, ";")
	if len(parts) != 3 {
		return // malformed -- drop silently
	}
	cb, err1 := strconv.Atoi(parts[0])
	cx, err2 := strconv.Atoi(parts[1])
	cy, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return
	}

	ev := framework.Event{
		Kind:   framework.EventKindMouse,
		MouseX: cx - 1,
		MouseY: cy - 1,
	}

	if cb&4 != 0 {
		ev.Modifiers |= framework.ModShift
	}
	if cb&8 != 0 {
		ev.Modifiers |= framework.ModAlt
	}
	if cb&16 != 0 {
		ev.Modifiers |= framework.ModCtrl
	}

	switch {
	case cb&64 != 0:
		// Wheel event: bits 0-1 distinguish direction, not a button.
		ev.MouseButton = framework.MouseButtonNone
		if cb&3 == 0 {
			ev.MouseAction = framework.MouseWheelUp
		} else {
			ev.MouseAction = framework.MouseWheelDown
		}
	case cb&32 != 0:
		// Motion flag set alongside a real button: a drag report.
		ev.MouseButton = framework.MouseButton(cb & 3)
		ev.MouseAction = framework.MouseDrag
	default:
		ev.MouseButton = framework.MouseButton(cb & 3)
		if press {
			ev.MouseAction = framework.MousePress
		} else {
			ev.MouseAction = framework.MouseRelease
		}
	}

	m.send(ev)
}

// decodeModifiedArrow parses the "N" or "N;M" parameter body of a
// modified arrow-key sequence (e.g. "1;5" before a trailing 'D' for
// Ctrl+Left) and sends the resulting Event. dir is the terminator byte
// already seen by the caller (A/B/C/D).
//
// The xterm convention always sends a leading "1" as the first
// parameter for these (a legacy "repeat count" that never actually
// varies for a plain keypress), followed by ";M" where M-1 is a
// bitmask: bit0=Shift, bit1=Alt, bit2=Ctrl.
func (m *Manager) decodeModifiedArrow(body string, dir byte) {
	var key framework.Key
	switch dir {
	case 'A':
		key = framework.KeyUp
	case 'B':
		key = framework.KeyDown
	case 'C':
		key = framework.KeyRight
	case 'D':
		key = framework.KeyLeft
	}

	ev := framework.Event{Key: key}

	parts := strings.Split(body, ";")
	if len(parts) == 2 {
		if mod, err := strconv.Atoi(parts[1]); err == nil && mod > 0 {
			bits := mod - 1
			if bits&1 != 0 {
				ev.Modifiers |= framework.ModShift
			}
			if bits&2 != 0 {
				ev.Modifiers |= framework.ModAlt
			}
			if bits&4 != 0 {
				ev.Modifiers |= framework.ModCtrl
			}
		}
	}

	m.send(ev)
}

// handleNormal decodes a single byte outside of any escape sequence.
// 0x01-0x1A is the C0 control range that plain Ctrl+<letter> arrives as
// on the wire -- a terminal gives no way to distinguish "the user
// pressed Ctrl+H" from "byte 0x08 arrived," so this range IS the
// modifier information for these keys, decoded directly rather than
// through an escape sequence.
//
// A few values in that range already had dedicated, load-bearing
// meanings before Modifiers existed -- 0x03 (KeyCtrlC), 0x09 (KeyTab),
// 0x0D/0x0A (KeyEnter) -- and keep those exact meanings unchanged
// rather than being reinterpreted as Ctrl+C/Ctrl+I/Ctrl+M now that a
// Modifiers field exists; see event.go's doc comment on why KeyCtrlC in
// particular stays a dedicated constant. Every OTHER byte in the range
// (previously falling through to the default case as an unprintable
// KeyRune -- e.g. Ctrl+H arriving as Rune(0x08)) now decodes as the
// actual letter with ModCtrl set instead.
func (m *Manager) handleNormal(ch byte) decodeState {
	switch ch {
	case 0x1b: // ESC
		return stateEsc
	case 0x03: // Ctrl+C
		m.send(framework.Event{Key: framework.KeyCtrlC})
	case '\r', '\n':
		m.send(framework.Event{Key: framework.KeyEnter})
	case '\t':
		m.send(framework.Event{Key: framework.KeyTab})
	default:
		if ch >= 0x01 && ch <= 0x1a {
			m.send(framework.Event{
				Key:       framework.KeyRune,
				Rune:      rune('a' + ch - 1),
				Modifiers: framework.ModCtrl,
			})
		} else {
			m.send(framework.Event{Key: framework.KeyRune, Rune: rune(ch)})
		}
	}
	return stateNormal
}

// send is non-blocking: a full Events() buffer means the consumer
// (App.Run's dispatch loop) isn't keeping up, and blocking the decode
// loop to wait for it would just stall reading the terminal too. A
// drop used to be completely silent -- no log, no counter -- which
// made a fast MouseDrag burst outrunning a 16-slot buffer under
// OnDemand render mode (see examples/mouse-paint-demo) look like
// unexplained jerky/broken painting with no diagnostic trail. Now it's
// a Debug-level log naming what got dropped and the buffer's capacity,
// cheap enough to leave in thanks to Logger's own severity-gated,
// non-blocking send (see framework.Logger's doc comment) -- this can't
// itself become a second thing stalling the decode loop.
func (m *Manager) send(e framework.Event) {
	select {
	case m.events <- e:
	default:
		if m.logger.Enabled(core.Debug) {
			m.logger.Debug(fmt.Sprintf("dropped %s event: input buffer full (cap %d)", e.Kind, cap(m.events)))
		}
	}
}
