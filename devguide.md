# glyph — Developer Guide

A terminal UI framework in Go. This doc explains what each package is
responsible for and how a frame actually gets from "widget tree" to
"pixels on screen."

This describes the code as it stands today, not where it's headed — see
`todo.md` for known gaps and planned refactors (concurrency on
`Propagator`, splitting focus behavior into a true mixin, etc.).

## Package layout and responsibilities

### core — primitive data, no framework knowledge

- `Color` — RGB + an `IsTransparent` flag. `Transparent` is a sentinel
  value, not "unset zero value" — see style inheritance below.
- `Cell` — one terminal cell: rune + fg/bg color.
- `Buffer` — a 2D grid of `*Cell`, plus a clip-rect stack (`PushClip`/
  `PopClip`) that `Set` respects. What everything ultimately draws into.
- `AppLog` / `Severity` — a log record and its level (Debug/Info/Warning/Fatal).
- `AppSignal` — control signals passed between App and FaultManager (NOOP, SIGTERM).

### geom — geometry primitives

- `Point`, `Vector`, `Bounds` (`Pos *Point` + W/H), `Axis`.
- `Bounds` has `Validate`/`ValidateIfInsideBounds`/`ValidateNoPanic` helpers
  for rejecting negative or out-of-parent geometry, though most
  constructors currently do their own bounds checks rather than calling
  these directly.

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
- `Event`/`Key` — input event vocabulary. `IsStructuralKey` marks
  Ctrl+C/Enter/Tab/Esc, which `App.Run` always dispatches global-first
  (see `app`, below) — this is also what `Propagator`'s shadow-key
  warning checks against.
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
  `widgets.Spinner`; a handful of `New*SpinnerContext` constructors ship
  built-in cycles.

### base — shared node behavior every widget embeds

- `BaseNode` — the thing almost every widget embeds: owns bounds, anchor,
  computed position, style + parent style, layer, and its `AppContext`
  (installed via `SetContext`, carrying the redraw hook, log channel,
  registry, and focus/global-key info). Implements `Layout`,
  `Style`/`ResolvedStyle`, `IsInBounds`, `Invalidate` (calls
  `ctx.Redraw()`), `Fault`/`Warn` (log helpers; `Fault` panics if there's
  no log channel to report through instead).
- `FocusableBaseNode` — embeds `BaseNode`, adds `focused bool`, a bound
  key-action map (`BindAction`/`HandleInput`), `BoundKeys()` (lists every
  key this node has an action for — used by `Propagator`'s shadow-key
  check), and `Focus`/`Blur`. Overrides `Style()` to blend in
  `focusStyle` when focused.
  **Gotcha:** `BaseNode.ResolvedStyle()` calls `Style()` on its own static
  `*BaseNode` receiver — Go embedding isn't virtual dispatch, so this does
  NOT pick up `FocusableBaseNode`'s focus-tinted override. Anything that
  needs the focus-tinted style has to call
  `focusableBaseNode.Style()` explicitly, not `.ResolvedStyle()`.
- `PalleteNode` — a grid of independently-colored cells (a color matrix
  instead of a single `Style`); bounds are derived from the matrix so
  they can never disagree with it. Backs `widgets.TileGrid`.
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
    its own stacking context).
    Currently used in exactly one place: `canvas.Container`. Composite
    widgets don't need their own copy of this — see canvas/widgets below.

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

### widgets — concrete shapes and composites

- `Rect` — leaf. A filled rectangle of a single character.
- `Border` — leaf. Draws a box outline (`BorderStyle`: corners + edges as
  runes) at a given thickness. Just `BaseNode`, nothing owned.
- `Text` — leaf. Draws a string at its computed position, always with a
  transparent background so it blends with whatever's behind it.
  `SetValue` resizes and invalidates — this is the hook background
  goroutines use to push live updates (e.g. a clock).
- `Panel` — a styled, filled rectangle you can add children to: an outer
  `Container` holding a `Rect` fill and a content `Container` as
  siblings, so the fill never shows up in `Panel.Children()`. The
  primitive `Bordered`/`TileGrid` build on for "area with a real
  background plus children."
- `Bordered` (built via `NewBox`/`BoxConfig`) — a `Border` around a
  `Panel`, inset by border thickness + padding. Embeds `*canvas.Container`
  directly and adds border+panel to it via ordinary `AddChild` — so
  style/context propagation is inherited for free; only
  `AddChild`/`RemoveChild`/`Children` are overridden, to redirect from
  the wrapper to the inner panel's content.
- `Window` (`NewWindow`/`WindowConfig`) — a `Bordered` + an optional title
  `Text`, plus focus behavior and `Raisable` (`RaiseToFront`, wired
  automatically the instant a `Propagator` tracks it — see `Focus()`
  calling `RaiseToFront()` to bring itself forward on focus).
- `FocusableBox` (`NewFocusableBox`) — `Bordered` + focus behavior.
  Can't embed both `FocusableBaseNode` and `*canvas.Container` (ambiguous
  promoted methods at the same depth — compile error, not a merge), so it
  holds its `Bordered` as a plain field and forwards
  `SetParentStyle`/`SetContext` to it directly. `Draw` re-pushes the
  focus-resolved style to its box every frame, since `Focus()`/`Blur()`
  only flip a bool and call `Invalidate()` — they don't re-propagate
  style themselves. (`Window`/`ListRow` follow the same shape and the
  same gotcha.)
- `List` (`NewList`/`ListConfig`) — thin wrapper: `struct { *canvas.Container }`
  configured with `StackLayout`. `AddItem` builds a `ListRow` and adds it
  via the normal path. Needs no propagation code of its own — it's not a
  second copy of the pattern, it rides on `Container`.
- `ListRow` — a focusable, bordered-free row `List.AddItem` builds; same
  field-forwarding shape as `FocusableBox`/`Window`.
- `TileGrid` — a grid of independently-colored, one-character cells
  (color picker, heatmap, minimap). Structurally `Panel` with its `Rect`
  fill swapped for a `base.PalleteNode`; `SetCell` recolors one tile at
  runtime.
- `Button` — a focusable widget with a bound action and its own
  rendering: fills its bounds, then draws a centered, width-clamped
  label.
- `Spinner` — ticks its own `SpinnerContext` on a background goroutine
  (started once, on first `SetContext`) and calls `Invalidate()` each
  frame — the same self-refreshing pattern `Text.SetValue` uses, just
  driven by a timer instead of external input.

### input — turning bytes into events, and tracking focus

- `Manager` — puts the terminal in raw mode, decodes stdin byte-by-byte
  (including multi-byte ANSI escape sequences for arrow keys) into
  `framework.Event`s on a buffered channel. Drops events if the consumer
  is slow rather than blocking the read loop.
- `FocusManager`/`FocusScope` — a stack of focus scopes. `Next`/`Prev`
  cycle within the current scope. `Enter` pushes a new scope when the
  focused widget is a `FocusContainer` (drilling in without blurring the
  parent, so both light up). `Exit` pops back out.

### render — the frame loop

- `Renderer` — owns the terminal output buffer, runs either on a fixed-FPS
  ticker or purely on-demand (redraw only when `RequestRedraw` fires,
  itself only reachable if some node's `Invalidate()` reaches the
  `AppContext.Redraw` hook it was given). Handles terminal resize via
  `term.WatchResize`. `Flush` walks the `core.Buffer` and writes ANSI
  escape codes per cell.

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
  logs still land.

### datastructs

- `RingBuffer` — fixed-capacity string ring buffer, oldest-entry-drops-on-overflow.
  Used only by `fault.FaultManager` for its write-retry buffer.

### app — top-level wiring

- `App` — owns `Canvas`, `Renderer`, `Input` manager, `Focus` manager, and
  the `FaultManager`. `NewApp` seeds `appEvents` with Ctrl+C bound to
  `QuitAction()` unconditionally (raw mode disables ISIG, so the
  terminal's own SIGINT never fires — without this a keyboard-only user
  has no way to exit), then layers `cfg.AppEvents` on top, so a caller
  can still redefine Ctrl+C but doesn't have to think about it if not.
  `BindKey`/`UnbindKey` change global bindings after construction.
  `Run()` is the main select loop: reads `AppSignal`s (stopping on
  `SIGTERM`) and input `Event`s, then dispatches each key with a
  deliberate asymmetry —
  - **structural keys** (Ctrl+C/Enter/Tab/Esc) go **global-first**: if
    `appEvents` has a binding, that runs and the focused widget never
    sees the key; otherwise it falls through to the focused widget.
  - **every other key** goes **widget-first**: the focused widget's
    `HandleInput` gets first refusal, and only if it doesn't claim the
    key (`handled == false`) does a global binding get a shot.
    This asymmetry is exactly what `Propagator`'s shadow-key warning is
    protecting against: a widget bound to a structural key that's _also_
    bound globally would silently never fire.

### examples

Each subfolder is a standalone `package main` demonstrating one thing:

- `clip-demo` — a child wider than its `Window`, to show the
  out-of-bounds warning/clipping in `Container.Draw`, next to one that
  fits.
- `clock-window` — a `Window` with a `Text` label updated live from a
  background goroutine.
- `fault-test` — a widget wired to `Warn`/`Fault`, exercising
  `FaultManager`'s log path including the retry-buffer/`Fatal` cases.
- `focus-drill-test` — nested `FocusableBox`es, exercising
  `FocusManager.Enter`/`Exit`.
- `list-example` — a `List` with several rows.
- `list-window-demo` — a `List` embedded inside a `Window`.
- `spinner-example` — a `Spinner`, exercising the self-refreshing-from-
  a-goroutine pattern.
- `status-bar` — a `Bordered` used as a simple status bar.
- `tile-grid-demo` — a `TileGrid` checkerboard, exercising the
  `PalleteNode` fill primitive.

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
