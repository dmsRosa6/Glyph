package primitive

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

type PaletteNode struct {
	mixin.Node
	colorMatrix [][]core.Color
}

func NewPaletteNode(pos *geom.Point, anchor framework.Anchor, colorMatrix [][]core.Color, layer int, source string) (*PaletteNode, error) {
	h := len(colorMatrix)
	if h == 0 {
		return nil, errors.New("color matrix must have at least one row")
	}
	w := len(colorMatrix[0])
	if w == 0 {
		return nil, errors.New("color matrix rows must have at least one column")
	}
	for _, row := range colorMatrix {
		if len(row) != w {
			return nil, errors.New("color matrix rows must all be the same length")
		}
	}

	bounds := geom.NewBounds(pos.X, pos.Y, w, h)

	bn, err := mixin.NewNode(bounds, anchor, *framework.NewTransparentStyle(), layer, source)
	if err != nil {
		return nil, err
	}

	return &PaletteNode{Node: bn, colorMatrix: colorMatrix}, nil
}

func (n *PaletteNode) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := n.ComputedPos()
	for y, row := range n.colorMatrix {
		for x, c := range row {
			buf.Set(vec.X+pos.X+x, vec.Y+pos.Y+y, ' ', c, c)
		}
	}
}

func (n *PaletteNode) SetCell(x, y int, c core.Color) error {
	if y < 0 || y >= len(n.colorMatrix) || x < 0 || x >= len(n.colorMatrix[y]) {
		return errors.New("cell out of bounds")
	}
	n.colorMatrix[y][x] = c
	n.Invalidate()
	return nil
}
