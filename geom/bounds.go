package geom

import "errors"

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

func (b *Bounds) Valid() error {
	if b.Pos.Y < 0 || b.Pos.X < 0 || b.W < 0 || b.H < 0 {
		return errors.New("bounds width and height must be > 0")
	}
	return nil
}
