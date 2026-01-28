package render

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
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

    var renderMode RenderMode

    switch mode {
        case FixedFPS:
            renderMode = FixedFPSMode(fps) 
        case OnDemand:
            renderMode = OnDemandMode()
    }

    r := &Renderer{
        out: bufio.NewWriter(os.Stdout),
        RenderMode: renderMode, 
        ctx: ctx,
        cancel: cancel,
    }

    r.Init()

    return r
}

func (r *Renderer) Init() {
    // Hide cursor
    fmt.Fprint(r.out, "\x1b[?25l")
    // Enter alternate screen
    fmt.Fprint(r.out, "\x1b[?1049h")
    // Clear screen
    fmt.Fprint(r.out, "\x1b[2J")
    // Move cursor home
    fmt.Fprint(r.out, "\x1b[H")
}

func (r *Renderer) Run(c *canvas.Canvas) {
    var ticker *time.Ticker
    
    if(r.RenderMode.Mode == FixedFPS){
        ticker = time.NewTicker(time.Second / time.Duration(r.Fps))
        defer ticker.Stop()
    }else{
        r.Redraw <- struct{}{}
    }

    applySize := func() {
    size, err := term.TermSize()
    if err != nil {
            panic("something went wrong resizing")
        }
        c.ApplySize(size.Cols, size.Rows)
    }

    applySize()

    // Resize listener
    term.WatchResize(func() {
        applySize()
        r.render(c)
    })

    for {
        select {
        case <-r.ctx.Done():
            r.restore()
            return

        case <-r.Redraw:
            if r.Mode == OnDemand {
                r.render(c)
            }

        case <-func() <-chan time.Time {
            if ticker != nil {
                return ticker.C
            }
            return nil
        }():
            if r.Mode == FixedFPS {
                r.render(c)
            }
        }
    }
}

func (r *Renderer) RequestRedraw() {
    
    if r.RenderMode.Mode != OnDemand {
        return
    }

    select {
    case r.Redraw <- struct{}{}:
    default:
        // already requested
    }
}

func (r *Renderer) render(c *canvas.Canvas) {
    
    fmt.Fprint(r.out, "\x1b[H") //move cursor home

    c.Compose()

    r.Flush(c.Buf)
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


func (r *Renderer) Flush(buf *core.Buffer){
    for y := 0; y < buf.H; y++ {
        for x := 0; x < buf.W; x++ {
            cell := buf.Cells[y][x]
            r.out.WriteString(fmt.Sprintf("\x1b[%d;%dH%s", y+1, x+1, term.CellToANSI(*cell)))
        }
    }
}