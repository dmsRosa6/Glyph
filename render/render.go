package render

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dmsRosa6/glyph/core"
)

type Renderer struct {
    out *bufio.Writer
}

func NewRenderer() *Renderer {
    return &Renderer{
        out: bufio.NewWriter(os.Stdout),
    }
}

//TODO probably pass this to another palce
func (r *Renderer) Init() {
    os.Stdout.Write([]byte("\x1b[?1049h")) // alt screen
    os.Stdout.Write([]byte("\x1b[3J"))     // clear scrollback
    os.Stdout.Write([]byte("\x1b[2J"))     // clear screen
    os.Stdout.Write([]byte("\x1b[H"))      // move cursor home
}

func (r *Renderer) Render(buf *core.Buffer) {
    fmt.Fprint(r.out, "\x1b[H")

    for y := 0; y < buf.H; y++ {
        for x := 0; x < buf.W; x++ {
            r.out.WriteRune(buf.Cells[y][x].Ch)
        }
        r.out.WriteByte('\n')
    }

    r.out.Flush()
}

func (r *Renderer) Restore() {
    fmt.Fprint(r.out, "\x1b[?1049l") // leave alt screen
    r.out.Flush()
}
