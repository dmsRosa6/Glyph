package widgets

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
	"github.com/dmsRosa6/glyph/primitive"
)

type FocusableBox struct {
	mixin.FocusableWrapper
}

type FocusableBoxConfig struct {
	Padding      int
	BorderConfig primitive.BorderConfig
	Style        framework.Style
	FocusStyle   *framework.Style
	Layer        int
	Anchor       framework.Anchor
}

func NewFocusableBox(bounds *geom.Bounds, cfg FocusableBoxConfig) (*FocusableBox, error) {
	bn, err := mixin.NewNode(bounds, cfg.Anchor, framework.Style{Bg: core.Transparent, Fg: core.Transparent}, cfg.Layer, "FocusableBox")
	if err != nil {
		return nil, err
	}

	box, err := NewBox(geom.NewBounds(0, 0, bounds.W, bounds.H), BoxConfig{
		Padding:      cfg.Padding,
		Style:        cfg.Style,
		BorderConfig: cfg.BorderConfig,
	})
	if err != nil {
		return nil, err
	}

	fb := &FocusableBox{
		FocusableWrapper: mixin.NewFocusableWrapper(bn, box),
	}
	if cfg.FocusStyle != nil {
		fb.SetFocusStyle(*cfg.FocusStyle)
	}

	return fb, nil
}
