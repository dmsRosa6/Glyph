package framework

type Severity int

const (
	Debug Severity = iota
	Info
	Warning
	Fatal
)

type AppLog struct {
	severity Severity
	msg      string
	source   string
	err      error
}

func NewWarningAppLog(err error) *AppLog {
	return &AppLog{severity: Warning, err: err}
}

func NewFatalAppLog(err error) *AppLog {
	return &AppLog{severity: Fatal, err: err}
}

func NewInfoAppLog(msg string) *AppLog {
	return &AppLog{severity: Info, msg: msg}
}

func NewDebugAppLog(msg string) *AppLog {
	return &AppLog{severity: Debug, msg: msg}
}
