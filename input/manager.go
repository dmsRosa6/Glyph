package input

import (
	"context"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/term"
)

// decodeState tracks how far into a multi-byte escape sequence we are.
type decodeState int

const (
	stateNormal     decodeState = iota
	stateEsc                    // just saw 0x1b, waiting to see if more follows
	stateEscBracket             // saw ESC [, waiting for the final letter
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
				m.send(framework.Event{Key: framework.KeyEsc})
				state = m.handleNormal(ch)
			}

		case stateEscBracket:
			switch ch {
			case 'A':
				m.send(framework.Event{Key: framework.KeyUp})
			case 'B':
				m.send(framework.Event{Key: framework.KeyDown})
			case 'C':
				m.send(framework.Event{Key: framework.KeyRight})
			case 'D':
				m.send(framework.Event{Key: framework.KeyLeft})
			default:
				// unrecognized escape sequence — drop it silently
			}
			state = stateNormal
		}
	}
}

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
		m.send(framework.Event{Key: framework.KeyRune, Rune: rune(ch)})
	}
	return stateNormal
}

func (m *Manager) send(e framework.Event) {
	select {
	case m.events <- e:
	default:
	}
}
