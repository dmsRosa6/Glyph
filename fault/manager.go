package fault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/datastructs"
)

// Config controls how a FaultManager persists logs. The zero value is
// safe to use as-is: LogLevel's own zero value is core.Warning (see
// core.Severity's doc comment), LogDir empty falls back to the
// package's historical default ("logs", relative to cwd), and
// DisableFileLog false keeps the original always-write-to-disk
// behavior. So app.NewApp(app.AppConfig{}) no longer means "log
// everything, forever, into an unconfigurable directory" -- it means
// "log Warning and above, into ./logs, same as always."
type Config struct {
	LogLevel core.Severity
	// LogDir overrides where per-run log files are written. Empty uses
	// the default ("logs"). Ignored if DisableFileLog is true.
	LogDir string
	// DisableFileLog skips creating a log directory/file entirely --
	// FaultManager still filters by LogLevel and still promotes a
	// Fatal-severity log to a SIGTERM shutdown (see run()'s handle
	// closure), it just never writes a line to disk. For embedding
	// glyph in a host app that doesn't want it unconditionally
	// littering the host's cwd with logs/log_*.txt files.
	DisableFileLog bool
}

// FaultManager bundles two separable jobs under one name: it's a
// generic log sink (buffer incoming AppLogs, write them to a file,
// retry on transient write failures), AND it's the thing that promotes
// a Fatal-severity log into a SIGTERM shutdown signal (see run()'s
// handle closure). Those two responsibilities can't currently be
// opted into independently -- there's no way to get file logging
// without also getting auto-shutdown-on-fatal. A full split (or at
// least a rename making that dual role obvious without reading this
// far) is a real API change with call sites across app/render/input;
// left as a deliberate "consider" rather than done here, since nothing
// today actually needs the two decoupled -- App is FaultManager's only
// caller, and it always wants both.
type FaultManager struct {
	appSignal chan<- core.AppSignal
	log       chan core.AppLog
	logLevel  core.Severity
	cancel    context.CancelFunc
	ctx       context.Context
	done      chan struct{}
	// file is opened synchronously in NewFaultManager -- see its doc
	// comment for why this can no longer happen inside run(). nil when
	// Config.DisableFileLog is set; run() treats nil as "don't persist,"
	// not as a bug to guard against everywhere -- see run()'s own notes.
	file *os.File
}

const logFileName string = "log_%s.txt"
const basePath string = "logs"

// NewFaultManager does its fallible setup -- creating the log
// directory and opening this run's log file -- synchronously, right
// here, rather than deferring it into the run() goroutine Start()
// launches later. Previously this constructor always returned a nil
// error despite its signature, and os.MkdirAll/os.OpenFile only
// actually ran once Start() called go f.run(): a permissions problem,
// a read-only filesystem, or a full disk would panic that detached
// goroutine and take the whole process down with it, bypassing
// App.Stop() and any graceful terminal restore -- and since that
// goroutine isn't started until App.Run() (before raw mode is
// engaged), this usually failed before the terminal was touched, but
// "usually" is a goroutine race, not a guarantee. Doing it here means
// a bad log path is reported the same way every other fallible
// constructor in this codebase reports a bad config: as a returned
// error the caller (app.NewApp) already checks and surfaces, before
// anything is running.
func NewFaultManager(cfg Config, signals chan core.AppSignal) (*FaultManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.DisableFileLog {
		return &FaultManager{
			appSignal: signals,
			log:       make(chan core.AppLog, 100),
			logLevel:  cfg.LogLevel,
			ctx:       ctx,
			cancel:    cancel,
			done:      make(chan struct{}),
			file:      nil,
		}, nil
	}

	dir := cfg.LogDir
	if dir == "" {
		dir = basePath
	}

	// MkdirAll is a no-op (nil error) if dir already exists, so this
	// replaces the old os.Stat/IsNotExist check outright rather than
	// needing to keep both.
	if err := os.MkdirAll(dir, 0755); err != nil {
		cancel()
		return nil, fmt.Errorf("could not create log directory %q: %w", dir, err)
	}

	ts := time.Now().Format("20060102_150405")
	resolvedFileName := filepath.Join(dir, fmt.Sprintf(logFileName, ts))

	file, err := os.OpenFile(resolvedFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not open log file %q: %w", resolvedFileName, err)
	}

	return &FaultManager{
		appSignal: signals,
		log:       make(chan core.AppLog, 100),
		logLevel:  cfg.LogLevel,
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
		file:      file,
	}, nil
}

func (f *FaultManager) Logs() chan<- core.AppLog {
	return f.log
}

func (f *FaultManager) Start() {
	go f.run()
}

func (f *FaultManager) Stop() {
	f.cancel()
	<-f.done
}

func (f *FaultManager) run() {
	defer close(f.done)
	file := f.file
	defer func() {
		if file == nil {
			return // DisableFileLog: nothing was ever opened
		}
		// A close failure here is reported, not panicked -- by the time
		// this runs the app may already be mid-shutdown, and taking the
		// whole process down over a failed Close() is exactly the
		// disproportionate failure mode this constructor was just
		// rewritten to avoid on the open side. Nothing else is set up
		// to receive this (the log channel this file itself backs is
		// going away), so stderr is the only honest place left for it.
		if cerr := file.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "fault: error closing log file: %v\n", cerr)
		}
	}()

	// TODO: Make retry buffer capacity configurable.
	pending := datastructs.NewRingBuffer(100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// write appends logLine to the log file, or queues it in pending on
	// a write failure so the 1s ticker below can retry it. A nil file
	// (DisableFileLog) makes this a deliberate no-op -- and since it's
	// the only place pending ever gets appended to, pending stays
	// permanently empty in that mode too: nothing is ever attempted, so
	// nothing ever needs retrying. Every pending.Size() > 0 check
	// elsewhere in this function can safely assume file != nil as a
	// result, rather than re-checking it every time.
	write := func(logLine string) {
		if file == nil {
			return
		}
		if pending.Size() > 0 {
			pending.Add(logLine) // keep queued lines in order rather than writing out of sequence
			return
		}
		if _, err := file.WriteString(logLine); err != nil {
			pending.Add(logLine)
		}
	}

	// handle is one queued AppLog's full journey: severity filter,
	// format, write-or-queue, and -- for Fatal specifically -- promote
	// to a SIGTERM shutdown. Pulled out into its own closure so the
	// normal receive case below and the shutdown-drain case (see
	// ctx.Done()) apply IDENTICAL handling to a log, rather than the
	// drain path needing its own hand-rolled copy that could quietly
	// drift from this one.
	handle := func(log core.AppLog) {
		if log.Severity() < f.logLevel {
			return
		}

		ts := time.Now().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("%s %s\n", ts, log.Reason())

		// Big fuck up this is the case where we just want to get over it we will not wait ticks to write to log
		if log.Severity() == core.Fatal {
			if file != nil {
				if _, err := file.WriteString(logLine); err != nil {
					fmt.Printf("could not write fatal log: %v\n%s", err, logLine)
				}
			}
			f.appSignal <- core.SIGTERM
			return
		}

		write(logLine)
	}

	for {
		select {
		case log, ok := <-f.log:
			if !ok {
				return
			}
			handle(log) // was: return -- keep draining so App.Stop()'s own shutdown logs still get written

		case <-ticker.C:
			for pending.Size() > 0 {
				logLine, err := pending.Read()
				if err != nil {
					break
				}
				if _, err := file.WriteString(logLine); err != nil {
					pending.Add(logLine)
					break
				}
			}

		case <-f.ctx.Done():
			// Drain whatever's still sitting in f.log's buffer FIRST.
			// select gives "ctx.Done() fired" and "a log is still
			// queued in f.log" equal priority -- so without this drain,
			// Stop() firing while messages were still in flight could
			// pick this branch first and silently drop every log still
			// sitting in the channel's buffer, including, plausibly,
			// App.Stop()'s own "Renderer Stopped"/"App Stopped" lines
			// sent moments before Stop() closed done. Non-blocking: an
			// already-empty f.log just falls straight through to
			// default.
		drain:
			for {
				select {
				case log, ok := <-f.log:
					if !ok {
						break drain
					}
					handle(log)
				default:
					break drain
				}
			}

			for pending.Size() > 0 {
				logLine, err := pending.Read()
				if err != nil {
					break
				}
				file.WriteString(logLine)
			}
			return
		}
	}
}
