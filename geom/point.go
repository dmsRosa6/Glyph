package geom

import "errors"

type Point struct {
	X, Y int
}

// NewPoint returns an error rather than panicking on a negative
// coordinate -- a computed position (terminal size minus a widget,
// user-supplied config, etc.) is a plausible runtime value, not a
// programmer invariant violation, so a caller should be able to check
// and report it rather than crash outright. See devguide.md's
// "construction failures" note for the panic-vs-error rule this
// follows across the codebase.
//
// NOTE: as of this change, nothing in glyph actually calls NewPoint --
// every real call site builds &Point{X: ..., Y: ...} directly as a
// struct literal (see base.NewBaseNode, widgets.NewText,
// widgets.NewWindow's title, widgets.NewTileGrid), which bypasses this
// check entirely regardless of what NewPoint itself does. Converting
// NewPoint's signature makes it consistent with the rest of the
// codebase for any FUTURE caller, but does not by itself close the gap
// -- that requires either routing those call sites through NewPoint or
// wiring geom.Bounds's already-unused Validate helpers into the
// constructors that build coordinates today (a separate todo item).
func NewPoint(x, y int) (*Point, error) {
	if x < 0 || y < 0 {
		return nil, errors.New("point coordinates must be >= 0")
	}

	return &Point{
		X: x,
		Y: y,
	}, nil
}

func (p *Point) AddVector(v Vector) {
	p.X = p.X + v.X
	p.Y = p.Y + v.Y
}

// SubVector previously duplicated AddVector's logic (both added).
// Fixed to actually subtract.
func (p *Point) SubVector(v Vector) {
	p.X = p.X - v.X
	p.Y = p.Y - v.Y
}
