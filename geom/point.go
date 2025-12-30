package geom

type Point struct {
	X, Y int
}

func (p Point) Add(v Vector) Point {
	return Point{
		X: p.X + v.X,
		Y: p.Y + v.Y,
	}
}

func (p Point) Sub(v Vector) Point {
	return Point{
		X: p.X - v.X,
		Y: p.Y - v.Y,
	}
}
