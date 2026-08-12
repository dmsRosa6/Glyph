package core

type AppSignal int

const (
	NOOP AppSignal = iota
	SIGTERM
)
