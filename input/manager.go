package input

import (
	"context"
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

type Manager struct {
	events chan framework.Event
	logs   chan<- core.AppLog

	ctx     context.Context
	cancel  context.CancelFunc
	restore func()
	stopped chan struct{}
}

func NewManager(logs chan<- core.AppLog) (*Manager, error) {

	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		events:  make(chan framework.Event, 16),
		logs:    logs,
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

func (m *Manager) Stop() {
	if m.restore == nil {
		return
	}
	m.logs <- *core.NewInfoAppLog("Input manager stopping", string(core.InputSource))
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

	m.logs <- *core.NewInfoAppLog("Input manager started", string(core.InputSource))

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		n, err := term.ReadStdin(buf[:])
		if err != nil {
			return
		}

		if n == 0 {
			if state == stateEsc {
				m.send(framework.Event{Key: framework.KeyEsc})
				state = stateNormal
			}
			// A CSI sequence that stalls mid-parameter (state == stateCSI
			// or stateCSIParams) on a read timeout is abandoned here --
			// an incomplete escape sequence has no safe single-key
			// interpretation, so it's silently dropped rather than
			// replayed as raw KeyRune events, same convention as an
			// unrecognized sequence below.
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

func (m *Manager) send(e framework.Event) {
	select {
	case m.events <- e:
	default:
	}
}
