package framework

type Severity int

const (
	Warning Severity = iota
	Fatal
)

type Fault struct {
	Err      error
	Severity Severity
	Source   string
}
