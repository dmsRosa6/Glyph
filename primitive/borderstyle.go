package primitive

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
