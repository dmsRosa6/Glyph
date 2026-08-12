package framework

import "github.com/dmsRosa6/glyph/core"

type Logger struct {
	logs   chan<- core.AppLog
	source string
}

func NewLogger(logs chan<- core.AppLog, source string) Logger {
	return Logger{
		logs:   logs,
		source: source,
	}
}

func (l Logger) Debug(msg string) {
	l.logs <- *core.NewDebugAppLog(msg, l.source)
}

func (l Logger) Info(msg string) {
	l.logs <- *core.NewInfoAppLog(msg, l.source)
}

func (l Logger) Warning(err error) {
	l.logs <- *core.NewWarningAppLog(err, l.source)
}

func (l Logger) Fatal(err error) {
	l.logs <- *core.NewFatalAppLog(err, l.source)
}
