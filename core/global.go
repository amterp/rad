package core

import (
	"context"
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
	RForceExit               func(int) // hard exit that skips defers; for double-Ctrl+C
	RReq                     *Requester
	RClock                   Clock
	RSleep                   func(ctx context.Context, duration time.Duration)
	RShell                   ShellExecutor
	RConfirm                 func(title string, prompt string) (bool, error)
	RSignal                  SignalSource
	RInteractive             InteractiveDriver
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
	RIo   *RadIo
	RExit *func(int)
	// RForceExit is the raw hard-exit (no defers/callbacks), distinct from
	// RExit which wraps into a RadExitHandler. Tests inject a record-only
	// variant; see global RForceExit.
	RForceExit *func(int)
	RReq       *Requester
	RClock     Clock
	RSleep     *func(ctx context.Context, duration time.Duration)
	RShell     *func(ctx context.Context, invocation ShellInvocation) (string, string, int)
	RConfirm   *func(title string, prompt string) (bool, error)
	RSignal    SignalSource
	RadHome    *string
	RTermWidth *int
	// RInteractive overrides the interactive-prompt driver. nil uses the real
	// terminal; tests inject a scripted driver. See InteractiveDriver.
	RInteractive InteractiveDriver
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
	RForceExit = nil
	RReq = nil
	RClock = nil
	RSleep = nil
	RShell = nil
	RConfirm = nil
	RSignal = nil
	RInteractive = nil
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

	// RForceExit is the raw, defer-skipping exit used by the double-Ctrl+C
	// force path. It runs on the signal-dispatch goroutine, so the test seam
	// records the code and returns rather than panicking to unwind (which only
	// works on the main goroutine).
	if runnerInput.RForceExit == nil {
		RForceExit = os.Exit
	} else {
		RForceExit = *runnerInput.RForceExit
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
		RSleep = func(ctx context.Context, duration time.Duration) {
			// select so a signal handler can interrupt sleep promptly.
			// If ctx is already canceled when we enter, we return immediately
			// without sleeping - which is what callers should expect after a
			// signal has fired.
			select {
			case <-time.After(duration):
			case <-ctx.Done():
			}
		}
	} else {
		RSleep = *runnerInput.RSleep
	}

	if runnerInput.RShell == nil {
		RShell = realShellExecutor
	} else {
		RShell = *runnerInput.RShell
	}

	if runnerInput.RConfirm == nil {
		RConfirm = InputConfirm
	} else {
		RConfirm = *runnerInput.RConfirm
	}

	if runnerInput.RSignal == nil {
		RSignal = realSignalSource{}
	} else {
		RSignal = runnerInput.RSignal
	}

	if runnerInput.RInteractive == nil {
		RInteractive = terminalDriver{}
	} else {
		RInteractive = runnerInput.RInteractive
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
