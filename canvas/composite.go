package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/utils"
)

type Composite struct{
	bounds *geom.Bounds
	children []Drawable

	layer int
}

type CompositeConfig struct {
	Layer int
}

//TODO probabily have like a composite with rect to accelarte things

//TODO eventually probabily is nice to have it like i do on the canvas and have colors by default on the composite
func NewComposite(bounds *geom.Bounds,cfg CompositeConfig) (*Composite, error) {
    c :=  &Composite{
        bounds:   bounds,
        children: []Drawable{},
    }

	if err := c.SetLayer(cfg.Layer); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Composite) Draw(buf *core.Buffer, origin geom.Point){
	t := geom.Point{}

	t.AddPoint(origin)
	t.AddPoint(*c.bounds.Pos)
	
	for _, s := range c.children {
		s.Draw(buf, t)
    }
}

func (r *Composite) IsInBounds(parent geom.Bounds) bool{
	if r.bounds.Pos.X < 0 {
		return false
	}

	if r.bounds.Pos.Y < 0 {
		return false
	}

	if r.bounds.Pos.Y + r.bounds.H > parent.H {
		return false
	}

	if r.bounds.Pos.X + r.bounds.W > parent.W {
		return false
	}

	return true
}

func (c *Composite) AddChild(child Drawable){
	
	if !child.IsInBounds(*c.bounds){
		panic("Shape out of composite bounds")
	}	

	c.children = utils.InsertSortLayered(c.children, child)
}

func (c *Composite) RemoveChild(target Drawable) {
	for i, child := range c.children {
		if child == target {
			c.children = append(c.children[:i], c.children[i+1:]...)
			return
		}
	}
}

func (r *Composite) SetLayer(l int) error{
    if l < 0{
		return errors.New("Layers must be greater or equal to 0")
	} 
	
	r.layer = l

	return nil
}

func (r *Composite) GetLayer() int{
    return r.layer
}