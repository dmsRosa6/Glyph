# glyph — Developer Guide

A terminal UI framework in Go. This doc explains what each package is
responsible for and how a frame actually gets from "widget tree" to
"pixels on screen."

This describes the code as it stands today, not where it's headed — see
`todo.md` for known gaps and planned work (scrollable containers, a
text-input widget, mouse hit-testing).

## Package layout and responsibilities

### core — primitive data, no framework knowledge

- `Color` — RGB + an `IsTransparent` flag. `Transparent` is a sentinel
  value, not "unset zero value" — see style inheritance below.
- `Cell` — one terminal cell: rune + fg/bg color. A plain value type,
  deliberately with no `NewCell` constructor — see its own doc comment
  for the bug a mismatched-argument-order constructor used to cause
  here (every cell's Fg/Bg silently swapped, everywhere, for every
  widget).
- `Buffer` — a `W*H` grid of `Cell` (flat, row-major storage — not
  `[][]*Cell`, so `Clear`/`Set` write values in place with zero heap
  allocation per cell), plus a clip-rect stack (`PushClip`/`PopClip`)
  that `Set` respects. What everything ultimately draws into. `Cells()`
  exposes the live flat slice (read-only by convention) for
  `render.Renderer.Flush`'s frame-diffing.
- `AppLog` / `Severity` — a log record and its level (Debug/Info/Warning/Fatal).
  `Severity`'s zero value is `Warning`, not `Debug` — see its own doc
  comment; `Severity` does double duty as both a record's own level and
  `AppConfig.LogLevel`'s filter threshold, so an unset `LogLevel`
  needed a safe default, not the loudest one.
- `AppSignal` — control signals passed between App and FaultManager (NOOP, SIGTERM).

### geom — geometry primitives

- `Point`, `Vector`, `Bounds` (`Pos *Point` + W/H), `Axis`.
- `Bounds` has a single `Valid()` helper for rejecting negative
  position/width/height. It used to be three separate helpers
  (`Validate`, `ValidateIfInsideBounds`, `ValidateNoPanic`) -- see
  `bounds.go`'s own comment for why they were consolidated: `Validate()`
  panicked on a plausible runtime value where every other fallible
  constructor in this codebase returns an error instead, and
  `ValidateIfInsideBounds`'s containment check duplicated
  `mixin.Node.IsInBounds`, which already does the same check. Most
  constructors currently do their own bounds checks rather than calling
  `Valid()` directly.

### framework — the contracts everything else implements against

- `Drawable` — the base interface every shape satisfies: `Draw`,
  `IsInBounds`, `SetLayer`/`GetLayer`, `SetParentStyle`, `SetContext`.
  `SetContext` is the single hook that carries redraw/logs/registry/focus
  info in one call (see `AppContext` below) — there's no separate
  invalidator or log-channel setter anymore.
- `Focusable` (a `Drawable` that can `HandleInput`/`Focus`/`Blur`/`IsFocused`),
  `FocusContainer` (a `Focusable` that also exposes `FocusableChildren()`
  for drill-in navigation), `ChildrenLister` (anything that can hand back
  its children for tree-walking, e.g. focus collection), `Identifiable`
  (has an `ID()`, so it can be registered/looked up), `Raisable` (accepts
  a raise-to-front callback from whatever's tracking it).
- `Composable` — `AddChild`/`RemoveChild`.
- `Style` + `ResolveStyle` — style inheritance. `core.Transparent` is the
  ONLY "inherit from parent" sentinel; any other color (including Black)
  is taken literally and does not inherit.
- `Anchor`/`AxisAnchor`/`ResolveAxis` — anchor resolution (Start/Center/End/NoAnchor).
- `LayoutPolicy` — `FreeLayout` (trust each child's declared position,
  resolve only its anchor) and `StackLayout` (stack children top-to-bottom,
  ignore declared Y, still resolve anchor for X). `CapacityAwareLayout` is
  an optional extension (`StackLayout` implements it) that lets a policy
  say whether a candidate child would overflow the frame, independent of
  the plain out-of-bounds check every child gets.
- `Event`/`Key` — input event vocabulary: `KeyRune`, the arrow keys,
  `KeyEnter`/`KeyEsc`/`KeyTab`/`KeyCtrlC`, and `KeyBackspace`/
  `KeyDelete`/`KeyHome`/`KeyEnd`. `IsStructuralKey` marks
  Ctrl+C/Enter/Tab/Esc, which `App.Run` always dispatches global-first
  (see `app`, below) — this is also what `Propagator`'s shadow-key
  warning checks against; Backspace/Delete/Home/End are ordinary,
  widget-first keys, not structural ones (see `input.Manager` below for
  the real terminal wire sequences each decodes from).
- `AppContext` — the bundle a node receives via `SetContext` once it's
  attached to the tree: `Redraw()` (the invalidate hook), `Log()`, the
  node `Registry`, a focus `Navigator`, `Lifecycle()` (closes on app
  stop), and `GlobalKeyBound(k)` for the shadow-key check.
- `Registry` — a concurrent id → `Drawable` map. `Register` reports if it
  just silently shadowed a previous node under the same id, so callers
  can log it. `FindAs[T]` is the generic lookup.
- `Logger` — thin wrapper around `chan<- core.AppLog`; nil-safe no-op by
  default, so a widget built before being attached to a `Canvas` never
  blocks or panics.
- `SpinnerContext` — the animation-cycle state (frames + index) behind
  `primitive.Spinner`; a handful of `New*SpinnerContext` constructors ship
  built-in cycles.

### mixin — reusable behavior every widget is built from

Three composition shapes recur throughout this framework: **leaf**
(`Node`, optionally `+FocusBehavior`), **is-a-container**
(`canvas.Container` and anything built directly on it — `Panel`,
`Bordered`, `TileGrid`, `List`), and **focusable-wrapper-around-one-
inner-container** (`Window`, `FocusableBox`, `ListRow`). This package
holds the mixins behind the first and third shapes; `canvas.Container`
itself is the second.

- `Node` (was `BaseNode`) — the thing almost every widget embeds: owns
  bounds, anchor, computed position, style + parent style, layer, and
  its `AppContext` (installed via `SetContext`, carrying the redraw
  hook, log channel, registry, and focus/global-key info). Implements
  `Layout`, `Style`/`ResolvedStyle`, `IsInBounds`, `Invalidate` (calls
  `ctx.Redraw()`), `Fault`/`Warn` (log helpers; `Fault` panics if
  there's no log channel to report through instead).
- `FocusableNode` (was `FocusableBaseNode`) — embeds `Node`, adds
  `focused bool`, a bound key-action map (`BindAction`/`HandleInput`),
  `BoundKeys()` (lists every key this node has an action for — used by
  `Propagator`'s shadow-key check), and `Focus`/`Blur`. Overrides
  `Style()` to blend in `focusStyle` when focused. This is composition
  shape 1's focusable variant — `primitive.Button` is the example.
  **Gotcha:** `Node.ResolvedStyle()` calls `Style()` on its own static
  `*Node` receiver — Go embedding isn't virtual dispatch, so this does
  NOT pick up `FocusableNode`'s focus-tinted override. Anything that
  needs the focus-tinted style has to call `focusableNode.Style()`
  explicitly, not `.ResolvedStyle()`.
- `FocusableWrapper` — composition shape 3: an outer `Node`+
  `FocusBehavior` identity wrapping exactly one inner `ContainerLike`
  (anything satisfying `Drawable`+`Composable`+`ChildrenLister`+
  `Resize` — `*canvas.Container` and `*widgets.Bordered` both already
  qualify with no changes needed). Gives `Style`, `Draw`,
  `AddChild`/`RemoveChild`/`Children`, `SetParentStyle`, `SetContext`,
  `Resize`, and `FocusableChildren` all for free. `Window`,
  `FocusableBox`, and `ListRow` are all built on this now — before this
  type existed, each hand-rolled the same combination independently
  (identical `Style()` overrides, near-identical `SetContext`, and
  `AddChild`/`RemoveChild`/`Children`/`SetParentStyle`/`Resize` all
  forwarding to one field under three different names). A composite
  that needs more than the generic case (`Window`'s title text + a
  clip) overrides just that one method on top, using `Inner()` to reach
  the wrapped `ContainerLike` — see `widgets.Window.Draw`.
- `Propagator` — fans `SetParentStyle`/`SetContext` out to a dynamic set
  of owned `Drawable`s (`Track`/`Untrack`), and does a few things beyond
  plain fan-out:
  - registers/unregisters `Identifiable` children in the `Registry`
    (`Track`/`Untrack` and `PropagateContext`), warning if an id just
    silently shadowed a previous node;
  - warns if a newly-attached child binds a structural key
    (Ctrl+C/Enter/Tab/Esc) that's _also_ bound globally — that widget
    binding will never fire, since structural keys always dispatch
    global-first (checked once, at the moment context first reaches the
    child; a global binding added later isn't retroactively checked);
  - wires `Raisable` children (anything with `SetRaiser`) to a closure
    that calls this `Propagator`'s own `BringToFront`, and implements
    `BringToFront`/`SendToBack` for simple z-order changes scoped to its
    own children (same idea as CSS z-index only ever comparing within
    its own stacking context);
  - calls `Stop()` automatically on any removed child implementing
    `framework.Stoppable` (`primitive.Spinner` does) — so `RemoveChild`
    alone is enough to stop a widget's own background goroutine, no
    separate cleanup call needed at every removal site.
    Currently used in exactly one place: `canvas.Container`. Composite
    widgets don't need their own copy of this — see canvas/widgets below.

### primitive — leaf drawables (no children)

Anything that never holds children (doesn't implement
`framework.Composable`/`ChildrenLister`) lives here, regardless of how
much behavior it has beyond rendering — a self-animating `Spinner` and
a focusable `Button` are exactly as much "primitive" in this sense as a
plain `Rect`, because the axis that matters for package placement is
composition shape (leaf vs. holds-children), not visual complexity.

- `Rect` — a filled rectangle of a single character.
- `Border` — draws a box outline (`BorderStyle`: corners + edges as
  runes) at a given thickness. Just `mixin.Node`, nothing owned.
- `Text` — draws a string at its computed position, always with a
  transparent background so it blends with whatever's behind it.
  `SetValue` resizes and invalidates — this is the hook background
  goroutines use to push live updates (e.g. a clock). Single-line only,
  deliberately — no wrapping/justification.
- `Button` — a focusable widget with its own rendering (fills its
  bounds, then draws a centered, width-clamped label).
  `ButtonConfig.OnActivate`, if set, is bound to Space — not Enter,
  since Enter is a structural key dispatched global-first (see `app`
  below) and would risk the exact shadow-key conflict
  `mixin.Propagator`'s own warning exists to catch.
- `Spinner` — ticks its own `SpinnerContext` on a background goroutine
  (started once, on first `SetContext`) and calls `Invalidate()` each
  frame — the same self-refreshing pattern `Text.SetValue` uses, just
  driven by a timer instead of external input. `Stop()` ends that
  goroutine independently of the whole app; also called automatically
  by `mixin.Propagator.Untrack` the moment a Spinner is removed from
  its container (see `framework.Stoppable`).
- `PaletteNode` — a grid of independently-colored cells (a color matrix
  instead of a single `Style`); bounds are derived from the matrix so
  they can never disagree with it. Backs `widgets.TileGrid`.

### canvas — the single composition primitive

- `Container` — embeds `BaseNode` + `Propagator`. Holds children (backed
  by `Propagator`, not a separate slice), sorts them by layer before each
  `Draw`, delegates arrangement to a `LayoutPolicy`. `AddChild` no longer
  does any bounds checking itself — it always `Track`s the child; the
  out-of-bounds check (`IsInBounds`) and, if the policy is a
  `CapacityAwareLayout`, the fit check both happen in `Draw`, once per
  child per frame, with a `warned` set so a standing condition (a window
  that's just always too small) logs once rather than once per frame.
- `Canvas` — the root of the whole tree. Wraps one root `Container` sized
  to the terminal (or a fixed size). `AddShape`/`Shapes`/`SetContext`/
  `SetParentStyle`/`BringToFront`/`SendToBack` all just delegate to the
  root container. `Compose()` clears the buffer and draws the tree.
  `CollectFocusable` walks the tree (via `ChildrenLister`) gathering
  `Focusable`s for `input.FocusManager`.

### widgets — composites (things that hold children)

- `Panel` — a styled, filled rectangle you can add children to: an outer
  `Container` holding a `primitive.Rect` fill and a content `Container`
  as siblings, so the fill never shows up in `Panel.Children()`. The
  primitive `Bordered`/`TileGrid` build on for "area with a real
  background plus children."
- `Bordered` (built via `NewBox`/`BoxConfig`) — a `primitive.Border`
  around a `Panel`, inset by border thickness + padding. Embeds
  `*canvas.Container` directly and adds border+panel to it via ordinary
  `AddChild` — so style/context propagation is inherited for free; only
  `AddChild`/`RemoveChild`/`Children` are overridden, to redirect from
  the wrapper to the inner panel's content. This is `mixin.
  FocusableWrapper`'s `ContainerLike` for `Window`/`FocusableBox` below.
- `Window` (`NewWindow`/`WindowConfig`) — built on `mixin.
  FocusableWrapper` wrapping a `Bordered`, plus an optional title
  `primitive.Text` and `Raisable` (`RaiseToFront`, wired automatically
  the instant a `Propagator` tracks it — see `Focus()` calling
  `RaiseToFront()` to bring itself forward on focus). `Style`, `Resize`,
  `AddChild`/`RemoveChild`/`Children`, `SetParentStyle`, and
  `FocusableChildren` all come from the wrapper; `Window` only overrides
  `Draw` (title + a clip, which the wrapper's generic case doesn't do)
  and `SetContext` (also needs to reach `title`).
- `FocusableBox` (`NewFocusableBox`) — `Bordered` + focus behavior, also
  built on `mixin.FocusableWrapper`. Needs no overrides beyond its own
  constructor — the generic wrapper is exactly what it is.
- `List` (`NewList`/`ListConfig`) — thin wrapper: `struct { *canvas.Container }`
  configured with `StackLayout`. `AddItem` builds a `ListRow` and adds it
  via the normal path. Needs no propagation code of its own — it's not a
  second copy of the pattern, it rides on `Container`.
- `ListRow` — a focusable, bordered-free row `List.AddItem` builds, also
  built on `mixin.FocusableWrapper` (wrapping a plain `*canvas.Container`,
  no border/fill). Like `FocusableBox`, needs no overrides beyond its
  constructor.
- `TileGrid` — a grid of independently-colored, one-character cells
  (color picker, heatmap, minimap). Structurally `Panel` with its
  `primitive.Rect` fill swapped for a `primitive.PaletteNode`; `SetCell`
  recolors one tile at runtime.

### input — turning bytes into events, and tracking focus

- `Manager` — puts the terminal in raw mode, decodes stdin byte-by-byte
  (including multi-byte ANSI escape sequences for arrow keys, and for
  Home/End/Delete — the latter two accept every terminal convention
  actually in use: xterm's unmodified `ESC[H`/`ESC[F`, vt220's
  `ESC[1~`/`ESC[4~`, and rxvt's `ESC[7~`/`ESC[8~`, plus the modified
  form of all three) into `framework.Event`s on a buffered channel (size configurable via
  `AppConfig.InputBufferSize`, default `input.DefaultEventBufferSize`).
  Drops events if the consumer is slow rather than blocking the read
  loop, logging the drop at Debug. Reads through an unexported
  `byteSource` interface, not `term.ReadStdin` directly — real usage
  wraps `term.ReadStdin`; tests substitute a scripted byte sequence so
  the whole escape-sequence decoder can be exercised without a real tty
  (see `manager_test.go`).
- `FocusManager`/`FocusScope` — a stack of focus scopes. `Next`/`Prev`
  cycle within the current scope. `Enter` pushes a new scope when the
  focused widget is a `FocusContainer` (drilling in without blurring the
  parent, so both light up). `Exit` pops back out.

### render — the frame loop

- `Renderer` — owns the terminal output buffer, runs either on a fixed-FPS
  ticker or purely on-demand (redraw only when `RequestRedraw` fires,
  itself only reachable if some node's `Invalidate()` reaches the
  `AppContext.Redraw` hook it was given). Handles terminal resize via
  `term.WatchResize`. `Flush` diffs each frame's cells against the
  previous one and only writes what actually changed, batching adjacent
  changed cells that share a style into a single escape + run of
  characters, rather than re-emitting a full cursor-move + color escape
  for every cell every frame.
- `RenderMode` (`FixedFPSMode(fps)` / `OnDemandMode()`) has unexported
  fields on purpose -- it's only constructible through those two
  functions, not a `RenderMode{}` literal. This closes a real trap: a
  hand-built or left-unset `RenderMode` used to compile fine but panic
  or hang at runtime (`FixedFPS` is `RenderMode`'s zero-value `Mode`,
  so `AppConfig{}`'s unset `RenderMode` produced `{Mode: FixedFPS, Fps:
  0}` -- a divide-by-zero the moment `Renderer.Run` built its ticker;
  separately, a hand-built `RenderMode{Mode: OnDemand}` skipped
  `OnDemandMode`'s channel allocation, leaving `Renderer.Run` blocked
  forever on a nil channel with no error and no panic, just silence).
  `render.NewRenderer` additionally validates at construction and
  returns an error rather than trusting whatever it's handed, since the
  all-zero `RenderMode{}` is still legal Go from any package regardless
  of the unexported fields.

### term — raw terminal I/O

- `raw.go` — enables/restores raw mode (`golang.org/x/sys/unix` ioctls),
  with signal-safe restoration on SIGTERM/SIGHUP and panic-safe restoration
  via a returned `restore func()`.
- `term.go` — `TermSize`, `WatchResize` (debounced SIGWINCH watcher).
- `ansi.go` — `CellToANSI`, converting a `core.Cell` to an escape sequence.

### fault — logging and crash handling

- `FaultManager` — owns the log channel, runs its own goroutine that
  writes to a timestamped log file, buffering failed writes in a
  `datastructs.RingBuffer` and retrying on a 1s ticker. A `Fatal`-severity
  log is written immediately (skipping the buffer) and sends `SIGTERM` up
  to the App, but keeps draining afterward so `App.Stop()`'s own shutdown
  logs still land -- including logs still sitting in the channel's
  buffer the instant `Stop()` is called, which are drained before
  returning rather than possibly lost to a race between "new log
  arrived" and "shutting down." `NewFaultManager` takes a `fault.Config`
  (`LogLevel`, `LogDir`, `DisableFileLog`) and does its fallible setup --
  creating the log directory, opening this run's file -- synchronously,
  right there, returning a real error on failure rather than deferring
  that into the goroutine `Start()` launches later. `Config{}`'s zero
  value is a safe default: `LogLevel` reads as `core.Warning` (see
  `core.Severity`'s own doc comment for why that's the zero value, not
  `Debug`), `LogDir` empty falls back to `"logs"`, and
  `DisableFileLog` false keeps writing to disk as before.

### datastructs

- `RingBuffer` — fixed-capacity string ring buffer, oldest-entry-drops-on-overflow.
  Used only by `fault.FaultManager` for its write-retry buffer. Capacity
  must be >= 2 (`NewRingBuffer` panics below that) -- a single-index
  read/write scheme can't tell full apart from empty otherwise, and as a
  consequence any capacity `N` only ever holds `N-1` items usable at
  once. `fault.NewFaultManager`'s own `NewRingBuffer(100)` really gives
  99 usable retry slots, not 100 -- expected, not a bug.

### app — top-level wiring

- `App` — owns `Canvas`, `Renderer`, `Input` manager, `Focus` manager, and
  the `FaultManager`. `NewApp` seeds `appEvents` with Ctrl+C bound to
  `QuitAction()` unconditionally (raw mode disables ISIG, so the
  terminal's own SIGINT never fires — without this a keyboard-only user
  has no way to exit), then layers `cfg.AppEvents` on top, so a caller
  can still redefine Ctrl+C but doesn't have to think about it if not.
  `BindKey`/`UnbindKey`/`BindMouse`/`UnbindMouse` change global bindings
  after construction, and are safe to call from any goroutine —
  `bindingsMu` (a `sync.RWMutex`, same idea as `Propagator`'s own guard
  on `owned`) protects `appEvents`/`mouseHandler` from racing against
  `Run`'s dispatch loop or `Propagator`'s shadow-key check, both of
  which read them from whatever goroutine happens to be running.
  `AppConfig.Fg`/`Bg` follow `framework.Style`'s own convention (plain
  `core.Color`, `Transparent` means inherit/let `NewCanvas` substitute
  its defaults) rather than the `*core.Color` + nil-check dance they
  used to require. `AppConfig.LogLevel`/`LogDir`/`DisableFileLog`
  configure the `FaultManager`; `AppConfig.InputBufferSize` overrides
  the input event channel's capacity.

  `Run()`'s per-event dispatch logic lives in its own method,
  `handleEvent` — extracted specifically so it can be exercised
  directly in tests with a synthetic focused widget and synthetic
  bindings, without a real terminal (see `app_test.go`). It reads
  `AppSignal`s (stopping on `SIGTERM`) and input `Event`s, then
  dispatches each key with a deliberate asymmetry:
  - **structural keys** (Ctrl+C/Enter/Tab/Esc) go **global-first**: if
    `appEvents` has a binding, that runs and the focused widget never
    sees the key; otherwise it falls through to the focused widget.
  - **every other key** goes **widget-first**: the focused widget's
    `HandleInput` gets first refusal, and only if it doesn't claim the
    key (`handled == false`) does a global binding get a shot.
    This asymmetry is exactly what `Propagator`'s shadow-key warning is
    protecting against: a widget bound to a structural key that's _also_
    bound globally would silently never fire.

  Logging throughout `App`/`Renderer`/`Input` goes through the same
  `framework.Logger` widgets use (via `mixin.Node.Logger()`) — not raw
  `chan<- core.AppLog` sends. `Debug`/`Info` are non-blocking and
  severity-filtered at the source (so a stalled log writer can't stall
  input dispatch, and a below-threshold message costs a channel-op
  check, not an allocation); `Warning`/`Fatal` still block, deliberately
  — they're rare, and `Fatal` specifically is what promotes to a
  `SIGTERM` shutdown, so silently dropping one under backpressure would
  disable that escape hatch exactly when something's already gone
  wrong. The per-keystroke/per-mouse-event dispatch logs are `Debug`
  severity and additionally guarded by `Logger.Enabled(core.Debug)`
  before the `fmt.Sprintf` that builds them runs at all, since Go
  evaluates a function's arguments before the call regardless of what
  the callee does with them.

### examples

Each subfolder is a standalone `package main` demonstrating one thing:

- `clip-demo` — a child wider than its `Window`, to show the
  out-of-bounds warning/clipping in `Container.Draw`, next to one that
  fits.
- `clock-window` — a `Window` with a `Text` label updated live from a
  background goroutine.
- `focus-drill-test` — nested `FocusableBox`es, exercising
  `FocusManager.Enter`/`Exit`.
- `list-example` — a `List` with several rows.
- `mouse-paint-demo` — a `TileGrid` painted via `BindMouse`, exercising
  mouse press/drag/wheel decoding end to end.
- `spinner-example` — a `Spinner`, exercising the self-refreshing-from-
  a-goroutine pattern.
- `tab-nav-example` — two `Window`s, exercising `app.NavActions()`
  (Tab/Shift+Tab) switching focus between them.
- `tile-grid-demo` — a `TileGrid` checkerboard, exercising the
  `PaletteNode` fill primitive.

## How a frame actually happens

1. `App.Run()` blocks on `appSignals` and `Input.Events()`.
2. An input event is dispatched per the structural-vs-other-key rule
   above: global-first with widget fallback for Ctrl+C/Enter/Tab/Esc,
   widget-first with global fallback for everything else.
3. A handler mutates widget state and (usually) calls `Invalidate()` on
   itself, or returns `redraw=true` from a global action.
4. `Invalidate()` is `n.ctx.Redraw()` — every node holds its own copy of
   `AppContext`, fanned out by `Propagator.Track`/`PropagateContext` the
   moment the node is attached, so this calls straight through to
   whatever `Renderer.RequestRedraw` was wired in as `AppContext.Invalidate`
   at app startup. There's no walking up a chain of ancestors — every
   node's context already points directly at the renderer.
5. `Renderer.render` calls `Canvas.Compose()` (clear buffer, `root.Draw`),
   then `Flush`s the buffer to the terminal as ANSI. (In `OnDemand` mode
   this only happens when step 4 actually fires `RequestRedraw`; in
   `FixedFPS` mode the renderer redraws every tick regardless of step 4.)

## Style/context propagation, in one sentence

A parent hands a child its resolved style and its `AppContext` (redraw
hook, logs, node registry, global-key lookup) either at `AddChild` time
(if the parent is already wired) or the next time the parent itself
receives one of those calls from further up — `Propagator` (inside
`Container`) and the composites' promoted `*canvas.Container` methods
both guarantee this regardless of add-order, and `Propagator` piggybacks
registry registration and the shadow-key warning onto the same moment.
