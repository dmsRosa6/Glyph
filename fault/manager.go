package fault

import (
	"fmt"
	"os"
	"time"

	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/datastructs"
	"github.com/dmsRosa6/glyph/framework"
)

type FaultManager struct {
	appSignal chan<- app.AppSignal
	log       <-chan framework.AppLog
	logLevel  framework.Severity
}

func NewFaultManager(logLevel framework.Severity) (*FaultManager, error) {
	s := make(chan app.AppSignal)
	l := make(chan framework.AppLog, 100)

	return &FaultManager{
		appSignal: s,
		log:       l,
		logLevel:  logLevel,
	}, nil
}

func (f *FaultManager) Start() {
	go f.run()
}

func (f *FaultManager) run() {
	fileName := "log.txt"

	err := os.Remove(fileName)
	if err != nil && !os.IsNotExist(err) {
		panic("could not start fault manager.")
	}

	file, err := os.Create(fileName)
	if err != nil {
		panic("could not start fault manager.")
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			panic("error closing the file.")
		}
	}()

	//propably should be dynamic
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
			if log.Severity() == framework.Fatal {
				if _, err := file.WriteString(logLine); err != nil {
					fmt.Printf("could not write fatal log: %v\n%s", err, logLine)
				}

				f.appSignal <- app.SIGTERM
				return
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
		}
	}
}
