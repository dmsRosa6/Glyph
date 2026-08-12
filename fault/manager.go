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

type FaultManager struct {
	appSignal chan<- core.AppSignal
	log       chan core.AppLog
	logLevel  core.Severity
	cancel    context.CancelFunc
	ctx       context.Context
	done      chan struct{}
}

const logFileName string = "log_%s.txt"
const basePath string = "logs"

func NewFaultManager(logLevel core.Severity, signals chan core.AppSignal) (*FaultManager, error) {
	l := make(chan core.AppLog, 100)
	ctx, cancel := context.WithCancel(context.Background())

	return &FaultManager{
		appSignal: signals,
		log:       l,
		logLevel:  logLevel,
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}), // NEW
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
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err := os.MkdirAll(basePath, 0755)

		if err != nil {
			panic(err)
		}
	}

	ts := time.Now().Format("20060102_150405")

	resolvedFileName := filepath.Join(basePath,
		fmt.Sprintf(logFileName, ts))

	file, err := os.OpenFile(resolvedFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			panic("error closing the file.")
		}
	}()

	// TODO: Make retry buffer capacity configurable.
	pending := datastructs.NewRingBuffer(100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case log, ok := <-f.log:
			if !ok {
				return
			}

			if log.Severity() < f.logLevel {
				continue
			}

			ts := time.Now().Format("2006-01-02 15:04:05")
			logLine := fmt.Sprintf("%s %s\n", ts, log.Reason())

			// Big fuck up this is the case where we just want to get over it we will not wait ticks to write to log
			if log.Severity() == core.Fatal {
				if _, err := file.WriteString(logLine); err != nil {
					fmt.Printf("could not write fatal log: %v\n%s", err, logLine)
				}
				f.appSignal <- core.SIGTERM
				continue // was: return -- keep draining so App.Stop()'s own shutdown logs still get written
			}

			if pending.Size() > 0 {
				pending.Add(logLine)
				continue
			}

			if _, err := file.WriteString(logLine); err != nil {
				pending.Add(logLine)
			}

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
