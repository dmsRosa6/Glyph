package render

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/term"
)

type Renderer struct {
    out *bufio.Writer
    RenderMode
    isDirty bool

    ctx    context.Context
    cancel context.CancelFunc
}

func NewRenderer(mode LoopMode, fps int) *Renderer {

    ctx, cancel := context.WithCancel(context.Background())

    r := &Renderer{
        out: bufio.NewWriter(os.Stdout),
        RenderMode: RenderMode{Mode: mode, Fps: fps},
        ctx: ctx,
        cancel: cancel,
    }

    r.Init()

    return r
}

//TODO probably pass this to another palce
func (r *Renderer) Init() {
    os.Stdout.Write([]byte("\x1b[?1049h")) // alt screen
    os.Stdout.Write([]byte("\x1b[3J"))     // clear scrollback
    os.Stdout.Write([]byte("\x1b[2J"))     // clear screen
    os.Stdout.Write([]byte("\x1b[H"))      // move cursor home
}

func (r *Renderer) Run(c *canvas.Canvas) {
    ticker := time.NewTicker(time.Second / time.Duration(r.Fps))
    defer ticker.Stop()

    applySize := func() {
    size, err := term.TermSize()
    if err != nil {
            return
        }
        c.ApplySize(size.Cols, size.Rows)
    }

    applySize()

    // Resize listener
    term.WatchResize(func() {
        applySize()
    })
    
    for {

        select {
        case <-r.ctx.Done():
            return
        default:
            // keep working
        }

        switch r.Mode {

        case OnDemand:
            if c.IsDirty {
                r.render(c)
                c.IsDirty = false
            }
            time.Sleep(1 * time.Millisecond)

        case FixedFPS:
            <-ticker.C
            r.render(c)
        }
    }
}


func (r *Renderer) render(c *canvas.Canvas) {
    fmt.Fprint(r.out, "\x1b[H")

    c.Compose()

    buf := c.Buf

    for y := 0; y < buf.H; y++ {
        for x := 0; x < buf.W; x++ {
            r.out.WriteString(cellToANSI(*buf.Cells[y][x]))
        }
        r.out.WriteByte('\n')
    }

    r.out.Flush()
}

func (r *Renderer) restore() {
    fmt.Fprint(r.out, "\x1b[?1049l") // leave alt screen
    r.out.Flush()
}

func (r *Renderer) Stop() {
    r.cancel()
    r.restore()
}
