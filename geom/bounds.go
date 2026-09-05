package geom

type Bounds struct {
	Pos  *Point
	W, H int
}

func NewBounds(x, y, w, h int) *Bounds {
	return &Bounds{
		Pos: &Point{X: x, Y: y},
		W:   w,
		H:   h,
	}
}

// Valid reports whether this Bounds describes a well-formed rectangle:
// non-negative position, width, and height. Renamed from
// ValidateNoPanic -- that name only made sense in contrast to a
// panicking sibling. There were two: Validate() (identical negativity
// check, panicked instead of returning bool) and
// ValidateIfInsideBounds (the same negativity check PLUS a containment
// check against a parent Bounds). Both are gone: Validate() panicking
// on a plausible runtime value (a widget's declared width/height)
// contradicted the panic-vs-error rule already applied to
// NewBorder/geom.NewPoint/render.FixedFPSMode elsewhere in this
// codebase, and ValidateIfInsideBounds's containment check duplicated
// base.BaseNode.IsInBounds, which already does the same check, already
// returns bool, and is already live (Container.Draw's out-of-bounds
// warning) -- keeping both was two sources of truth for one rule.
//
// Wired into base.NewBaseNode, the one construction point nearly every
// Bounds in this codebase eventually flows through.
func (b *Bounds) Valid() bool {
	return b.Pos.Y >= 0 && b.Pos.X >= 0 && b.W >= 0 && b.H >= 0
}
