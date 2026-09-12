# Glyph

![glyph logo](go_glyph.png)

A terminal UI framework for Go.

Glyph provides a small set of composable primitives — containers, borders,
text, lists, and focusable widgets — for building interactive terminal
applications without wiring up rendering, layout, and input handling by
hand.

## Features

- **Composable widgets.** A single `Container` type parameterized by a
  `LayoutPolicy` replaces the need for separate container types — free
  positioning and stacked layouts are both just configuration.
- **Style inheritance.** Colors and styles cascade from parent to child
  automatically, with `Transparent` as an explicit "inherit" value.
- **Focus management.** Tab/Enter/Esc navigation, including drilling into
  nested focusable containers and back out, plus a warning if a widget
  binds a structural key that's already claimed globally (it would never
  fire).
- **Self-refreshing components.** Any widget can update its own state from
  a background goroutine and trigger a redraw — `Spinner` and the
  live-updating `Text` clock example both work this way.
- **Two render modes.** Fixed frame rate, or on-demand redraw only when
  something actually changes.

## Installation

```bash
go get github.com/dmsRosa6/glyph
```

## Quick start

```go
package main

import (
	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/render"
)

func main() {
	rm, err := render.FixedFPSMode(30)
	if err != nil {
		panic(err)
	}

	a, err := app.NewApp(app.AppConfig{
		Bg:         core.Black,
		RenderMode: rm,
	})
	if err != nil {
		panic(err)
	}

	// build your UI, then:
	a.Run()
}
```

See the `examples/` directory for fuller examples: bordered boxes and
out-of-bounds clipping, a drillable focusable widget tree, stacked lists,
a tile grid, a spinner, mouse painting, tab navigation between windows,
and a self-updating clock window.

## Widgets

Split across two packages by composition shape, not by "how simple it
looks": `primitive` holds leaves (nothing implements `Composable`/
`ChildrenLister`), `widgets` holds composites (things that hold
children). A self-animating `Spinner` and a focusable `Button` are just
as much a primitive as a plain `Rect` in this sense — see `devguide.md`
for the full rationale, including the third composition shape
(`mixin.FocusableWrapper`) that `Window`/`FocusableBox`/`ListRow` are
all built on.

### `primitive` — leaves

- **Rect** — a filled rectangle of a single character, with optional
  clipping.
- **Text** — a single line of styled text. Can be updated at runtime,
  including from a background goroutine, and will ask for a redraw when
  it changes.
- **Border** — draws a frame (corners, edges) around a bounds. Comes with
  a few built-in styles (single line, double line, rounded) and supports
  custom ones.
- **Button** — a focusable widget with a bound action and its own
  rendering (a filled, centered label).
- **Spinner** — an animated loading indicator; ticks itself on a
  background goroutine and requests a redraw each frame. Several
  built-in cycles (slash, dots, pulse, braille, blocks, clock, and more).
  `Stop()` ends its own ticking goroutine independently of the whole
  app; it's also called automatically the moment a Spinner is removed
  from its container (`RemoveChild`), so a removed Spinner doesn't keep
  ticking (and requesting redraws) for the rest of the process's life.
- **PaletteNode** — a grid of independently-colored cells; backs
  `widgets.TileGrid` below.

### `widgets` — composites

- **Panel** — a styled, filled rectangle you can add children to. What
  `Bordered` puts inside its border; also usable directly for content
  areas that don't need a frame.
- **Bordered** (built via `NewBox`/`BoxConfig`) — wraps a `Panel` with a
  `primitive.Border`, inset by border thickness + padding. This is the
  general "frame around something" composite.
- **FocusableBox** — a bordered, padded, focusable container. Supports a
  distinct style while focused, and can hold further focusable children
  that `FocusManager.Enter()` can drill into.
- **List** — a container with a stacked layout, plus a convenience
  method (`AddItem`) for adding bordered, padded rows (`ListRow`).
- **Window** — a `Bordered` box with a title overlaid on the border
  itself, plus bring-to-front-on-focus behavior.
- **TileGrid** — a grid of independently-colored, single-character cells
  (color swatches, heatmaps, minimaps), with children still supported on
  top.

## Render modes

Passed as `RenderMode` in `app.AppConfig`, built via `FixedFPSMode(fps)`
or `OnDemandMode()` -- not a `RenderMode{}` literal. `RenderMode`'s
fields are unexported specifically to rule that out: a hand-built or
left-unset `RenderMode` used to compile fine but panic (`FixedFPS` with
`Fps: 0`, a divide-by-zero) or hang forever with no error at all
(`OnDemand` missing its channel allocation) the moment the renderer
actually started. `app.NewApp`/`render.NewRenderer` now return an error
for an invalid `RenderMode` instead of trusting it.

- **FixedFPS** — redraws on a fixed timer regardless of whether anything
  changed. Simple and predictable, at the cost of drawing frames that
  don't need it.
- **OnDemand** — redraws only when something calls `Invalidate()`
  (directly, or indirectly via a widget's own `Invalidate()`/`Warn` calls)
  and that reaches `Renderer.RequestRedraw`. Typically triggered by an
  input handler returning `redraw=true`, or a widget mutating its own
  state from a background goroutine (see `Spinner`, or `Text.SetValue`).

## The `framework` package

`framework` holds the shared contracts every other package depends on,
with no rendering or terminal logic of its own.

- **interfaces.go** — the core interfaces: `Drawable` (can be drawn and
  positioned), `Focusable` (can take input and focus), `Composable` (can
  hold children), `Layoutable`, `Clippable`, and `ChildrenLister` (lets a
  container expose its children without callers needing to know its
  concrete type).
- **style.go** — `Style` and `ResolveStyle`, the parent/child style
  inheritance logic. `Transparent` is the sentinel that means "inherit."
- **layout.go** — `Anchor` and axis resolution for positioning a widget
  within its parent (start, center, end, or an explicit position).
- **layoutpolicy.go** — `LayoutPolicy` and its two implementations,
  `FreeLayout` (children keep their own declared position) and
  `StackLayout` (children stack top to bottom). `CapacityAwareLayout` is
  an optional extension a policy can implement so `Container` can also
  warn when a child won't fit, not just when it's out of bounds.
- **event.go** — `Event` and `Key`, the input event vocabulary, plus
  `IsStructuralKey` (Ctrl+C/Enter/Tab/Esc), which get global-first
  dispatch — see `app`, below.
- **appcontext.go** — `AppContext`, the bundle every node gets once it's
  attached to the tree: redraw hook, log channel, node `Registry`, focus
  navigator, and a way to check whether a key is bound globally.
- **registry.go** — `Registry`, a concurrent id → `Drawable` map so any
  widget with an id can be looked up from anywhere via
  `framework.FindAs[T]`.
- **logger.go** — `Logger`, a thin wrapper around `chan<- core.AppLog`;
  nil-safe no-op by default, so a widget built before being attached to
  a `Canvas` never blocks or panics.

See `devguide.md` to more information, I also use it as a guideline for things :)

## Still Adding stuff...probably
