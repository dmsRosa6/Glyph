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

type Config struct {
	LogLevel       core.Severity
	LogDir         string
	DisableFileLog bool
}

type FaultManager struct {
	appSignal chan<- core.AppSignal
	log       chan core.AppLog
	logLevel  core.Severity
	cancel    context.CancelFunc
	ctx       context.Context
	done      chan struct{}
	file      *os.File
}

const logFileName string = "log_%s.txt"
const basePath string = "logs"

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
			return
		}
		if cerr := file.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "fault: error closing log file: %v\n", cerr)
		}
	}()

	pending := datastructs.NewRingBuffer(100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	write := func(logLine string) {
		if file == nil {
			return
		}
		if pending.Size() > 0 {
			pending.Add(logLine)
			return
		}
		if _, err := file.WriteString(logLine); err != nil {
			pending.Add(logLine)
		}
	}

	handle := func(log core.AppLog) {
		if log.Severity() < f.logLevel {
			return
		}

		ts := time.Now().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("%s %s\n", ts, log.Reason())

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
			handle(log)

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
