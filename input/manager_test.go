package input

import (
	"io"
	"reflect"
	"testing"

	"github.com/dmsRosa6/glyph/framework"
)

type step struct {
	b       byte
	timeout bool
}

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

func TestDecodeBackspace(t *testing.T) {
	got := decode(t, []step{{b: 0x7f}})
	want := []framework.Event{{Key: framework.KeyBackspace}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// TestDecodeHomeEnd covers every wire form this decoder accepts for
// Home/End: xterm's unmodified "ESC[H"/"ESC[F", and the two numeric-
// code families terminals disagree on (vt220's 1/4, rxvt's 7/8) via
// "ESC[N~". A terminal only ever sends one of these per key, but the
// decoder has to accept whichever one shows up without knowing in
// advance which convention the user's terminal follows.
func TestDecodeHomeEnd(t *testing.T) {
	cases := []struct {
		name string
		seq  string
		want framework.Key
	}{
		{"xterm home", "\x1b[H", framework.KeyHome},
		{"xterm end", "\x1b[F", framework.KeyEnd},
		{"vt220 home", "\x1b[1~", framework.KeyHome},
		{"vt220 end", "\x1b[4~", framework.KeyEnd},
		{"rxvt home", "\x1b[7~", framework.KeyHome},
		{"rxvt end", "\x1b[8~", framework.KeyEnd},
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

func TestDecodeDelete(t *testing.T) {
	got := decode(t, bytesOf("\x1b[3~"))
	want := []framework.Event{{Key: framework.KeyDelete}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// TestDecodeModifiedHomeEndDelete covers the modifier-parameter form
// each family uses: letter-terminated Home/End ("1;N<H|F>", same shape
// Ctrl+Left etc. already use) via decodeModifiedKey, and '~'-terminated
// Delete ("3;N~") via decodeNumericKey.
func TestDecodeModifiedHomeEndDelete(t *testing.T) {
	cases := []struct {
		name string
		seq  string
		want framework.Event
	}{
		{"shift+home (xterm form)", "\x1b[1;2H", framework.Event{Key: framework.KeyHome, Modifiers: framework.ModShift}},
		{"ctrl+end (xterm form)", "\x1b[1;5F", framework.Event{Key: framework.KeyEnd, Modifiers: framework.ModCtrl}},
		{"ctrl+delete", "\x1b[3;5~", framework.Event{Key: framework.KeyDelete, Modifiers: framework.ModCtrl}},
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

// TestDecodeUnrecognizedNumericCodeIsDropped covers a '~'-terminated
// code this decoder has no Key for yet (5 is Page Up) -- dropped
// silently, same convention as every other unrecognized sequence, and
// decoding resumes cleanly on the next byte rather than misreading it.
func TestDecodeUnrecognizedNumericCodeIsDropped(t *testing.T) {
	got := decode(t, concat(bytesOf("\x1b[5~"), bytesOf("a")))
	want := []framework.Event{{Key: framework.KeyRune, Rune: 'a'}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
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
			name:  "a CSI sequence abandoned by a timeout does not leak into the next byte",
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
