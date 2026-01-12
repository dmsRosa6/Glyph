package canvas

import "github.com/dmsRosa6/glyph/geom"

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

type Layout struct{
	computedPos *geom.Point
	anchor *Anchor
}

func resolveAxis(anchor AxisAnchor, parentStart, parentSize, size, original int) int {
    switch anchor {
    case Start:
        return parentStart
    case Center:
        return parentStart + (parentSize-size)/2
    case End:
        return parentStart + parentSize - size
    default:
        return original
    }
}

