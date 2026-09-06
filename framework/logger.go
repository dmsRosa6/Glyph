package framework

import "github.com/dmsRosa6/glyph/core"

// Logger is the ONE structured logging path in this codebase -- used
// by the widget tree (base/canvas/widgets) via BaseNode.Logger()/
// FocusBehavior's own logger, and by app/render/input for their own
// startup/shutdown/dispatch logging. Before this, app.go/render.go/
// input/manager.go each did raw `logs <- *core.NewXAppLog(...)` sends
// by hand at every call site while widgets went through this type --
// two conventions for the same job, and the five-plus copy-pasted
// construct-and-send blocks in app.go in particular. Routing everything
// through Logger collapses that to one convention and one place that
// knows how a log is allowed to be sent.
//
// The zero value is a valid, silent no-op -- same contract as
// BaseNode.Invalidate -- so a widget built before being attached to a
// Canvas never blocks or panics on a nil channel.
type Logger struct {
	logs        chan<- core.AppLog
	minSeverity core.Severity
	source      string
	id          string
}

func NewLogger(logs chan<- core.AppLog, minSeverity core.Severity, source, id string) Logger {
	return Logger{logs: logs, minSeverity: minSeverity, source: source, id: id}
}

// Enabled reports whether a message at sev would actually be sent
// anywhere, without building one. Callers on a genuinely hot path
// (App.Run's per-keystroke/per-mouse-event dispatch, in particular)
// should guard an expensive fmt.Sprintf with this BEFORE calling
// Debug/Info -- Debug("expensive message")'s argument is evaluated
// unconditionally by Go regardless of what Debug does internally, so
// only a guard at the call site actually avoids paying for it. Debug/
// Info/Warning/Fatal below still re-check severity themselves too, for
// every OTHER call site that isn't hot enough to bother with this.
func (l Logger) Enabled(sev core.Severity) bool {
	return l.logs != nil && sev >= l.minSeverity
}

// Debug and Info are the high-volume, low-stakes levels -- a widget
// focusing/blurring, a key being decoded, a frame being requested. Two
// things protect the rest of this codebase from them:
//
//  1. Enabled(sev) is checked here too (not just at hot call sites),
//     so a message built without bothering to guard still costs only
//     an AppLog struct, not a channel operation, once filtered out.
//  2. The send itself is non-blocking (see sendBestEffort): if
//     FaultManager's writer goroutine ever stalls (slow disk, a wedged
//     filesystem) and its 100-slot buffer fills, a Debug/Info log is
//     dropped rather than blocking whatever called Logger.Debug/Info --
//     which, on App.Run's dispatch loop, would otherwise mean ALL input
//     handling stalls right along with the log writer. Same drop-not-
//     block convention input.Manager.send already uses for events.
//
// Warning and Fatal deliberately do NOT drop: they're rare by
// definition, and Fatal specifically is what FaultManager promotes to
// a SIGTERM shutdown (see fault.FaultManager.run) -- silently dropping
// one under backpressure would silently disable that escape hatch
// exactly when something has already gone wrong enough to reach for it.
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
