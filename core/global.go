package core

import (
	"os"
	"path/filepath"
	"time"
)

var (
	RP         Printer
	RIo        RadIo
	RExit      func(int)
	RReq       *Requester
	RClock     Clock
	RSleep     func(duration time.Duration)
	ScriptPath string
	ScriptDir  string
	ScriptName string
)

type RunnerInput struct {
	RIo    *RadIo
	RExit  *func(int)
	RReq   *Requester
	RClock Clock
	RSleep *func(duration time.Duration)
}

func SetScriptPath(path string) {
	ScriptPath = path
	ScriptDir = filepath.Dir(path)
	ScriptName = filepath.Base(path)
}

// primarily for tests
func ResetGlobals() {
	RP = nil
	RIo = RadIo{}
	RExit = nil
	RReq = nil
	RClock = nil
	RSleep = nil
}

func setGlobals(runnerInput RunnerInput) {
	if runnerInput.RIo == nil {
		RIo = RadIo{
			StdIn:  os.Stdin,
			StdOut: os.Stdout,
			StdErr: os.Stderr,
		}
	} else {
		RIo = *runnerInput.RIo
	}

	if runnerInput.RExit == nil {
		RExit = os.Exit
	} else {
		RExit = *runnerInput.RExit
	}

	if runnerInput.RReq == nil {
		RReq = NewRequester()
	} else {
		RReq = runnerInput.RReq
	}

	if runnerInput.RClock == nil {
		RClock = NewRealClock()
	} else {
		RClock = runnerInput.RClock
	}
	if runnerInput.RSleep == nil {
		RSleep = func(duration time.Duration) {
			time.Sleep(duration)
		}
	} else {
		RSleep = *runnerInput.RSleep
	}
}
