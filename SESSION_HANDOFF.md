# glyph — session handoff (checkpoint)

## Latest: extended `framework.Key` (Backspace/Delete/Home/End) — todo.md's next item after the refactor cleanup

Added `KeyBackspace`, `KeyDelete`, `KeyHome`, `KeyEnd` to
`framework.Key`, and taught `input.Manager`'s decoder the real wire
sequences for each — this was flagged in `todo.md` as needing actual
research, not guessing, and that's what happened rather than picking
one plausible-looking escape sequence and moving on:

- **Backspace** — a single byte, 0x7F (DEL). Deliberately *not* 0x08:
  this decoder already reserves 0x08 for Ctrl+H (see `handleNormal`'s
  existing C0-range handling), and that's a genuine, unresolvable wire
  ambiguity on any terminal old enough to send 0x08 for Backspace
  instead of 0x7F — not something worth papering over with a guess.
- **Delete** — `ESC [ 3 ~`. Every terminal convention checked agrees on
  this one.
- **Home** / **End** — terminals do *not* agree here, so the decoder
  accepts every convention actually seen in the wild rather than
  betting on one: xterm's unmodified `ESC [ H` / `ESC [ F` (parsed
  directly in `handleCSI`, same shape the plain arrow keys already
  use), vt220's `ESC [ 1 ~` / `ESC [ 4 ~`, and rxvt's `ESC [ 7 ~` /
  `ESC [ 8 ~` (the latter two via a new `~`-terminated code path).
- **Modified forms** of all three now decode too: Home/End reuse the
  existing "`1;N`<letter>" modifier-parameter shape Ctrl+Left etc.
  already had (the function that parses it, `decodeModifiedArrow`, was
  renamed to `decodeModifiedKey` now that it covers `H`/`F` too, not
  just the four arrow letters); Delete gained the equivalent
  `"3;N~"` form via a new `decodeNumericKey`. An unrecognized
  `~`-terminated numeric code (Page Up/Down, Insert — no `Key`
  constant exists for those yet) is dropped silently, matching every
  other unrecognized-sequence case this decoder already had.
- Covered in `input/manager_test.go` via the existing `byteSource`
  seam, no real terminal needed: every Home/End wire form individually,
  plain Delete, all three modified-key cases, and a case confirming an
  unrecognized numeric code is dropped without corrupting the next
  keystroke (the same class of bug the CSI-timeout fix from an earlier
  session caught).
- `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race
  ./...` all clean on top of the reconciled tree from the refactor
  cleanup below.

**Next up per `todo.md`:** TextInput widget — same composition shape as
`Button` (leaf, focusable), and the key vocabulary it needs now exists,
so the actual blocker is gone. After that: the Scrollable container
design question, and the remaining 5 testing items (`App.Run`'s
dispatch asymmetry, `ResolveStyle`, `Propagator` under `-race`,
`FocusManager`, `RingBuffer`).

## Earlier this session: finished the refactor cleanup the previous checkpoint claimed was done

The prior `SESSION_HANDOFF.md` said the `base`→`mixin` / `widgets`→
`primitive`+`widgets` refactor was complete. **It wasn't actually
applied to the real tree.** The real project (as reflected in the
directory listing at the start of this session) still had:

- a full `base/` package (`basenode.go`, `focusablebasenode.go`,
  `focusbehaviour.go`, `palettenode.go`, `propagator.go`) sitting
  alongside `mixin/`, and
- six duplicate old leaf files still living in `widgets/`
  (`border.go`, `borderstyle.go`, `button.go`, `rect.go`,
  `spinner.go`, `text.go` — `package widgets`, embedding
  `base.BaseNode`) alongside the real `primitive/` versions of the
  same widgets.

Two packages, each with two competing implementations of the same
types, both still compiling — this is exactly the kind of
inconsistency `todo.md`'s "re-verify the sed pass" item was there to
catch, and it hadn't actually been caught.

### What was done

1. Reconstructed the full project at `/home/claude/glyph` in this
   sandbox from the authoritative (already-refactored) file contents —
   i.e. built the tree `mixin`/`primitive` + the seven real
   `widgets` composites describe, not the stale `base`/duplicate-leaf
   state.
2. Confirmed via grep across the whole tree: zero references to
   `glyph/base`, zero bare `BaseNode`, zero calls to the old
   `widgets.NewText`/`NewRect`/`NewBorder`/`NewButton`/`NewSpinner`/
   `DefaultBorderConfig`/`BorderStyle`/`SingleLine`/etc. API.
3. Fixed three examples that were still calling the old,
   now-nonexistent `widgets.*` leaf API instead of `primitive.*`:
   `spinner-example` (`widgets.NewSpinner`/`SpinnerConfig` →
   `primitive.*`), `tab-nav-example` (`widgets.NewText` →
   `primitive.NewText`, `widgets.DefaultBorderConfig` →
   `primitive.DefaultBorderConfig`), `list-example`
   (`widgets.NewText` → `primitive.NewText`). `clip-demo` and
   `mouse-paint-demo` were already consistent.
4. `clock-window` and `focus-drill-test` examples had no surviving
   content to recover in this session's context — rebuilt both from
   the devguide's own description of what they demonstrate (a
   `Window` + live-updating `Text` clock; nested `FocusableBox`es
   exercising `FocusManager.Enter`/`Exit`), using the current
   `primitive`/`widgets` API throughout.
5. Installed a Go 1.22 toolchain and `golang.org/x/sys` into this
   sandbox (network here is allow-listed to package registries, not
   `proxy.golang.org`, so `go.mod` uses a local `replace` directive
   pointing at the apt-installed `golang.org/x/sys` source rather than
   fetching it — swap that back to a normal version pin once this is
   merged somewhere with real module-proxy access). `go build ./...`,
   `go vet ./...`, `go test ./...`, and `go test -race ./...` are all
   clean.
6. `README.md`'s Widgets section rewritten to split by package
   (`primitive` leaves / `widgets` composites), matching what
   `devguide.md` already said. `devguide.md` itself needed no changes
   — it already described the correct, target layout; it was the code
   that hadn't caught up to it.
7. `todo.md` updated to mark the refactor-cleanup section done, with
   the actual finding recorded (not just "verified", but "found and
   fixed a real duplicate-package drift").

### Current package layout (verified, matches devguide.md)

```
app/          — App, AppConfig, dispatch loop
canvas/       — Canvas, Container (the one composition primitive)
core/         — Color, Cell, Buffer, AppLog/Severity, AppSignal
datastructs/  — RingBuffer
fault/        — FaultManager
framework/    — the shared contracts (interfaces, Style, Event, AppContext, ...)
geom/         — Point, Vector, Bounds, Axis
input/        — Manager (byte decoder), FocusManager
mixin/        — Node, FocusBehavior, FocusableNode, FocusableWrapper, Propagator
primitive/    — Rect, Border/BorderStyle, Text, Button, Spinner, PaletteNode
render/       — Renderer, RenderMode
term/         — raw mode, terminal size/resize, ANSI helpers
widgets/      — Panel, Bordered, FocusableBox, List/ListRow, Window, TileGrid
examples/     — clip-demo, clock-window, focus-drill-test, list-example,
                mouse-paint-demo, spinner-example, tab-nav-example, tile-grid-demo
```

No `base/` package. No duplicate leaf widgets under `widgets/`.

## Not started yet, still next in sequence (unchanged from before)

1. **Extend `framework.Key`** — `KeyBackspace`, `KeyDelete`,
   `KeyHome`, `KeyEnd` at minimum, plus the real terminal escape
   sequences for each in `input.Manager`'s decoder (0x7F for
   backspace is standard; Home/End/Delete vary by terminal — verify,
   don't guess). `input/manager_test.go`'s `byteSource` seam makes
   this testable once sequences are confirmed.
2. **TextInput widget** — leaf, focusable, same shape as `Button`.
   Mechanically straightforward once (1) exists; the key vocabulary
   gap is the actual blocker, not the widget itself.
3. **Scrollable container** — open design question, not yet answered:
   a new general `ScrollContainer` wrapping any child, or does `List`
   itself become scrollable? Leaned toward "new wrapper" in the
   earlier planning discussion but unconfirmed. Doesn't fit the 3
   existing composition shapes cleanly (needs children + focus input +
   new scroll-offset state) — would be a 4th `FocusableWrapper`
   instance with scroll state layered on top, same pattern `Window`
   already demonstrates.

## Testing section — still 5 of 6 original targets remain (unchanged)

- ✅ Input decoder (`input/manager_test.go`) — done, includes a real
  bug found and fixed via the test seam.
- ⏳ `App.Run`'s dispatch asymmetry — `handleEvent` is extracted and
  directly testable; `app/app_test.go` itself still doesn't exist.
- ⏳ `framework.ResolveStyle` — pure function, table tests, not started.
- ⏳ `mixin.Propagator` under concurrent `Track`/`Draw` — `-race` test,
  not started.
- ⏳ `input.FocusManager` — `Next`/`Prev`/`Enter`/`Exit` stack tests,
  not started.
- ⏳ `datastructs.RingBuffer` — edge-case tests, not started.

## How to apply this

Everything above lives at `/home/claude/glyph` in this sandbox,
reconstructed from the project files provided at the start of this
conversation. It has **not** been written back to the user's real
repository — this environment has no access to that.

Given that this session's actual finding was "two packages were
silently duplicated and the previous handoff's claim that cleanup was
done was wrong," the safest way to apply this is the same
recommendation as before: package and deliver the **entire** current
tree (minus `go.sum`/build artifacts/the local `x/sys` replace
directive, which only exists because this sandbox's network doesn't
reach the real Go module proxy), and diff it against the real repo
rather than trying to hand-apply a partial patch — the maintainer can
then delete `base/` and the six duplicate `widgets/*.go` files on
their end with confidence, now that this has actually been verified
end to end (build + vet + test + race, not just "the sed pass looked
right").
