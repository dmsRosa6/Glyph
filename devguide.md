# glyph — Developer Guide

A terminal UI framework in Go. This explains what each package does and
how a frame gets from "widget tree" to "pixels on screen." It describes
the code as it stands — see `README.md`'s Roadmap section for what's
deliberately not built yet.

## Package layout

### core — primitive data, no framework knowledge

- `Color` — RGB + `IsTransparent`. `Transparent` is the "inherit from
  parent" sentinel used throughout style resolution.
- `Cell` — one terminal cell (rune + fg/bg). Plain value type.
- `Buffer` — a `W*H` grid of `Cell` (flat, row-major), plus a clip-rect
  stack (`PushClip`/`PopClip`) that `Set` respects. Everything draws
  into this. `Cells()` exposes the flat slice for `render.Flush`'s
  frame diffing.
- `AppLog`/`Severity` — a log record and level (Debug/Info/Warning/
  Fatal). `Severity`'s zero value is `Warning`, not `Debug`, so an
  unset `AppConfig.LogLevel` defaults to something reasonable.
- `AppSignal` — NOOP/SIGTERM, passed between `App` and `FaultManager`.

### geom — Point, Vector, Bounds, Axis

`Bounds.Valid()` rejects negative position/width/height. Most
constructors do their own bounds checks rather than calling this
directly.

### framework — the contracts everything implements against

- `Drawable` — `Draw`, `IsInBounds`, `SetLayer`/`GetLayer`,
  `SetParentStyle`, `SetContext`. Every shape satisfies this.
- `Focusable` (`Drawable` + `HandleInput`/`Focus`/`Blur`/`IsFocused`),
  `FocusContainer` (`Focusable` + `FocusableChildren()` for drill-in),
  `ChildrenLister`, `Identifiable` (has `ID()`), `Raisable` (accepts a
  raise-to-front callback), `Composable` (`AddChild`/`RemoveChild`).
- `Style`/`ResolveStyle` — style inheritance. `core.Transparent` is the
  only "inherit" sentinel; any other color (including Black) is literal.
- `Anchor`/`AxisAnchor`/`ResolveAxis` — Start/Center/End/NoAnchor.
- `LayoutPolicy` — `FreeLayout` (trust each child's declared position)
  and `StackLayout` (stack top-to-bottom). `CapacityAwareLayout` is an
  optional extension a policy can implement so a container can also
  warn when a child won't fit, not just when it's out of bounds.
- `Event`/`Key` — `KeyRune`, arrows, `KeyEnter`/`KeyEsc`/`KeyTab`/
  `KeyCtrlC`, `KeyBackspace`/`KeyDelete`/`KeyHome`/`KeyEnd`.
  `IsStructuralKey` marks Ctrl+C/Enter/Tab/Esc, which `App.Run`
  dispatches global-first — see `app` below.
- `AppContext` — what a node gets via `SetContext`: `Redraw()`, `Log()`,
  a `Registry`, a focus `Navigator`, `Lifecycle()`, `GlobalKeyBound(k)`.
- `Registry` — concurrent id → `Drawable` map, `FindAs[T]` for lookup.
- `Logger` — thin wrapper over `chan<- core.AppLog`. Nil-safe no-op by
  default, so a widget built before attaching to a `Canvas` never
  blocks or panics.
- `SpinnerContext` — animation-cycle state behind `primitive.Spinner`.

### mixin — reusable behavior every widget is built from

Three composition shapes recur: **leaf** (`Node`, optionally
`+FocusBehavior`), **is-a-container** (`canvas.Container` and anything
built on it — `Panel`, `Bordered`, `TileGrid`, `List`), and
**focusable-wrapper-around-one-inner-container** (`Window`,
`FocusableBox`, `ListRow`). This package holds the mixins behind shapes
1 and 3; `canvas.Container` is shape 2.

- `Node` — bounds, anchor, computed position, style, layer, `AppContext`.
  Implements `Layout`, `Style`/`ResolvedStyle`, `IsInBounds`,
  `Invalidate`, `Fault`/`Warn`.
- `FocusableNode` — `Node` + `focused bool` + a bound key-action map
  (`BindAction`/`HandleInput`) + `Focus`/`Blur`. Overrides `Style()` to
  blend a focus tint. Composition shape 1's focusable variant —
  `primitive.Button` is the example. **Gotcha:** `Node.ResolvedStyle()`
  calls `Style()` on its own static receiver, so it does NOT pick up
  `FocusableNode`'s override — call `.Style()` directly for the
  focus-tinted version.
- `FocusableWrapper` — composition shape 3: an outer `Node`+
  `FocusBehavior` wrapping exactly one inner `ContainerLike`
  (`Drawable`+`Composable`+`ChildrenLister`+`Resize` — `*canvas.Container`
  and `*widgets.Bordered` both qualify). Gives `Style`, `Draw`,
  `AddChild`/`RemoveChild`/`Children`, `SetParentStyle`, `SetContext`,
  `Resize`, `FocusableChildren` for free. `Window`, `FocusableBox`,
  `ListRow` are all built on this; a composite needing more (e.g.
  `Window`'s title) overrides just that method, using `Inner()`.
- `TextBuffer` — a mutex-protected `[]rune` behind `Text` (and the
  future `TextInput`). Not used by `StaticText`, which never changes
  after construction and needs no lock.
- `Propagator` — fans `SetParentStyle`/`SetContext` out to owned
  `Drawable`s. Also: registers/unregisters `Identifiable` children in
  the `Registry`; warns if a child binds a structural key that's also
  bound globally (it would never fire); wires `Raisable` children to
  `BringToFront`/`SendToBack`; calls `Stop()` on any removed
  `Stoppable` child. Used by `canvas.Container`.

### primitive — leaf drawables (no children)

Package placement is by composition shape (leaf vs. holds-children),
not visual complexity — a self-animating `Spinner` and a focusable
`Button` are exactly as much "primitive" as a plain `Rect`.

- `Rect` — filled rectangle.
- `Border` — box outline (`BorderStyle`: corners + edges as runes).
- `StaticText` — immutable, single-line, set once at construction, no
  lock.
- `Text` — mutable, single-line, built on `mixin.TextBuffer`. `\n`
  draws as a literal glyph, not a line break.
- `MultilineText` — mutable, wrapping, fixed-size. `\n` is a hard
  break; each line word-wraps to fit width; content past the declared
  height is TRUNCATED (not scrolled) with a `Warning` logged once per
  overflow — including the constructor's own value, via a check in
  `SetContext` (a `Warning` before `ctx` exists would otherwise
  silently no-op). `Overflowed()` exposes the same condition.
- `Button` — focusable, filled + centered label. `OnActivate` binds to
  Space, not Enter, to avoid shadowing a global Enter binding.
- `Spinner` — ticks its `SpinnerContext` on a goroutine (started on
  first `SetContext`), calls `Invalidate()` each frame. `Stop()` ends
  it; also called automatically by `Propagator.Untrack` on removal.
- `TextInput` — focusable, editable, single-line. Does not embed
  `TextBuffer` (a raw `SetValue` mid-edit would strand the cursor);
  mutation goes through cursor-aware methods bound to Backspace/
  Delete/Left/Right/Home/End/plain runes. Fixed width; content scrolls
  horizontally to keep the cursor visible (`scrollOffset`, recomputed
  fresh every `Draw`). `OnSubmit` binds Enter directly — deliberately
  taking the shadow-key risk `Button` avoids, since "Enter submits" is
  a near-universal text-field convention.
- `PaletteNode` — grid of independently-colored cells; backs
  `widgets.TileGrid`.

### canvas — the single composition primitive

- `Container` — `mixin.Node` + `mixin.Propagator`. Holds children via
  `Propagator`, sorts by layer before each `Draw`, delegates
  arrangement to a `LayoutPolicy`. Out-of-bounds/fit checks happen in
  `Draw`, with a `warned` set so a standing condition logs once, not
  every frame. `ScrollY`/`SetScrollY` — a vertical scroll offset
  (mechanism, zero by default): `Draw` shifts where children land
  without moving the clip rectangle, so content scrolls within a fixed
  viewport. `widgets.List`'s `Scrollable` mode is the only consumer.
- `Canvas` — the tree root. Wraps one root `Container` sized to the
  terminal (or fixed). `Compose()` clears + draws. `CollectFocusable`
  walks the tree gathering `Focusable`s for `input.FocusManager`.

### widgets — composites (things that hold children)

- `Panel` — styled filled rectangle you can add children to.
- `Bordered` (`NewBox`) — a `Border` around a `Panel`, inset by
  thickness + padding.
- `Window` — `FocusableWrapper` wrapping a `Bordered`, plus an
  optional title and `Raisable` (raises to front on focus).
- `FocusableBox` — `Bordered` + focus behavior, built on
  `FocusableWrapper`, no overrides needed.
- `List` — stacked-layout container; `AddItem` builds a `ListRow`.
  Fixed by default (overflow clips). `ListConfig.Scrollable` opts into
  auto-scroll-to-keep-focus-visible: `List.Draw` runs `followFocus`
  before delegating, nudging `ScrollY` if the focused row isn't fully
  visible. `List` isn't `Focusable` itself, so it binds no scroll key
  or wheel — `ScrollBy`/`ScrollTo` are the manual escape hatch.
- `ListRow` — focusable, borderless row, built on `FocusableWrapper`.
- `TileGrid` — grid of colored cells; `Panel` with its `Rect` fill
  swapped for a `PaletteNode`.

### input — bytes → events, and focus tracking

- `Manager` — raw mode, decodes stdin (including ANSI escape
  sequences: arrows, and Home/End/Delete across every terminal
  convention in use — xterm's `ESC[H`/`ESC[F`, vt220's `ESC[1~`/
  `ESC[4~`, rxvt's `ESC[7~`/`ESC[8~`) into `framework.Event`s on a
  buffered channel. Drops events under backpressure rather than
  blocking the read loop (logged at Debug). Reads through an
  unexported `byteSource` so tests can script a byte sequence without
  a real tty (`manager_test.go`).
- `FocusManager`/`FocusScope` — a stack of focus scopes. `Next`/`Prev`
  cycle the current scope; `Enter` pushes a new scope into a
  `FocusContainer` without blurring the parent; `Exit` pops back out.

### render — the frame loop

- `Renderer` — fixed-FPS ticker or on-demand redraws. Handles resize
  via `term.WatchResize`. `Flush` diffs against the previous frame and
  only writes what changed, batching same-style runs into one escape.
  `safeCompose` recovers a panic anywhere in `Draw` and reports it as
  Fatal rather than crashing the process mid-raw-mode.
- `RenderMode` — unexported fields; build via `FixedFPSMode(fps)` or
  `OnDemandMode()`, never a literal (a zero-value one would panic or
  hang). `NewRenderer` validates regardless.

### term — raw terminal I/O

`raw.go` (raw mode, signal-safe restore), `term.go` (`TermSize`,
`WatchResize`), `ansi.go` (`CellToANSI`).

### fault — logging and crash handling

- `FaultManager` — owns the log channel, writes to a timestamped file,
  retries failed writes via a `RingBuffer` on a 1s ticker. A `Fatal`
  log writes immediately and sends `SIGTERM`, but keeps draining
  afterward so `App.Stop()`'s own shutdown logs still land.
  `Config{}`'s zero value is a safe default.

### datastructs

`RingBuffer` — fixed-capacity string ring, oldest-drops-on-overflow.
Capacity must be >= 2 (a single-index scheme can't tell full from
empty otherwise); any capacity `N` holds `N-1` usable slots.

### app — top-level wiring

- `App` — owns `Canvas`, `Renderer`, `Input`, `Focus`, `FaultManager`.
  `NewApp` seeds Ctrl+C to quit unconditionally, then layers
  `cfg.AppEvents` on top. `BindKey`/`BindMouse`/etc. are safe to call
  from any goroutine (`bindingsMu` guards them).

  `Run()`'s dispatch (`handleEvent`) splits on `IsStructuralKey`:
  **structural keys** (Ctrl+C/Enter/Tab/Esc) go **global-first**;
  **everything else** goes **widget-first**, global as fallback. This
  asymmetry is exactly what `Propagator`'s shadow-key warning protects
  against.

  Logging goes through the same `framework.Logger` widgets use.
  `Debug`/`Info` are non-blocking and severity-filtered at the source;
  `Warning`/`Fatal` block deliberately (rare, and `Fatal` is what
  triggers a clean shutdown — dropping one under backpressure would
  disable that escape hatch when something's already wrong).

  **Panic recovery.** A panic on any goroutine kills the whole process,
  which could leave the terminal stuck in raw mode. Two boundaries
  recover from this: `Renderer.safeCompose` (any `Draw`) and
  `mixin.FocusBehavior.invoke`/`App.invokeAction` (bound key/mouse
  handlers). Both report the panic as `Fatal`, reusing the existing
  Fatal → `SIGTERM` → clean shutdown path. Deliberately NOT covering
  every interface method a custom widget could implement (`SetContext`,
  `AddChild`, etc.) — those run far less often and rarely contain
  application logic; `Draw` and bound handlers are where real risk is.

## How a frame actually happens

1. `App.Run()` blocks on `appSignals` and `Input.Events()`.
2. An input event dispatches per the structural/other split above.
3. A handler mutates state and calls `Invalidate()`, or returns
   `redraw=true` from a global action.
4. `Invalidate()` is `n.ctx.Redraw()` — every node holds its own
   `AppContext`, fanned out at attach time, so this calls straight
   through to `Renderer.RequestRedraw`. No walking up an ancestor chain.
5. `Renderer.render` calls `Canvas.Compose()` then `Flush`es to the
   terminal. (`OnDemand`: only on step 4. `FixedFPS`: every tick.)

## Style/context propagation, in one sentence

A parent hands a child its resolved style and `AppContext` either at
`AddChild` time or the next time the parent itself receives one from
further up — `Propagator` guarantees this regardless of add-order, and
piggybacks registry registration and the shadow-key warning onto the
same moment.
