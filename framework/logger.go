package framework

import "github.com/dmsRosa6/glyph/core"

// Logger is the structured logging path for the widget tree (base/canvas/
// widgets). The zero value is a valid, silent no-op -- same contract as
// BaseNode.Invalidate -- so a widget built before being attached to a
// Canvas never blocks or panics on a nil channel.
type Logger struct {
	logs   chan<- core.AppLog
	source string
}

func NewLogger(logs chan<- core.AppLog, source string) Logger {
	return Logger{logs: logs, source: source}
}

func (l Logger) Debug(msg string) {
	if l.logs == nil {
		return
	}
	l.logs <- *core.NewDebugAppLog(msg, l.source)
}

func (l Logger) Info(msg string) {
	if l.logs == nil {
		return
	}
	l.logs <- *core.NewInfoAppLog(msg, l.source)
}

func (l Logger) Warning(err error) {
	if l.logs == nil || err == nil {
		return
	}
	l.logs <- *core.NewWarningAppLog(err, l.source)
}

func (l Logger) Fatal(err error) {
	if l.logs == nil || err == nil {
		return
	}
	l.logs <- *core.NewFatalAppLog(err, l.source)
}
