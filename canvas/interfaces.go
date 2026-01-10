package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Drawable interface{
	Draw(buf *core.Buffer)
	IsInBounds(parent geom.Bounds) bool
	SetLayer(l int) error
	GetLayer() int
}

type Moveable interface{
	MoveTo(p geom.Point)
	Translate(v geom.Vector)
}

type Composable interface {
	Layout(parent geom.Bounds)
}

type Layoutable interface {
	AddChild(child Drawable)
	RemoveChild(target Drawable)
}

