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

type decodeState int

const (
	stateNormal decodeState = iota
	stateEsc
	stateCSI
	stateCSIParams
)

const DefaultEventBufferSize = 16

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
		m.send(framework.Event{Key: framework.KeyTab, Modifiers: framework.ModShift})
		return stateNormal
	case 'H':
		// ESC [ H -- xterm's unmodified Home. Modified Home (e.g.
		// Shift+Home, "1;2H") starts with a digit and is handled by
		// the parameterized path below via decodeModifiedKey, exactly
		// like Ctrl+Left already is.
		m.send(framework.Event{Key: framework.KeyHome})
		return stateNormal
	case 'F':
		// ESC [ F -- xterm's unmodified End. See 'H' above for the
		// modified case.
		m.send(framework.Event{Key: framework.KeyEnd})
		return stateNormal
	case '<':
		*csiBuf = append(*csiBuf, ch)
		return stateCSIParams
	default:
		if ch >= '0' && ch <= '9' {
			*csiBuf = append(*csiBuf, ch)
			return stateCSIParams
		}
		return stateNormal
	}
}

func (m *Manager) handleCSIParams(ch byte, csiBuf *[]byte) decodeState {
	isMouse := len(*csiBuf) > 0 && (*csiBuf)[0] == '<'

	if isMouse && (ch == 'M' || ch == 'm') {
		m.decodeMouseSequence(string((*csiBuf)[1:]), ch == 'M')
		return stateNormal
	}

	if !isMouse {
		switch ch {
		case 'A', 'B', 'C', 'D', 'H', 'F':
			// Modified arrow, or modified Home/End ("1;2H" for
			// Shift+Home, "1;5F" for Ctrl+End) -- same "1;N<letter>"
			// shape xterm already uses for Ctrl+Left etc.
			m.decodeModifiedKey(string(*csiBuf), ch)
			return stateNormal
		case '~':
			// The other family of xterm-ish sequences: a bare numeric
			// code (optionally followed by ";N" for a modifier),
			// terminated by '~' instead of a letter -- how Delete,
			// and vt220/rxvt's own Home/End, arrive on the wire.
			m.decodeNumericKey(string(*csiBuf))
			return stateNormal
		}
	}

	if (ch >= '0' && ch <= '9') || ch == ';' {
		*csiBuf = append(*csiBuf, ch)
		return stateCSIParams
	}

	return stateNormal
}

func (m *Manager) decodeMouseSequence(body string, press bool) {
	parts := strings.Split(body, ";")
	if len(parts) != 3 {
		return
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
		ev.MouseButton = framework.MouseButtonNone
		if cb&3 == 0 {
			ev.MouseAction = framework.MouseWheelUp
		} else {
			ev.MouseAction = framework.MouseWheelDown
		}
	case cb&32 != 0:
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

// decodeModifiedKey parses the "N" or "N;M" parameter body of a
// modified xterm CSI sequence terminated by a letter (e.g. "1;5" before
// a trailing 'D' for Ctrl+Left, or "1;2" before 'H' for Shift+Home) and
// sends the resulting Event. dir is the terminator byte already seen by
// the caller. Was decodeModifiedArrow -- renamed and extended to also
// cover Home/End ('H'/'F') once those needed the same "1;N<letter>"
// modifier-parameter handling arrows already had; the parsing itself
// didn't need to change, only the letter-to-Key mapping below.
func (m *Manager) decodeModifiedKey(body string, dir byte) {
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
	case 'H':
		key = framework.KeyHome
	case 'F':
		key = framework.KeyEnd
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

// decodeNumericKey parses the "N" or "N;M" parameter body of a
// '~'-terminated CSI sequence (ESC [ N ~, optionally ESC [ N ; M ~ for
// a modifier) -- the other shape terminal Delete/Home/End/etc. arrive
// in, distinct from decodeModifiedKey's letter-terminated shape above.
// Terminal convention for which numeric code means what isn't fully
// standardized: Delete is universally 3, but Home/End show up as
// either vt220's 1/4 or rxvt's 7/8 depending on the terminal, so both
// pairs are accepted here rather than picking just one and guessing
// wrong for the other family. Any other code (2 Insert, 5/6 Page Up/
// Down, anything unrecognized) is silently dropped -- same convention
// as every other unrecognized sequence in this decoder -- since this
// codebase has no Key constant for them yet.
func (m *Manager) decodeNumericKey(body string) {
	parts := strings.Split(body, ";")

	var key framework.Key
	switch parts[0] {
	case "3":
		key = framework.KeyDelete
	case "1", "7":
		key = framework.KeyHome
	case "4", "8":
		key = framework.KeyEnd
	default:
		return // unrecognized numeric code -- drop, don't guess
	}

	ev := framework.Event{Key: key}

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

func (m *Manager) handleNormal(ch byte) decodeState {
	switch ch {
	case 0x1b:
		return stateEsc
	case 0x03:
		m.send(framework.Event{Key: framework.KeyCtrlC})
	case '\r', '\n':
		m.send(framework.Event{Key: framework.KeyEnter})
	case '\t':
		m.send(framework.Event{Key: framework.KeyTab})
	case 0x7f:
		// DEL. This is what the Backspace key actually sends on
		// virtually every modern terminal (xterm, the Linux console,
		// macOS Terminal, ...) -- not 0x08, which this codebase
		// already reserves for Ctrl+H via the C0 range below. A
		// terminal old/unusual enough to send 0x08 for Backspace
		// instead would decode as Ctrl+H here, same as it always has;
		// that's a real ambiguity in the wire protocol itself, not
		// something this decoder can resolve from the byte alone.
		m.send(framework.Event{Key: framework.KeyBackspace})
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
		if m.logger.Enabled(core.Debug) {
			m.logger.Debug(fmt.Sprintf("dropped %s event: input buffer full (cap %d)", e.Kind, cap(m.events)))
		}
	}
}
