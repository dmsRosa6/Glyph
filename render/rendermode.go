package render

import "errors"

type LoopMode int

const (
	FixedFPS LoopMode = iota
	OnDemand
)

// RenderMode configures how Renderer drives its frame loop: a fixed
// timer (FixedFPS) or purely on-demand redraws (OnDemand).
//
// Fields are unexported so the only way to build a RenderMode from
// outside this package is through OnDemandMode() or FixedFPSMode(fps)
// below. Before this, a hand-built RenderMode{Mode: OnDemand} literal
// compiled fine but skipped OnDemandMode's Redraw channel allocation,
// leaving Redraw nil -- Renderer.Run would then block forever on
// startup with no error and no panic, just silence.
//
// The all-zero RenderMode{} is still legal Go from any package (an
// empty composite literal never needs field access, unexported or
// not), and it is still NOT a valid mode: it reads as
// {mode: FixedFPS, fps: 0}, which used to divide by zero the moment
// Renderer.Run built its ticker. That's exactly the shape
// AppConfig{} produces when RenderMode is left unset, so unexporting
// the fields alone doesn't close the gap -- NewRenderer additionally
// calls valid() below and returns an error instead of trusting
// whatever it was handed.
type RenderMode struct {
	mode   LoopMode
	fps    int
	redraw chan struct{}
}

// valid reports whether m was actually produced by OnDemandMode or
// FixedFPSMode, as opposed to a zero-value RenderMode{}.
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
		mode: FixedFPS,
		fps:  fps,
	}, nil
}
