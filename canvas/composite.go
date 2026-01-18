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
	parentStyle *Style
	style Style
	layer int
	layout *Layout
}

type CompositeConfig struct {
	Style Style
	Layer int
	Anchor Anchor
}

func NewComposite(bounds *geom.Bounds, cfg CompositeConfig) (*Composite, error) {
    c :=  &Composite{
        bounds:   bounds,
        children: []Drawable{},
		layout: &Layout{
					computedPos: bounds.Pos,
					anchor: &cfg.Anchor,
				},
		style: cfg.Style,
		}

	if err := c.SetLayer(cfg.Layer); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Composite) Draw(buf *core.Buffer, vec geom.Vector){
	v := geom.Vector{}

	v.AddVector(vec)
	v.AddVector(*geom.VectorFromPoint(*c.layout.computedPos))
	
	for _, s := range c.children {
		s.Draw(buf, v)
    }
}

func (r *Composite) IsInBounds(parent geom.Bounds) bool{
	if r.layout.computedPos.X < 0 {
		return false
	}

	if r.layout.computedPos.Y < 0 {
		return false
	}

	if r.layout.computedPos.Y + r.bounds.H > parent.H {
		return false
	}

	if r.layout.computedPos.X + r.bounds.W > parent.W {
		return false
	}

	return true
}

func (c *Composite) AddChild(child Drawable){
	
	if !child.IsInBounds(*c.bounds){
		panic("Shape out of composite bounds")
	}	

	child.SetParentStyle(c.parentStyle)

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

func (c *Composite) Layout(parent geom.Bounds) {
    c.layout.computedPos.X = resolveAxis(c.layout.anchor.H, parent.Pos.X, parent.W, c.bounds.W, c.bounds.Pos.X)
    c.layout.computedPos.Y = resolveAxis(c.layout.anchor.H, parent.Pos.Y, parent.H, c.bounds.H, c.bounds.Pos.Y)
}


func (c *Composite) SetParentStyle(s *Style){
    c.parentStyle = s
}