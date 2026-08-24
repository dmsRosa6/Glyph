package core

import "fmt"

type Severity int

const (
	Debug Severity = iota
	Info
	Warning
	Fatal
)

func (s Severity) String() string {
	switch s {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Fatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type AppLog struct {
	severity Severity
	msg      string
	source   string
	id       string
	err      error
}

func NewWarningAppLog(err error, source string) *AppLog {
	return &AppLog{
		severity: Warning,
		err:      err,
		source:   source,
	}
}

func NewFatalAppLog(err error, source string) *AppLog {
	return &AppLog{
		severity: Fatal,
		err:      err,
		source:   source,
	}
}

func NewInfoAppLog(msg, source string) *AppLog {
	return &AppLog{
		severity: Info,
		msg:      msg,
		source:   source,
	}
}

func NewDebugAppLog(msg, source string) *AppLog {
	return &AppLog{
		severity: Debug,
		msg:      msg,
		source:   source,
	}
}

// WithID attaches a specific instance id to an already-built log,
// without touching the four constructors above -- so the top-level
// App/Renderer/Input logs (which have no per-instance id, just a fixed
// core.InternalSource) keep working exactly as before, and only
// widget-level logs (routed through framework.Logger, which does know
// a BaseNode's id) opt into the extra detail. Returns the same *AppLog
// so it chains directly off a constructor call.
func (l *AppLog) WithID(id string) *AppLog {
	l.id = id
	return l
}

// Reason formats source (the component's *kind*, e.g. "Text") and, when
// present, id (the specific *instance*, e.g. "Text#4" or a caller-chosen
// "scoreLabel") together -- source alone can't tell two Text widgets'
// log lines apart, id alone loses the readable "this is a Text" context
// once a caller has overridden it to something like "scoreLabel". They
// complement each other; neither replaces the other in the log line.
func (l AppLog) Reason() string {
	rsn := l.msg
	if l.severity >= Warning && l.err != nil {
		rsn = l.err.Error()
	}
	if l.id != "" {
		return fmt.Sprintf("[%s] %s. In component '%s' (id: %s)", l.severity.String(), rsn, l.source, l.id)
	}
	return fmt.Sprintf("[%s] %s. In component '%s'", l.severity.String(), rsn, l.source)
}

func (l AppLog) Severity() Severity {
	return l.severity
}

func (l AppLog) ID() string {
	return l.id
}
