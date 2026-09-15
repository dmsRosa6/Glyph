<h1 align="center">Glyph</h1>

<p align="center">A terminal UI framework for Go.</p>
<p align="center">
  <img src="go_glyph.png" alt="glyph logo" width="319">
</p>

## About

Glyph is a mini for fun project framework that provides composable primitives — containers, borders, text,
lists, focusable widgets — for building interactive terminal apps
without wiring up rendering, layout, and input handling by hand.

## Install

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

## Related projects

- **[Glyph-Examples](https://github.com/dmsRosa6/Glyph-Examples)** —
  standalone runnable examples: bordered boxes and out-of-bounds
  clipping, a drillable focusable tree, fixed and scrollable lists, a
  tile grid, a spinner, mouse painting, tab navigation, an editable
  text field, and a self-updating clock. Start here if you want to
  run something rather than read about it.
- **[Glyph-Explorer](https://github.com/dmsRosa6/Glyph-Explorer)** —
  an interactive explorer built on top of glyph.

## Docs

**[devguide.md](devguide.md) is the actual documentation** — what
every package does, how a frame gets from widget tree to terminal, and
the handful of gotchas worth knowing before you build something custom
on top of the mixins. This README is just a landing page.

## What's here

- **Widgets.** `primitive` (leaves: `Rect`, `Border`, `StaticText`,
  `Text`, `MultilineText`, `Button`, `Spinner`, `TextInput`,
  `PaletteNode`) and `widgets` (composites: `Panel`, `Bordered`,
  `Window`, `FocusableBox`, `List`, `TileGrid`).
- **Style inheritance**, with `Transparent` as an explicit "inherit"
  value.
- **Focus management** — Tab/Enter/Esc navigation, drilling into
  nested containers, and a warning if a widget's key binding would be
  silently shadowed by a global one.
- **Two render modes** — fixed FPS, or on-demand redraw only when
  something changed.
- **Scrollable lists** — `List` is fixed by default; `Scrollable`
  auto-scrolls to keep the focused row visible, with `ScrollBy`/
  `ScrollTo` as a manual escape hatch.
- **Panic recovery** — a panic in a `Draw` method or a bound handler
  is caught and turned into a clean shutdown instead of a wrecked
  terminal.

Full detail on all of it, including the reasoning behind the less
obvious choices, is in [devguide.md](devguide.md).

## Roadmap

Status: heading toward a v0.0.1 tag. Core is stable enough to build
real things on.

**Deliberately not built, on purpose:**

- **Percentage-based sizing** ("50% of available width" instead of a
  fixed cell count). `Anchor` already resolves _position_ against the
  parent every frame; _size_ has no equivalent yet — every widget's
  W/H is a fixed int at construction. Doing this properly needs a
  `Sizing` concept threaded through every widget config so a
  percentage stays correct across terminal resizes — real surface
  area across ~15 constructors, worth doing on purpose rather than
  half-built. (Cheap fallback if ever wanted without the full
  commitment: a one-shot `geom.Percent(w, h, parent)` helper that
  resolves once at construction, no live tracking.)
- **Mouse hit-testing** — no per-widget mouse dispatch; `BindMouse`
  gives you raw coordinates, you do your own hit-testing.
- **A general scroll container** for arbitrary content — `List`'s
  `Scrollable` mode covers stacked rows; a `ScrollContainer` for any
  child is a different, unbuilt thing.
- **A real multi-line `TextInput`** — single-line only for now.
- **Perfomace** - Add the real notion of dirty components and "partial hydration" and review render package as a whole
- **More widgets**

If you're picking this back up later: start with `devguide.md` for
the architecture, then `todo.md` for the itemized backlog.
