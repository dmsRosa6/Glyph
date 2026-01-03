package render

type LoopMode int

const (
    OnDemand LoopMode = iota
    FixedFPS
)

type RenderMode struct {
	Mode LoopMode
    Fps int
    Redraw chan struct{}
}

func OnDemandMode() RenderMode {
    return RenderMode{
        Mode: OnDemand,
        Redraw: make(chan struct{}, 1),
    }
}

func FixedFPSMode(fps int) RenderMode {
    //TODO maybe make it like not be 4 or smth
    if fps <= 0 {
        panic("FixedFPS requires fps > 0")
    }
    return RenderMode{
        Mode: FixedFPS,
        Fps:  fps,
    }
}