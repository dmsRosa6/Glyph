package geom

type Vector struct {
	X, Y int
}

func (v Vector) Add(o Vector) Vector {
	return Vector{
		X: v.X + o.X,
		Y: v.Y + o.Y,
	}
}

func (v Vector) Neg() Vector {
	return Vector{
		X: -v.X,
		Y: -v.Y,
	}
}
