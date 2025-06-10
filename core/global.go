package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/pflag"
)

var (
	RFlagSet                 *pflag.FlagSet
	RP                       Printer
	RIo                      RadIo
	RExit                    func(int)
	RReq                     *Requester
	RClock                   Clock
	RSleep                   func(duration time.Duration)
	HasScript                bool
	ScriptPath               string
	ScriptDir                string
	ScriptName               string
	IsTest                   bool
	AlreadyExportedShellVars bool
)

type RunnerInput struct {
	RIo     *RadIo
	RExit   *func(int)
	RReq    *Requester
	RClock  Clock
	RSleep  *func(duration time.Duration)
	RadHome *string
}

func SetScriptPath(path string) {
	ScriptPath = path
	ScriptDir = filepath.Dir(path)
	if path == "" {
		ScriptName = ""
	} else {
		ScriptName = filepath.Base(path)
	}
}

// primarily for tests
func ResetGlobals() {
	RFlagSet = nil
	FlagsUsedInScript = []string{}
	RP = nil
	RIo = RadIo{}
	RExit = nil
	RReq = nil
	RClock = nil
	RSleep = nil
	HasScript = false
	ScriptPath = ""
	ScriptDir = ""
	ScriptName = ""
	IsTest = false
	AlreadyExportedShellVars = false

	FlagHelp = BoolRadArg{}
	FlagDebug = BoolRadArg{}
	FlagRadDebug = BoolRadArg{}
	FlagColor = StringRadArg{}
	FlagQuiet = BoolRadArg{}
	FlagShell = BoolRadArg{}
	FlagVersion = BoolRadArg{}
	FlagConfirmShellCommands = BoolRadArg{}
	FlagSrc = BoolRadArg{}
	FlagRadTree = BoolRadArg{}
	FlagMockResponse = MockResponseRadArg{}
}

func setGlobals(runnerInput RunnerInput) {
	if runnerInput.RIo == nil {
		RIo = RadIo{
			StdIn:  NewFileReader(os.Stdin),
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

	var home string
	if runnerInput.RadHome == nil {
		home = os.Getenv(ENV_RAD_HOME)

		if home == "" {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				home = filepath.Join(homeDir, ".rad")
			}
		}

		if home == "" {
			failedToDetermineRadHomeDir()
		}
	} else {
		home = *runnerInput.RadHome
	}

	home, err := filepath.Abs(home)
	if err != nil {
		failedToDetermineRadHomeDir()
	}
	RadHomeInst = NewRadHome(home)
}

func failedToDetermineRadHomeDir() {
	fmt.Fprintf(RIo.StdErr, "Unable to determine home directory for rad. Please define a valid path '%s' as an environment variable.\n", ENV_RAD_HOME)
	RExit(1)
}
