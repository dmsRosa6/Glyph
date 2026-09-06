package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// TileGrid is a grid of independently-colored, one-character cells --
// think a color-picker swatch, a heatmap, or a minimap. Structurally
// it's Panel with its Rect fill swapped for a base.PalleteNode: the
// outer Container holds [fill, content] as siblings so Children() (via
// content) stays pure user content, exactly like Panel/Bordered do.
type TileGrid struct {
	*canvas.Container
	fill    *base.PaletteNode
	content *canvas.Container
}

type TileGridConfig struct {
	ColorMatrix [][]core.Color
	Layer       int
	Anchor      framework.Anchor
}

func NewTileGrid(pos *geom.Point, cfg TileGridConfig) (*TileGrid, error) {
	fill, err := base.NewPaletteNode(&geom.Point{}, framework.Anchor{}, cfg.ColorMatrix, 0, "TileGrid")
	if err != nil {
		return nil, err
	}
	w, h := fill.Size()

	outer, err := canvas.NewContainer(geom.NewBounds(pos.X, pos.Y, w, h), canvas.ContainerConfig{
		Style:  framework.Style{Bg: core.Transparent, Fg: core.Transparent},
		Layer:  cfg.Layer,
		Anchor: cfg.Anchor,
	})
	if err != nil {
		return nil, err
	}

	content, err := canvas.NewContainer(geom.NewBounds(0, 0, w, h), canvas.ContainerConfig{
		Style: framework.Style{Bg: core.Transparent, Fg: core.Transparent},
	})
	if err != nil {
		return nil, err
	}

	outer.AddChild(fill)
	outer.AddChild(content)

	return &TileGrid{Container: outer, fill: fill, content: content}, nil
}

// Resize shadows the promoted *canvas.Container.Resize to at least
// cascade to content, matching Panel/Bordered. It deliberately does
// NOT resize fill: fill's bounds are derived from -- and by
// base.NewPaletteNode's own contract can never disagree with -- its
// color matrix (see PaletteNode's doc comment). Stretching fill's
// reported bounds without also reshaping the matrix would just make
// the two disagree, which is exactly what that contract exists to
// prevent. A TileGrid's true size IS the size of its palette; this
// only changes how much room its content children (drawn on top of the
// palette) have to work with.
func (t *TileGrid) Resize(w, h int) {
	t.Container.Resize(w, h)
	t.content.Resize(w, h)
}

// AddChild puts user content into the content container, not the outer
// wrapper -- shadows the promoted Container.AddChild, same reason
// Bordered/Panel shadow it: otherwise it'd land next to fill.
func (t *TileGrid) AddChild(child framework.Drawable) {
	t.content.AddChild(child)
}

func (t *TileGrid) RemoveChild(target framework.Drawable) {
	t.content.RemoveChild(target)
}

func (t *TileGrid) Children() []framework.Drawable {
	return t.content.Children()
}

// SetCell recolors one tile at runtime.
func (t *TileGrid) SetCell(x, y int, c core.Color) error {
	return t.fill.SetCell(x, y, c)
}
