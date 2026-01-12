package geom

type Point struct {
	X, Y int
}

func NewPoint(x,y int) *Point{
	if x < 0 || y < 0 {
		panic("Point coordinates need to be positive")
	}
	
	return &Point{
		X: x,
		Y: y,
	}
}

func (p *Point) AddVector(v Vector) {
	p.X = p.X + v.X
	p.Y = p.Y + v.Y
}

func (p *Point) SubVector(v Vector) {
	p.X = p.X + v.X
	p.Y = p.Y + v.Y
}