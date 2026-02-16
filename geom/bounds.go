package geom

type Bounds struct {
    Pos *Point
    W, H int }

//TODO Decidee where i want the verification    
func NewBounds(x, y, w, h int) *Bounds {
    if w <= 0 || h <= 0 {
        //panic("bounds width and height must be > 0")
    }
    return &Bounds{
        Pos: &Point{X: x, Y: y},
        W:   w,
        H:   h,
    }
}

func (b *Bounds) Validate() {
    if b.Pos.Y < 0 || b.Pos.X < 0 || b.W < 0 || b.H < 0 {
        panic("bounds width and height must be > 0")
    }
}

func (b *Bounds) ValidateIfInsideBounds(other Bounds) {
    if b.Pos.X < 0 || b.Pos.Y < 0 || b.W < 0 || b.H < 0 {
        panic("bounds width and height must be >= 0")
    }
    if b.Pos.X+b.W > other.W || b.Pos.Y+b.H > other.H {
        panic("bounds exceed parent bounds")
    }
}


func (b *Bounds) ValidateNoPanic() bool{
    return b.Pos.Y >= 0 && b.Pos.X >= 0 && b.W >= 0 && b.H >= 0
}