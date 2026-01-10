package canvas

type Anchor int

const (
	AnchorTopLeft Anchor = iota
	AnchorTop
	AnchorTopRight
	AnchorLeft
	AnchorCenter
	AnchorRight
	AnchorBottomLeft
	AnchorBottom
	AnchorBottomRight
)

type Layout struct{
	anchor Anchor
}

type LayoutConfig struct{
	Anchor Anchor
}

func NewLayout(lc LayoutConfig) *Layout{
	return &Layout{anchor: lc.Anchor}
}
