package render

import "errors"

type LoopMode int

const (
	FixedFPS LoopMode = iota
	OnDemand
)

type RenderMode struct {
	Mode   LoopMode
	Fps    int
	Redraw chan struct{}
}

func OnDemandMode() RenderMode {
	return RenderMode{
		Mode:   OnDemand,
		Redraw: make(chan struct{}, 1),
	}
}

// FixedFPSMode returns an error rather than panicking on fps <= 0 --
// this is a plausible caller-supplied config value (from a flag, a
// config file, etc.), not a programmer invariant violation, so it
// should be reportable rather than fatal. See geom.NewPoint's doc
// comment for the same rule applied elsewhere.
func FixedFPSMode(fps int) (RenderMode, error) {
	if fps <= 0 {
		return RenderMode{}, errors.New("FixedFPS requires fps > 0")
	}
	return RenderMode{
		Mode: FixedFPS,
		Fps:  fps,
	}, nil
}
