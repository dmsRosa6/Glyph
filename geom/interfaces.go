package geom

import "github.com/dmsRosa6/glyph/core"

//Maybe separate the markasDIrty
type Drawable interface{
	Draw(buf *core.Buffer)
	MarkAsDirty()
}


type Moveable interface{
	MoveTo(p Point)
	Translate(v Vector)
}