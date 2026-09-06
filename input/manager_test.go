package input

import (
	"io"
	"reflect"
	"testing"

	"github.com/dmsRosa6/glyph/framework"
)

// step is one scripted result fed to a test byteSource: either a
// single byte, or a "timeout" (n=0, err=nil) -- exactly what
// term.ReadStdin returns when its ~100ms VTIME elapses with nothing
// typed (see term.EnableRawMode's doc comment). Modeling timeouts
// explicitly, not just concatenating bytes, is what lets these tests
// reach states real typing can actually produce -- a lone ESC with no
// follow-up, or a CSI sequence abandoned mid-parameter -- without
// needing a real tty at all.
type step struct {
	b       byte
	timeout bool
}

// bytesOf turns a plain string into one step per byte -- the common
// case for scripting a sequence of keys/escape codes.
func bytesOf(s string) []step {
	steps := make([]step, len(s))
	for i := 0; i < len(s); i++ {
		steps[i] = step{b: s[i]}
	}
	return steps
}

func timeout() step { return step{timeout: true} }

func concat(groups ...[]step) []step {
	var out []step
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

// scriptSource is the byteSource this whole test file substitutes for
// stdinSource: it replays a fixed sequence of steps, then reports
// io.EOF -- a real read error, which is exactly what ends Manager.run's
// loop in production too (see run's `if err != nil { return }`).
type scriptSource struct {
	steps []step
	i     int
}

func (s *scriptSource) Read(buf []byte) (int, error) {
	if s.i >= len(s.steps) {
		return 0, io.EOF
	}
	st := s.steps[s.i]
	s.i++
	if st.timeout {
		return 0, nil
	}
	buf[0] = st.b
	return 1, nil
}

// decode runs Manager.run() to completion against a scripted byte
// sequence and returns every framework.Event it emitted, in order.
// run() is called directly, not Start() -- Start() would try to put a
// real tty into raw mode, which the byteSource seam exists specifically
// to avoid ever needing in a test.
func decode(t *testing.T, steps []step) []framework.Event {
	t.Helper()

	m, err := NewManager(framework.Logger{}, 64)
	if err != nil {
		t.Fatal(err)
	}
	m.source = &scriptSource{steps: steps}

	m.run()

	var got []framework.Event
	for e := range m.events {
		got = append(got, e)
	}
	return got
}

func TestDecodePlainKeys(t *testing.T) {
	cases := []struct {
		name  string
		steps []step
		want  []framework.Event
	}{
		{
			name:  "plain runes",
			steps: bytesOf("ab"),
			want: []framework.Event{
				{Key: framework.KeyRune, Rune: 'a'},
				{Key: framework.KeyRune, Rune: 'b'},
			},
		},
		{
			name:  "ctrl+c",
			steps: []step{{b: 0x03}},
			want:  []framework.Event{{Key: framework.KeyCtrlC}},
		},
		{
			name:  "enter (CR)",
			steps: []step{{b: '\r'}},
			want:  []framework.Event{{Key: framework.KeyEnter}},
		},
		{
			name:  "enter (LF)",
			steps: []step{{b: '\n'}},
			want:  []framework.Event{{Key: framework.KeyEnter}},
		},
		{
			name:  "tab",
			steps: []step{{b: '\t'}},
			want:  []framework.Event{{Key: framework.KeyTab}},
		},
		{
			name:  "ctrl+h decodes as the letter with ModCtrl, not a bare control byte",
			steps: []step{{b: 0x08}},
			want:  []framework.Event{{Key: framework.KeyRune, Rune: 'h', Modifiers: framework.ModCtrl}},
		},
		{
			name:  "ctrl+a is the low end of the C0 range",
			steps: []step{{b: 0x01}},
			want:  []framework.Event{{Key: framework.KeyRune, Rune: 'a', Modifiers: framework.ModCtrl}},
		},
		{
			name:  "lone ESC, abandoned by a timeout with nothing following",
			steps: []step{{b: 0x1b}, timeout()},
			want:  []framework.Event{{Key: framework.KeyEsc}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decode(t, tc.steps)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestDecodeArrowKeys(t *testing.T) {
	cases := []struct {
		name string
		seq  string
		want framework.Key
	}{
		{"up", "\x1b[A", framework.KeyUp},
		{"down", "\x1b[B", framework.KeyDown},
		{"right", "\x1b[C", framework.KeyRight},
		{"left", "\x1b[D", framework.KeyLeft},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decode(t, bytesOf(tc.seq))
			want := []framework.Event{{Key: tc.want}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

func TestDecodeShiftTab(t *testing.T) {
	got := decode(t, bytesOf("\x1b[Z"))
	want := []framework.Event{{Key: framework.KeyTab, Modifiers: framework.ModShift}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestDecodeModifiedArrows(t *testing.T) {
	cases := []struct {
		name string
		seq  string
		want framework.Event
	}{
		{"ctrl+left", "\x1b[1;5D", framework.Event{Key: framework.KeyLeft, Modifiers: framework.ModCtrl}},
		{"shift+right", "\x1b[1;2C", framework.Event{Key: framework.KeyRight, Modifiers: framework.ModShift}},
		{"alt+up", "\x1b[1;3A", framework.Event{Key: framework.KeyUp, Modifiers: framework.ModAlt}},
		{"ctrl+shift+down", "\x1b[1;6B", framework.Event{Key: framework.KeyDown, Modifiers: framework.ModShift | framework.ModCtrl}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decode(t, bytesOf(tc.seq))
			want := []framework.Event{tc.want}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

func TestDecodeSGRMouse(t *testing.T) {
	cases := []struct {
		name string
		seq  string
		want framework.Event
	}{
		{
			name: "left press at (10,5)",
			seq:  "\x1b[<0;11;6M",
			want: framework.Event{
				Kind: framework.EventKindMouse, MouseButton: framework.MouseButtonLeft,
				MouseAction: framework.MousePress, MouseX: 10, MouseY: 5,
			},
		},
		{
			name: "left release at (10,5)",
			seq:  "\x1b[<0;11;6m",
			want: framework.Event{
				Kind: framework.EventKindMouse, MouseButton: framework.MouseButtonLeft,
				MouseAction: framework.MouseRelease, MouseX: 10, MouseY: 5,
			},
		},
		{
			name: "left drag (motion flag set)",
			seq:  "\x1b[<32;11;6M",
			want: framework.Event{
				Kind: framework.EventKindMouse, MouseButton: framework.MouseButtonLeft,
				MouseAction: framework.MouseDrag, MouseX: 10, MouseY: 5,
			},
		},
		{
			name: "wheel up",
			seq:  "\x1b[<64;1;1M",
			want: framework.Event{
				Kind: framework.EventKindMouse, MouseButton: framework.MouseButtonNone,
				MouseAction: framework.MouseWheelUp, MouseX: 0, MouseY: 0,
			},
		},
		{
			name: "wheel down",
			seq:  "\x1b[<65;1;1M",
			want: framework.Event{
				Kind: framework.EventKindMouse, MouseButton: framework.MouseButtonNone,
				MouseAction: framework.MouseWheelDown, MouseX: 0, MouseY: 0,
			},
		},
		{
			name: "right press with ctrl+shift held (cb = 2 | 4 | 16 = 22)",
			seq:  "\x1b[<22;1;1M",
			want: framework.Event{
				Kind: framework.EventKindMouse, MouseButton: framework.MouseButtonRight,
				MouseAction: framework.MousePress, MouseX: 0, MouseY: 0,
				Modifiers: framework.ModShift | framework.ModCtrl,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decode(t, bytesOf(tc.seq))
			want := []framework.Event{tc.want}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

// TestDecodeMalformedSequencesAreDroppedNotMisread covers the
// "unrecognized/malformed/truncated" branches the decoder itself
// documents as silently-drop cases -- and confirms each one actually
// recovers to stateNormal afterward instead of corrupting how the next
// real keystroke decodes.
func TestDecodeMalformedSequencesAreDroppedNotMisread(t *testing.T) {
	cases := []struct {
		name  string
		steps []step
		want  []framework.Event
	}{
		{
			name:  "unrecognized CSI first byte is dropped, decoding resumes cleanly",
			steps: concat(bytesOf("\x1b[Q"), bytesOf("a")),
			want:  []framework.Event{{Key: framework.KeyRune, Rune: 'a'}},
		},
		{
			name:  "SGR mouse body with the wrong number of fields is dropped, not sent",
			steps: concat(bytesOf("\x1b[<1;2;3;4M"), bytesOf("a")),
			want:  []framework.Event{{Key: framework.KeyRune, Rune: 'a'}},
		},
		{
			name:  "garbage byte mid mouse-parameter sequence is dropped",
			steps: concat(bytesOf("\x1b[<12X"), bytesOf("a")),
			want:  []framework.Event{{Key: framework.KeyRune, Rune: 'a'}},
		},
		{
			name: "a CSI sequence abandoned by a timeout does not leak into the next byte",
			// ESC [ 1 <timeout> -- before the fix documented on run()'s
			// n==0 branch, `state` stayed stuck at stateCSIParams
			// across the timeout, so this next 'A' got fed into
			// handleCSIParams as if it were still completing "ESC [ 1"
			// -- misread as a bare KeyUp using the stale csiBuf, rather
			// than the plain 'A' rune it actually is.
			steps: concat(bytesOf("\x1b[1"), []step{timeout()}, bytesOf("A")),
			want:  []framework.Event{{Key: framework.KeyRune, Rune: 'A'}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decode(t, tc.steps)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestDecodeMouseNeverMisreadAsKeyRuneZero guards the exact invariant
// app.go's dispatch loop depends on: a mouse Event's Key field sits at
// its zero value (KeyRune), so mouse events MUST be distinguishable via
// Kind alone, never by Key.
func TestDecodeMouseNeverMisreadAsKeyRuneZero(t *testing.T) {
	got := decode(t, bytesOf("\x1b[<0;1;1M"))
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
	if got[0].Kind != framework.EventKindMouse {
		t.Fatalf("Kind = %v, want EventKindMouse", got[0].Kind)
	}
	if got[0].Key != framework.KeyRune {
		t.Fatalf("Key = %v, want the zero value KeyRune (Kind is what must be checked, not Key)", got[0].Key)
	}
}
