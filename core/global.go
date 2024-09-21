package core

import "os"

var (
	RP     Printer
	RIo    RadIo
	RExit  func(int)
	RReq   *Requester
	RClock Clock
)

type CmdInput struct {
	RIo    *RadIo
	RExit  *func(int)
	RReq   *Requester
	RClock Clock
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
