package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

type ListRow struct {
	mixin.FocusableWrapper
}

func newListRow(bounds *geom.Bounds, anchor framework.Anchor) (*ListRow, error) {
	bn, err := mixin.NewNode(bounds, anchor, framework.Style{Bg: core.Transparent, Fg: core.Transparent}, 0, "ListRow")
	if err != nil {
		return nil, err
	}

	content, err := canvas.NewContainer(geom.NewBounds(0, 0, bounds.W, bounds.H), canvas.ContainerConfig{
		Style: framework.Style{Bg: core.Transparent, Fg: core.Transparent},
	})
	if err != nil {
		return nil, err
	}

	return &ListRow{
		FocusableWrapper: mixin.NewFocusableWrapper(bn, content),
	}, nil
}
