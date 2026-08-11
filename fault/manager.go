package fault

import (
	"fmt"
	"os"
	"time"

	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/framework"
)

type FaultManager struct {
	appSignal chan<- app.AppSignal
	log       <-chan framework.AppLog
}

func NewFaultManager() (*FaultManager, error) {
	s := make(chan app.AppSignal)
	l := make(chan framework.AppLog, 100)

	return &FaultManager{appSignal: s, log: l}, nil
}

func (f *FaultManager) Start() {
	go f.run()
}

func (f *FaultManager) run() {
	fileName := "log.txt"

	file, err := os.Create(fileName)
	if err != nil {
		panic("could not start fault manager.")
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			panic("error closing the file.")
		}
	}()

	for log := range f.log {
		rsn := log.Reason()

		ts := time.Now().Format("2006-01-02 15:04:05")
	
		logLine = fmt.Sprintf("%s %s\n", ts, rsn)
		file.
	}

}
