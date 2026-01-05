package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Composite struct{
	bounds geom.Bounds
	Children []Drawable
}

type CompositeConfig struct {
    Bounds geom.Bounds
}

//TODO probabily have like a composite with rect to accelarte things

//TODO eventually probabily is nice to have it like i do on the canvas and have colors by default on the composite
func NewComposite(cfg CompositeConfig) *Composite {
    return &Composite{
        bounds:   cfg.Bounds,
        Children: []Drawable{},
    }
}

func NewCompositeAt(x, y, w, h int) *Composite {
    return NewComposite(CompositeConfig{
        Bounds: geom.Bounds{
            Pos: geom.Point{X: x, Y: y},
            W:   w,
            H:   h,
        },
    })
}

func (c *Composite) Draw(buf *core.Buffer){
	for _, s := range c.Children {
		s.Draw(buf)
    }
}

func (r *Composite) IsInBounds(parent geom.Bounds) bool{
	if r.bounds.Pos.X < parent.Pos.X {
		return false
	}

	if r.bounds.Pos.Y < parent.Pos.Y {
		return false
	}

	if r.bounds.Pos.Y + r.bounds.H > parent.Pos.Y + parent.H {
		return false
	}

	if r.bounds.Pos.X + r.bounds.W > parent.Pos.X + parent.W {
		return false
	}

	return true
}

func (c *Composite) AddChild(child Drawable){
	
	if !child.IsInBounds(c.bounds){
		panic("Shape out of composite bounds")
	}	

	c.Children = append(c.Children, child)
}

func (c *Composite) RemoveChild(target Drawable) {
	for i, child := range c.Children {
		if child == target {
			c.Children = append(c.Children[:i], c.Children[i+1:]...)
			return
		}
	}
}

