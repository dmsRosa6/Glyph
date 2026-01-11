package geom

type Point struct {
	X, Y int
}

func (p *Point) AddVector(v Vector) {
	p.X = p.X + v.X
	p.Y = p.Y + v.Y
}

func (p *Point) SubVector(v Vector) {
	p.X = p.X + v.X
	p.Y = p.Y + v.Y
}

func (p *Point) AddPoint(p1 Point) {
	p.X = p.X + p1.X
	p.Y = p.Y + p1.Y
}

func (p *Point) SubPoint(p1 Point) {
	p.X = p.X + p1.X
	p.Y = p.Y + p1.Y
}