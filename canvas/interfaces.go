package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

//Maybe separate the markasDIrty
type Drawable interface{
	Draw(buf *core.Buffer)
}


type Moveable interface{
	MoveTo(p geom.Point)
	Translate(v geom.Vector)
}