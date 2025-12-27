package shape

import "github.com/dmsRosa6/glyph/core"


type Shape interface{
	Draw(buf *core.Buffer)
}