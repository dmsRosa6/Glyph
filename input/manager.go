package input

import (
	"context"

	"github.com/dmsRosa6/glyph/term"
)

type Key int

const (
	KeyRune Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEsc
	KeyCtrlC //Reserved to kill process cleanly
)

type Event struct {
	Key  Key
	Rune rune
}

// decodeState tracks how far into a multi-byte escape sequence we are.
type decodeState int

const (
	stateNormal decodeState = iota
	stateEsc          // just saw 0x1b, waiting to see if more follows
	stateEscBracket   // saw ESC [, waiting for the final letter
)

type Manager struct {
	events  chan Event
	ctx     context.Context
	cancel  context.CancelFunc
	restore func()
	stopped chan struct{} // closed once the read loop has fully exited
}

func NewManager() (*Manager, error) {
	restore, err := term.SafeRawMode()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		events:  make(chan Event, 16),
		ctx:     ctx,
		cancel:  cancel,
		restore: restore,
		stopped: make(chan struct{}),
	}, nil
}

func (m *Manager) Events() <-chan Event { return m.events }

func (m *Manager) Start() {
	go m.run()
}

// Stop cancels the read loop, waits for it to actually exit (so it's
// no longer touching stdin), then restores the terminal.
func (m *Manager) Stop() {
	m.cancel()
	<-m.stopped
	m.restore()
}

func (m *Manager) run() {
	defer close(m.stopped)
	defer close(m.events)

	state := stateNormal
	var buf [1]byte

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		n, err := term.ReadStdin(buf[:])
		if err != nil {
			// fd closed or real read error, not just a timeout — stop.
			return
		}

		if n == 0 {
			// VTIME timeout: nothing was typed this tick. If we were
			// mid-way through decoding an escape sequence, the silence
			// means it was a lone ESC press, not the start of one.
			if state == stateEsc {
				m.send(Event{Key: KeyEsc})
				state = stateNormal
			}
			continue
		}

		ch := buf[0]

		switch state {
		case stateNormal:
			state = m.handleNormal(ch)

		case stateEsc:
			if ch == '[' {
				state = stateEscBracket
			} else {
				// ESC wasn't followed by '[', so it was a lone ESC.
				// Emit it, then process ch as a fresh normal byte —
				// it hasn't been consumed yet.
				m.send(Event{Key: KeyEsc})
				state = m.handleNormal(ch)
			}

		case stateEscBracket:
			switch ch {
			case 'A':
				m.send(Event{Key: KeyUp})
			case 'B':
				m.send(Event{Key: KeyDown})
			case 'C':
				m.send(Event{Key: KeyRight})
			case 'D':
				m.send(Event{Key: KeyLeft})
			default:
				// unrecognized escape sequence — drop it silently
			}
			state = stateNormal
		}
	}
}

// handleNormal decodes a single byte that isn't part of an escape
// sequence, and returns the next decode state (usually stateNormal,
// or stateEsc if this byte itself was 0x1b).
func (m *Manager) handleNormal(ch byte) decodeState {
	switch ch {
	case 0x1b: // ESC
		return stateEsc
	case 0x03: // Ctrl+C
		m.send(Event{Key: KeyCtrlC})
	case '\r', '\n':
		m.send(Event{Key: KeyEnter})
	default:
		m.send(Event{Key: KeyRune, Rune: rune(ch)})
	}
	return stateNormal
}

// send drops the event if the channel is full rather than blocking
// the whole read loop on a slow consumer.
func (m *Manager) send(e Event) {
	select {
	case m.events <- e:
	default:
	}
}