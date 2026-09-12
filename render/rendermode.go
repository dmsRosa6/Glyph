package render

import "errors"

type LoopMode int

const (
	FixedFPS LoopMode = iota
	OnDemand
)

type RenderMode struct {
	mode   LoopMode
	fps    int
	redraw chan struct{}
}

func (m RenderMode) valid() bool {
	switch m.mode {
	case FixedFPS:
		return m.fps > 0
	case OnDemand:
		return m.redraw != nil
	default:
		return false
	}
}

func OnDemandMode() RenderMode {
	return RenderMode{
		mode:   OnDemand,
		redraw: make(chan struct{}, 1),
	}
}

func FixedFPSMode(fps int) (RenderMode, error) {
	if fps <= 0 {
		return RenderMode{}, errors.New("FixedFPS requires fps > 0")
	}
	return RenderMode{
		mode: FixedFPS,
		fps:  fps,
	}, nil
}
