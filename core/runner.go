package core

import (
	"fmt"
	"io"
	"os"
	"strings"

	ra "github.com/amterp/ra"
	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts"

	"github.com/amterp/color"
)

type InvocationType int

const (
	NoScript        InvocationType = iota // help, version, no args
	ScriptFile                            // existing file
	StdinScript                           // "rad -"
	EmbeddedCommand                       // built-in commands
	Repl                                  // interactive REPL mode
)

type RadRunner struct {
	scriptData  *ScriptData
	globalFlags []RadArg
	scriptArgs  []RadArg
}

func NewRadRunner(runnerInput RunnerInput) *RadRunner {
	setGlobals(runnerInput)
	return &RadRunner{}
}

func (r *RadRunner) Run() error {
	// Phase 1: Detection & Setup
	invocationType, sourceCode, err := r.detectAndSetup(os.Args[1:])
	if err != nil {
		// Set up minimal printer for errors
		RP = NewPrinter(r, false, false, false, false)
		RP.ErrorExit(err.Error())
	}

	// Phase 2: Registration
	r.setupRootCommand()

	if invocationType != NoScript {
		err := r.registerScript(sourceCode)
		if err != nil {
			RP.ErrorExit(err.Error())
		}
	}

	// Let ra handle help flags properly through hooks - removed manual processing

	// Phase 3: Parse & Execute
	return r.parseAndExecute(invocationType)
}

// detectInvocationType analyzes the command line args to determine what type of invocation this is
func (r *RadRunner) detectInvocationType(args []string) (InvocationType, string, error) {
	// No args means global-only (will show help)
	if len(args) == 0 {
		return NoScript, "", nil
	}

	firstArg := args[0]

	// Handle stdin script ("rad -")
	if firstArg == "-" {
		if !RIo.StdIn.HasContent() {
			return NoScript, "", fmt.Errorf("Requested reading from stdin ('-'), but found no input")
		}
		source, err := io.ReadAll(RIo.StdIn)
		if err != nil {
			return NoScript, "", fmt.Errorf("Could not read from stdin: %v", err)
		}
		return StdinScript, string(source), nil
	}

	// Skip flags (anything starting with -)
	// todo don't think this correctly handles e.g. `rad -- myscript myarg` which should be equivalent to `rad myscript myarg`
	if strings.HasPrefix(firstArg, "-") {
		return NoScript, "", nil
	}

	// Check if it's an existing file
	if com.FileExists(firstArg) {
		source, err := readSource(firstArg)
		if err != nil {
			return NoScript, "", fmt.Errorf("Could not read script: %v", err)
		}
		return ScriptFile, source, nil
	}

	// Check if it's an embedded command
	cmdSource := GetEmbeddedCommandSrc(firstArg)
	if cmdSource != nil {
		AddInternalFuncs()
		return EmbeddedCommand, *cmdSource, nil
	}

	// Unknown file or command
	return NoScript, "", fmt.Errorf("Unknown file or command: %s", firstArg)
}

// detectAndSetup analyzes args and sets up basic state
func (r *RadRunner) detectAndSetup(args []string) (InvocationType, string, error) {
	invocationType, sourceCode, err := r.detectInvocationType(args)
	if err != nil {
		return NoScript, "", err
	}

	scriptPath := ""
	if invocationType == ScriptFile && len(args) > 0 {
		scriptPath = args[0]
	} else if invocationType == EmbeddedCommand && len(args) > 0 {
		// For embedded commands, use the command name as the script name
		scriptPath = args[0]
	} else if invocationType == StdinScript {
		// Remove the '-' from os.Args so Ra doesn't try to parse it as a flag
		os.Args = append([]string{os.Args[0]}, args[1:]...)
	}

	// Set up minimal printer for error handling during metadata extraction
	RP = NewPrinter(r, false, false, false, false)

	// Set up globals
	HasScript = invocationType != NoScript
	SetScriptPath(scriptPath)

	if HasScript {
		r.scriptData = ExtractMetadata(sourceCode)
	}

	return invocationType, sourceCode, nil
}

// setupRootCommand creates the root command and registers global flags
func (r *RadRunner) setupRootCommand() {
	// Use script name as the command name if we have a script, otherwise use the binary name
	cmdName := os.Args[0]
	if r.scriptData != nil && ScriptName != "" {
		cmdName = ScriptName
	}

	// In test mode, use a clean command name to match expected test output

	RRootCmd = ra.NewCmd(cmdName)

	RRootCmd.SetUsageHeaders(ra.UsageHeaders{
		Usage:         "Usage:",
		Commands:      "Commands:",
		Arguments:     "Script args:",
		GlobalOptions: "Global options:",
	})

	if r.scriptData == nil || !r.scriptData.DisableGlobalOpts {
		r.globalFlags = CreateAndRegisterGlobalFlags()
	}

	if r.scriptData != nil && r.scriptData.Description != nil {
		RRootCmd.SetDescription(*r.scriptData.Description)
	}

	RRootCmd.SetHelpEnabled(false) // Disable help initially, enable after script registration
	RRootCmd.SetAutoHelpOnNoArgs(true)

	// Set up PostParse hook to apply color settings after parsing but before output
	RRootCmd.SetParseHooks(&ra.ParseHooks{
		PostParse: func(cmd *ra.Cmd, err error) {
			// Apply color settings based on the parsed color flag
			switch FlagColor.Value {
			case COLOR_NEVER:
				color.NoColor = true
			case COLOR_ALWAYS:
				color.NoColor = false
			}
		},
	})
}

// registerScript registers the script as a subcommand with its flags
func (r *RadRunner) registerScript(sourceCode string) error {
	if r.scriptData == nil {
		return fmt.Errorf("Bug! Script data expected but not found")
	}

	// Validate args block if present
	if HasScript {
		radParser, err := rts.NewRadParser()
		if err != nil {
			return fmt.Errorf("Failed to load Rad parser: %v", err)
		}
		tree := radParser.Parse(sourceCode)
		_, hasArgsBlock := tree.FindArgBlock()
		if hasArgsBlock && r.scriptData.DisableArgsBlock {
			return fmt.Errorf("Macro '%s' disabled, but args block found.\n", MACRO_ENABLE_ARGS_BLOCK)
		}
	}

	r.scriptArgs = r.createAndRegisterScriptArgs()

	// Re-enable help after script registration, unless global options are disabled
	if !r.scriptData.DisableGlobalOpts {
		RRootCmd.SetHelpEnabled(true)
	}

	return nil
}

// parseAndExecute handles the final parsing and execution
func (r *RadRunner) parseAndExecute(invocationType InvocationType) error {

	// Do initial parse to get global flags working
	var argsToRead []string
	if invocationType == ScriptFile || invocationType == EmbeddedCommand {
		// Script invoked via rad like: rad ./test_simple.rad "World"
		// Skip the script path, parse just the script args: ["World"]
		if len(os.Args) > 2 {
			argsToRead = os.Args[2:]
		} else {
			argsToRead = []string{}
		}
	} else {
		// Other invocations (including direct script execution with shebang)
		// Parse everything after the script name: ["World"]
		argsToRead = os.Args[1:]
	}

	// Prepare parse options
	parseOpts := []ra.ParseOpt{ra.WithIgnoreUnknown(true), ra.WithVariadicUnknownFlags(true)}
	if FlagRadArgsDump.Value {
		parseOpts = append(parseOpts, ra.WithDump(true))
	}

	RRootCmd.ParseOrExit(argsToRead, parseOpts...)

	// Set up printer with global flags
	RP = NewPrinter(r, FlagShell.Value, FlagQuiet.Value, FlagDebug.Value, FlagRadDebug.Value)

	// Handle mock responses
	mockResponse := FlagMockResponse.Value
	if !com.IsBlank(mockResponse) {
		split := strings.Split(mockResponse, ":")
		pattern := split[0]
		path := split[1]
		RReq.AddMockedResponse(pattern, path)
		RP.RadDebugf(fmt.Sprintf("Mock response added: %q -> %q", pattern, path))
	}

	if FlagVersion.Value {
		RP.Printf(fmt.Sprintf("rad %s\n", Version))
		RExit(0)
	}

	// Handle REPL mode (but not if help is requested)
	if FlagRepl.Value && !FlagHelp.Value {
		return r.runRepl()
	}

	// Handle global-only invocations
	if invocationType == NoScript {
		unknownArgs := RRootCmd.GetUnknownArgs()
		if len(unknownArgs) > 0 {
			RP.UsageErrorExit(fmt.Sprintf("Unknown arguments: %v\n", unknownArgs))
		}
		// For global-only invocations without args, show help and exit
		// Ra will handle the help generation properly
		r.printScriptlessUsage(false)
		RExit(0)
	}

	// Handle debug flags for script output
	shouldExit := false
	if FlagSrc.Value {
		shouldExit = true
		if FlagVersion.Value {
			RP.Printf("\n")
		}
		if !com.IsBlank(ScriptPath) && com.IsTty {
			RP.RadInfo(com.YellowS("%s:\n", ScriptPath))
		}
		RP.Printf(r.scriptData.Src + "\n")
	}

	if FlagRadTree.Value {
		shouldExit = true
		if FlagSrc.Value {
			RP.Printf("\n")
		}
		RP.Printf(r.scriptData.Tree.Dump())
	}

	if shouldExit {
		RExit(0)
	}

	// Final parse with correct ignore settings
	// Ignore unknown args when args block is disabled (so they can be accessed via get_args())
	ignoreUnknown := r.scriptData.DisableArgsBlock

	// Prepare final parse options
	finalParseOpts := []ra.ParseOpt{ra.WithIgnoreUnknown(ignoreUnknown), ra.WithVariadicUnknownFlags(true)}
	if FlagRadArgsDump.Value {
		finalParseOpts = append(finalParseOpts, ra.WithDump(true))
	}

	RRootCmd.ParseOrExit(argsToRead, finalParseOpts...)

	// Execute the script
	if r.scriptData == nil {
		return fmt.Errorf("Bug! Script expected by this point, but found none")
	}

	interpreter := NewInterpreter(r.scriptData)
	interpreter.InitBuiltIns()
	interpreter.InitArgs(r.scriptArgs)
	interpreter.RegisterWithExit()
	interpreter.Run()

	if FlagShell.Value {
		interpreter.env.PrintShellExports()
	}

	RExit(0) // explicit exit to trigger deferred statements
	return nil
}

func (r *RadRunner) createAndRegisterScriptArgs() []RadArg {
	if r.scriptData == nil {
		return nil
	}

	// Register script flags directly on the root command (no subcommand)
	// This makes the script appear as the main command
	flags := make([]RadArg, 0, len(r.scriptData.Args))
	for _, arg := range r.scriptData.Args {
		flag := CreateFlag(arg)
		flag.Register(RRootCmd, false)
		flags = append(flags, flag)
	}

	return flags
}

func readSource(scriptPath string) (string, error) {
	source, err := os.ReadFile(scriptPath)
	return string(source), err
}

// runRepl starts the interactive REPL mode
func (r *RadRunner) runRepl() error {
	return RunRepl()
}
