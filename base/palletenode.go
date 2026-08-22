package base

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// PalleteNode is a grid of independently-colored cells -- BaseNode plus
// a color matrix instead of a single Style. Bounds are derived from the
// matrix (W = row length, H = row count), not passed in separately, so
// the two can never disagree.
type PalleteNode struct {
	BaseNode
	colorMatrix [][]core.Color
}

func NewPalleteNode(pos *geom.Point, anchor framework.Anchor, colorMatrix [][]core.Color, layer int, source string) (*PalleteNode, error) {
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

	bn, err := NewBaseNode(bounds, anchor, *framework.NewTransparentStyle(), layer, source)
	if err != nil {
		return nil, err
	}

	return &PalleteNode{BaseNode: bn, colorMatrix: colorMatrix}, nil
}

func (n *PalleteNode) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := n.ComputedPos()
	for y, row := range n.colorMatrix {
		for x, c := range row {
			buf.Set(vec.X+pos.X+x, vec.Y+pos.Y+y, ' ', c, c)
		}
	}
}

// SetCell updates a single cell's color and requests a redraw. Returns
// an error rather than panicking so callers driving this from user
// input (e.g. a click handler) can no-op on an out-of-range coordinate
// instead of crashing the app.
func (n *PalleteNode) SetCell(x, y int, c core.Color) error {
	if y < 0 || y >= len(n.colorMatrix) || x < 0 || x >= len(n.colorMatrix[y]) {
		return errors.New("cell out of bounds")
	}
	n.colorMatrix[y][x] = c
	n.Invalidate()
	return nil
}
