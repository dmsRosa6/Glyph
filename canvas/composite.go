package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/utils"
)

type Composite struct{
	bounds geom.Bounds
	Children []Drawable

	layer int
}

type CompositeConfig struct {
    Bounds geom.Bounds
	Layer int
}

//TODO probabily have like a composite with rect to accelarte things

//TODO eventually probabily is nice to have it like i do on the canvas and have colors by default on the composite
func NewComposite(cfg CompositeConfig) (*Composite, error) {
    c :=  &Composite{
        bounds:   cfg.Bounds,
        Children: []Drawable{},
    }

	if err := c.SetLayer(cfg.Layer); err != nil {
		return nil, err
	}

	return c, nil
}

func NewCompositeAt(x, y, w, h int) (*Composite, error) {
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

	c.Children = utils.InsertSortLayered(c.Children, child)
}

func (c *Composite) RemoveChild(target Drawable) {
	for i, child := range c.Children {
		if child == target {
			c.Children = append(c.Children[:i], c.Children[i+1:]...)
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