package canvas

type BorderStyle struct {
    TopLeft     rune
    TopRight    rune
    BottomLeft  rune
    BottomRight rune
    Horizontal  rune
    Vertical    rune
}

func UniformBorderStyle(ch rune) BorderStyle {
    return BorderStyle{
        TopLeft:     ch,
        TopRight:    ch,
        BottomLeft:  ch,
        BottomRight: ch,
        Horizontal:  ch,
        Vertical:    ch,
    }
}

var SingleLine = BorderStyle{
    TopLeft: '┌', TopRight: '┐',
    BottomLeft: '└', BottomRight: '┘',
    Horizontal: '─', Vertical: '│',
}

var DoubleLine = BorderStyle{
    TopLeft: '╔', TopRight: '╗',
    BottomLeft: '╚', BottomRight: '╝',
    Horizontal: '═', Vertical: '║',
}

var ThickLine = BorderStyle{
    TopLeft: '┏', TopRight: '┓',
    BottomLeft: '┗', BottomRight: '┛',
    Horizontal: '━', Vertical: '┃',
}

var Rounded = BorderStyle{
    TopLeft: '╭', TopRight: '╮',
    BottomLeft: '╰', BottomRight: '╯',
    Horizontal: '─', Vertical: '│',
}   

var EmptyBorder = BorderStyle{
    TopLeft: ' ', TopRight: ' ',
    BottomLeft: ' ', BottomRight: ' ',
    Horizontal: ' ', Vertical: ' ',
}