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
	fill    *base.PalleteNode
	content *canvas.Container
}

type TileGridConfig struct {
	ColorMatrix [][]core.Color
	Layer       int
	Anchor      framework.Anchor
}

func NewTileGrid(pos *geom.Point, cfg TileGridConfig) (*TileGrid, error) {
	fill, err := base.NewPalleteNode(&geom.Point{}, framework.Anchor{}, cfg.ColorMatrix, 0, "TileGrid")
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
