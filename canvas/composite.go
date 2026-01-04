package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Composite struct{
	Pos geom.Point
	Bounds geom.Bounds
	Children []Drawable
}

func (c *Composite) Draw(buf *core.Buffer){
	for _, s := range c.Children {
		s.Draw(buf)
    }
}

func (c *Composite) AddChild(child Drawable){
	
	if !child.IsInBounds(c.Pos,c.Bounds){
		panic("Shape out of composite bounds")
	}	

	c.Children = append(c.Children, child)
}
