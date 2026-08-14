package widgets

// BorderStyle is a comparable value type (see border.go's `style ==
// (BorderStyle{})` zero-value checks), so it's plain runes, not pointers.
type BorderStyle struct {
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	Horizontal  rune
	Vertical    rune
}

var (
	EmptyBorder = BorderStyle{}

	SingleLine = BorderStyle{
		TopLeft: '+', TopRight: '+', BottomLeft: '+', BottomRight: '+',
		Horizontal: '-', Vertical: '|',
	}

	Rounded = BorderStyle{
		TopLeft: '╭', TopRight: '╮', BottomLeft: '╰', BottomRight: '╯',
		Horizontal: '─', Vertical: '│',
	}

	DoubleLine = BorderStyle{
		TopLeft: '╔', TopRight: '╗', BottomLeft: '╚', BottomRight: '╝',
		Horizontal: '═', Vertical: '║',
	}
)
