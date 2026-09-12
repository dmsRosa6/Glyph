package framework

import "github.com/dmsRosa6/glyph/core"

type Logger struct {
	logs        chan<- core.AppLog
	minSeverity core.Severity
	source      string
	id          string
}

func NewLogger(logs chan<- core.AppLog, minSeverity core.Severity, source, id string) Logger {
	return Logger{logs: logs, minSeverity: minSeverity, source: source, id: id}
}

func (l Logger) Enabled(sev core.Severity) bool {
	return l.logs != nil && sev >= l.minSeverity
}

func (l Logger) Debug(msg string) {
	if !l.Enabled(core.Debug) {
		return
	}
	l.sendBestEffort(*core.NewDebugAppLog(msg, l.source).WithID(l.id))
}

func (l Logger) Info(msg string) {
	if !l.Enabled(core.Info) {
		return
	}
	l.sendBestEffort(*core.NewInfoAppLog(msg, l.source).WithID(l.id))
}

func (l Logger) Warning(err error) {
	if err == nil || !l.Enabled(core.Warning) {
		return
	}
	l.logs <- *core.NewWarningAppLog(err, l.source).WithID(l.id)
}

func (l Logger) Fatal(err error) {
	if err == nil || l.logs == nil {
		return
	}
	l.logs <- *core.NewFatalAppLog(err, l.source).WithID(l.id)
}

func (l Logger) sendBestEffort(log core.AppLog) {
	select {
	case l.logs <- log:
	default:
	}
}
