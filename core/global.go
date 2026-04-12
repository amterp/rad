package core

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/amterp/color"
	"github.com/amterp/ra"
	"golang.org/x/term"
)

// defaultTermWidth is returned when the terminal width cannot be determined
// (e.g. when stdout is not a TTY, such as in piped contexts). Effectively "very
// wide" so that no truncation occurs.
const defaultTermWidth = 9999

// GetTermWidth returns the terminal width to use for rendering. This is the
// canonical way to obtain terminal width in production code - callers do not
// need to concern themselves with testing overrides. Tests can inject a fixed
// width via RunnerInput.RTermWidth.
func GetTermWidth() int {
	if RTermWidth != nil {
		return *RTermWidth
	}
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		RP.RadDebugf(fmt.Sprintf("Error getting terminal width, defaulting to %d: %v\n", defaultTermWidth, err))
		return defaultTermWidth
	}
	return width
}

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
	RTermWidth               *int // nil = use real terminal width; non-nil overrides for tests

	StartEpochMillis int64
)

type RunnerInput struct {
	RIo        *RadIo
	RExit      *func(int)
	RReq       *Requester
	RClock     Clock
	RSleep     *func(duration time.Duration)
	RShell     *func(invocation ShellInvocation) (string, string, int)
	RadHome    *string
	RTermWidth *int
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
	RTermWidth = nil

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
	FlagTlsInsecure = BoolRadArg{}

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

	RTermWidth = runnerInput.RTermWidth

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
