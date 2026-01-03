package render

type LoopMode int

const (
    OnDemand LoopMode = iota
    FixedFPS
)

type RenderMode struct {
	Mode LoopMode
    Fps int
}
