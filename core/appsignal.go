package core

type AppSignal int

const (
	NOOP AppSignal = iota
	SIGTERM
)

func (a AppSignal) String() string {
	switch a {
	case NOOP:
		return "NO OP"

	case SIGTERM:
		return "SIGTERM"

	default:
		return ""
	}
}
