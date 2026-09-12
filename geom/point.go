package geom

import "errors"

type Point struct {
	X, Y int
}

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

func (p *Point) SubVector(v Vector) {
	p.X = p.X - v.X
	p.Y = p.Y - v.Y
}
