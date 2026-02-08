package core

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/amterp/color"
	"github.com/amterp/ra"
)

var (
	RRootCmd                 *ra.Cmd
	RConfig                  *RadConfig
	RP                       Printer
	RIo                      RadIo
	RExit                    *RadExitHandler
	RReq                     *Requester
	RClock                   Clock
	RSleep                   func(duration time.Duration)
	RShell                   ShellExecutor
	RNG                      *rand.Rand
	HasScript                bool
	ScriptPath               string
	ScriptDir                string
	ScriptName               string
	IsTest                   bool
	AlreadyExportedShellVars bool

	StartEpochMillis int64
)

type RunnerInput struct {
	RIo     *RadIo
	RExit   *func(int)
	RReq    *Requester
	RClock  Clock
	RSleep  *func(duration time.Duration)
	RShell  *func(invocation ShellInvocation) (string, string, int)
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

	if IsTest {
		ScriptName = "TestCase"
	}
}

// primarily for tests
func ResetGlobals() {
	RConfig = nil
	RP = nil
	RIo = RadIo{}
	RExit = nil
	RReq = nil
	RClock = nil
	RSleep = nil
	RShell = nil
	RNG = nil
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
	FlagCstTree = BoolRadArg{}
	FlagAstTree = BoolRadArg{}
	FlagRadArgsDump = BoolRadArg{}
	FlagMockResponse = StringRadArg{}
	FlagRepl = BoolRadArg{}

	StartEpochMillis = 0

	color.NoColor = false
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
		RExit = NewExitHandler(os.Exit)
	} else {
		RExit = NewExitHandler(*runnerInput.RExit)
	}

	ra.SetStderrWriter(RIo.StdErr)
	ra.SetStdoutWriter(RIo.StdOut)
	ra.SetExitFunc(RExit.Exit)

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
	StartEpochMillis = RClock.Now().UnixMilli()
	if runnerInput.RSleep == nil {
		RSleep = func(duration time.Duration) {
			time.Sleep(duration)
		}
	} else {
		RSleep = *runnerInput.RSleep
	}

	if runnerInput.RShell == nil {
		RShell = realShellExecutor
	} else {
		RShell = *runnerInput.RShell
	}

	// Initialize RNG with clock-based seed (respects RClock abstraction)
	RNG = rand.New(rand.NewSource(RClock.Now().UnixNano()))

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
	fmt.Fprintf(
		RIo.StdErr,
		"Unable to determine home directory for rad. Please define a valid path '%s' as an environment variable.\n",
		ENV_RAD_HOME,
	)
	RExit.Exit(1)
}

func init() {
	// just in case RNG is invoked before setGlobals
	RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
}
