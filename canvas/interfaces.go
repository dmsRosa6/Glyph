package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Drawable interface{
	Draw(buf *core.Buffer, origin geom.Point)
	IsInBounds(parent geom.Bounds) bool
	SetLayer(l int) error
	GetLayer() int
}

type Moveable interface{
	MoveTo(p geom.Point)
	Translate(v geom.Vector)
}

type Layoutable interface {
	Layout(parent geom.Bounds)
}

type Composable interface {
	AddChild(child Drawable)
	RemoveChild(target Drawable)
}



