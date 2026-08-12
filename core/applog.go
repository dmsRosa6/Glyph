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

func (l AppLog) Reason() string {
	rsn := l.msg
	if l.severity >= Warning && l.err != nil {
		rsn = l.err.Error()
	}
	return fmt.Sprintf("[%s] %s. In component '%s'", l.severity.String(), rsn, l.source)
}

func (l AppLog) Severity() Severity {
	return l.severity
}
