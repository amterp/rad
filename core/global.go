package core

import (
	"os"
	"path/filepath"
)

var (
	RP         Printer
	RIo        RadIo
	RExit      func(int)
	RReq       *Requester
	RClock     Clock
	ScriptPath string
	ScriptDir  string
	ScriptName string
)

type CmdInput struct {
	RIo    *RadIo
	RExit  *func(int)
	RReq   *Requester
	RClock Clock
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
}

func setGlobals(cmdInput CmdInput) {
	if cmdInput.RIo == nil {
		RIo = RadIo{
			StdIn:  os.Stdin,
			StdOut: os.Stdout,
			StdErr: os.Stderr,
		}
	} else {
		RIo = *cmdInput.RIo
	}

	if cmdInput.RExit == nil {
		RExit = os.Exit
	} else {
		RExit = *cmdInput.RExit
	}

	if cmdInput.RReq == nil {
		RReq = NewRequester()
	} else {
		RReq = cmdInput.RReq
	}

	if cmdInput.RClock == nil {
		RClock = NewRealClock()
	} else {
		RClock = cmdInput.RClock

	}
}
