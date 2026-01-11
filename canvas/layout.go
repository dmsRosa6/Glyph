package canvas

type Anchor int

const (
	NoAnchor Anchor = iota
	AnchorTop
	AnchorCenter
	AnchorBottom
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
