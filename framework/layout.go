package framework

type AxisAnchor int

const (
	NoAnchor AxisAnchor = iota
	Start
	Center
	End
)

type Anchor struct {
	V AxisAnchor
	H AxisAnchor
}

func ResolveAxis(anchor AxisAnchor, parentSize, size, original int) int {
	switch anchor {
	case Start:
		return 0
	case Center:
		return (parentSize - size) / 2
	case End:
		return parentSize - size
	default:
		return original
	}
}
