package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Composite struct{
	Bounds geom.Bounds
	Children []Drawable
}

//TODO probabily have like a composite with rect to accelarte things

//TODO eventually probabily is nice to have it like i do on the canvas and have colors by default on the composite
func NewComposite(x, y, w, h int) *Composite{
	return &Composite{
		Bounds: geom.Bounds{
			Pos: geom.Point{X: x, Y: y},
			W:   w,
			H:   h,
		},
		Children: []Drawable{},
	}
}

func (c *Composite) Draw(buf *core.Buffer){
	for _, s := range c.Children {
		s.Draw(buf)
    }
}

func (r *Composite) IsInBounds(parent geom.Bounds) bool{
	if r.Bounds.Pos.X < parent.Pos.X {
		return false
	}

	if r.Bounds.Pos.Y < parent.Pos.Y {
		return false
	}

	if r.Bounds.Pos.Y + r.Bounds.H > parent.Pos.Y + parent.H {
		return false
	}

	if r.Bounds.Pos.X + r.Bounds.W > parent.Pos.X + parent.W {
		return false
	}

	return true
}

func (c *Composite) AddChild(child Drawable){
	
	if !child.IsInBounds(c.Bounds){
		panic("Shape out of composite bounds")
	}	

	c.Children = append(c.Children, child)
}
