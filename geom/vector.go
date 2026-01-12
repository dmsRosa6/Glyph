package geom

type Vector struct {
	X, Y int
}

func VectorFromPoint(p Point) *Vector{
	return &Vector{
		X: p.X,
		Y: p.Y,
	} 
}

func (v *Vector) AddVector(o Vector) {
	v.X = v.X + o.X
	v.Y = v.Y + o.Y
}

func (v *Vector) NegVector() {
	v.X = -v.X
	v.Y = -v.Y
}
