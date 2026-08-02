package canvas

import (
	"sort"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

// Container is the one generic "holds children" primitive. What used to
// be three separate hand-written types (Composite, List, and the
// composite-half of Box) are now this one type parameterized by a
// LayoutPolicy:
//   - Container{Layout: FreeLayout{}}  behaves like the old Composite
//   - Container{Layout: StackLayout{}} behaves like the old List
// A Container can hold ANY Drawable -- another Container, a Bordered box,
// a bare Text or Rect -- there's no restriction to a specific child type
// the way the old List only accepted *Box.
type Container struct {
	BaseNode
	children []Drawable
	layout   LayoutPolicy
}

type ContainerConfig struct {
	Style  Style
	Layer  int
	Anchor Anchor
	Layout LayoutPolicy // defaults to FreeLayout if nil
}

func NewContainer(bounds *geom.Bounds, cfg ContainerConfig) (*Container, error) {
	base, err := newBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	policy := cfg.Layout
	if policy == nil {
		policy = FreeLayout{}
	}

	return &Container{
		BaseNode: base,
		children: []Drawable{},
		layout:   policy,
	}, nil
}

func (c *Container) Draw(buf *core.Buffer, vec geom.Vector) {
	v := geom.Vector{X: vec.X + c.computedPos.X, Y: vec.Y + c.computedPos.Y}
	frame := c.LocalFrame()

	// Sort by current layer right before drawing, not just once at
	// AddChild time. A child's layer can change after it's added via
	// SetLayer -- sorting here (stable, so equal layers keep insertion
	// order) means draw order always reflects whatever GetLayer()
	// reports right now, instead of going stale the moment someone
	// re-layers a child post-insertion.
	sort.SliceStable(c.children, func(i, j int) bool {
		return c.children[i].GetLayer() < c.children[j].GetLayer()
	})

	c.layout.Arrange(c.children, frame)

	for _, child := range c.children {
		child.Draw(buf, v)
	}
}

// AddChild checks the child against this container's own interior size
// (LocalFrame), not against c.bounds directly -- c.bounds.Pos is where
// the container itself sits in *its* parent, a different coordinate
// space than where a child sits within the container. Insertion order
// doesn't matter here anymore since Draw sorts by layer every frame.
func (c *Container) AddChild(child Drawable) {
	if !child.IsInBounds(c.LocalFrame()) {
		panic("shape out of container bounds")
	}

	child.SetParentStyle(c.ResolvedStyle())
	child.SetInvalidator(c.invalidate)

	c.children = append(c.children, child)
}

func (c *Container) RemoveChild(target Drawable) {
	for i, child := range c.children {
		if child == target {
			c.children = append(c.children[:i], c.children[i+1:]...)
			return
		}
	}
}

func (c *Container) SetParentStyle(s *Style) {
	c.BaseNode.SetParentStyle(s)
	for _, child := range c.children {
		child.SetParentStyle(c.ResolvedStyle())
	}
}

func (c *Container) SetInvalidator(fn func()) {
	c.BaseNode.SetInvalidator(fn)
	for _, child := range c.children {
		child.SetInvalidator(fn)
	}
}
