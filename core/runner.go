package core

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/amterp/color"
	"github.com/amterp/ra"
	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts"
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
	scriptData     *ScriptData
	globalFlags    []RadArg
	scriptArgs     []RadArg
	cmdInvocations []cmdInvocation
}

type cmdInvocation struct {
	cmd     *ScriptCommand
	usedPtr *bool
	args    []RadArg // Command-specific arguments
}

func NewRadRunner(runnerInput RunnerInput) *RadRunner {
	setGlobals(runnerInput)
	return &RadRunner{}
}

func (r *RadRunner) Run() error {
	RConfig = LoadRadConfig()

	// Phase 1: Detection & Setup
	invocationType, sourceCode, err := r.detectAndSetup(os.Args[1:])
	if err != nil {
		// Set up minimal printer for errors
		RP = NewPrinter(r, false, false, false, false)
		RP.ErrorExit(err.Error())
	}

	if RConfig.InvocationLogging.Enabled && invocationType == ScriptFile {
		RegisterInvocationLogging()
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

	// Register commands if any exist
	if len(r.scriptData.Commands) > 0 {
		err := r.registerCommands()
		if err != nil {
			return err
		}
	}

	// Re-enable help after script registration, unless global options are disabled
	if !r.scriptData.DisableGlobalOpts {
		RRootCmd.SetHelpEnabled(true)
	}

	return nil
}

// parseAndExecute handles the final parsing and execution
func (r *RadRunner) parseAndExecute(invocationType InvocationType) error {

	// Double-parse pattern: parse once to get global flags for intermediate logic (printer,
	// version, etc.), reset state, then parse again with correct ignoreUnknown setting.
	// See commit message for detailed rationale.

	var argsToRead []string
	if invocationType == ScriptFile || invocationType == EmbeddedCommand {
		if len(os.Args) > 2 {
			argsToRead = os.Args[2:]
		} else {
			argsToRead = []string{}
		}
	} else {
		argsToRead = os.Args[1:]
	}

	// First parse: ignoreUnknown=true since script args aren't registered yet
	parseOpts := []ra.ParseOpt{ra.WithIgnoreUnknown(true), ra.WithVariadicUnknownFlags(true)}
	if FlagRadArgsDump.Value {
		parseOpts = append(parseOpts, ra.WithDump(true))
	}

	RRootCmd.ParseOrExit(argsToRead, parseOpts...)

	// Set up printer with global flags from first parse
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
		printVersion()
		RExit.Exit(0)
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
		RExit.Exit(0)
	}

	// Handle inspection flags for script output
	shouldExit := false
	if FlagSrc.Value {
		shouldExit = true
		printSource(r.scriptData.Src, FlagVersion.Value)
	}

	if FlagSrcTree.Value {
		shouldExit = true
		printTree(r.scriptData.Tree, FlagSrc.Value)
	}

	if shouldExit {
		RExit.Exit(0)
	}

	// Cache dump flag value before reset (needed for second parse options)
	dumpFlag := FlagRadArgsDump.Value

	// Reset all parse state (flag values, configured flags, unknown args, etc.)
	RRootCmd.ResetParseState()

	// Second parse with correct ignoreUnknown setting based on script metadata
	ignoreUnknown := false
	if r.scriptData != nil {
		ignoreUnknown = r.scriptData.DisableArgsBlock
	}
	finalParseOpts := []ra.ParseOpt{ra.WithIgnoreUnknown(ignoreUnknown), ra.WithVariadicUnknownFlags(true)}
	if dumpFlag {
		finalParseOpts = append(finalParseOpts, ra.WithDump(true))
	}

	RRootCmd.ParseOrExit(argsToRead, finalParseOpts...)

	// Determine which command was invoked (if any)
	var invokedCommand *ScriptCommand
	var commandArgs []RadArg
	for _, inv := range r.cmdInvocations {
		if *inv.usedPtr {
			invokedCommand = inv.cmd
			commandArgs = inv.args
			break
		}
	}

	// Check if command is required but none was invoked
	// (Commands exist but none invoked and not help/version/inspection flags)
	if len(r.cmdInvocations) > 0 && invokedCommand == nil {
		if !FlagHelp.Value && !FlagVersion.Value && !FlagSrc.Value && !FlagSrcTree.Value {
			RP.UsageErrorExit("Must specify a command")
		}
	}

	// Execute the script
	if r.scriptData == nil {
		return fmt.Errorf("Bug! Script expected by this point, but found none")
	}

	interpreter := NewInterpreter(InterpreterInput{
		Src:            r.scriptData.Src,
		Tree:           r.scriptData.Tree,
		ScriptName:     r.scriptData.ScriptName,
		InvokedCommand: invokedCommand,
	})
	interpreter.InitBuiltIns()
	interpreter.InitArgs(r.scriptArgs)
	// Initialize command-specific args if a command was invoked
	if invokedCommand != nil && len(commandArgs) > 0 {
		interpreter.InitArgs(commandArgs)
	}
	interpreter.RegisterWithExit()
	interpreter.Run()

	if FlagShell.Value {
		interpreter.env.PrintShellExports()
	}

	RExit.Exit(0) // explicit exit to trigger exit handlers (e.g. deferred statements)
	return nil
}

func (r *RadRunner) createAndRegisterScriptArgs() []RadArg {
	if r.scriptData == nil {
		return nil
	}

	hasCommands := len(r.scriptData.Commands) > 0

	flags := make([]RadArg, 0, len(r.scriptData.Args))
	for _, arg := range r.scriptData.Args {
		flag := CreateFlag(arg)
		flags = append(flags, flag)
	}

	// When NO commands: register script args on root as positional+flag
	// When commands exist: DON'T register on root - will be registered on each subcommand
	if !hasCommands {
		for _, flag := range flags {
			flag.Register(RRootCmd, AsScriptArg)
		}
	}

	return flags
}

func (r *RadRunner) registerCommands() error {
	r.cmdInvocations = make([]cmdInvocation, 0, len(r.scriptData.Commands))

	for _, scriptCmd := range r.scriptData.Commands {
		// Create Ra subcommand
		raSubCmd := ra.NewCmd(scriptCmd.Name)
		if scriptCmd.Description != nil {
			raSubCmd.SetDescription(*scriptCmd.Description)
		}

		// Configure subcommand usage headers
		raSubCmd.SetUsageHeaders(ra.UsageHeaders{
			Usage:         "Usage:",
			Arguments:     "Command args:",
			GlobalOptions: "Global options:",
		})

		// Enable help for the subcommand
		raSubCmd.SetHelpEnabled(true)

		// Register script args on this subcommand as flag-only
		// (Script args are shared across all commands but only accept flag syntax)
		for _, scriptArg := range r.scriptArgs {
			scriptArg.Register(raSubCmd, AsScriptFlagOnly)
		}

		// Register command-specific args on the subcommand as positional+flag
		cmdArgs := make([]RadArg, 0, len(scriptCmd.Args))
		for _, arg := range scriptCmd.Args {
			flag := CreateFlag(arg)
			flag.Register(raSubCmd, AsCommandArg)
			cmdArgs = append(cmdArgs, flag)
		}

		// Register the subcommand with the root command
		usedPtr, err := RRootCmd.RegisterCmd(raSubCmd)
		if err != nil {
			return fmt.Errorf("failed to register command '%s': %w", scriptCmd.Name, err)
		}

		// Store the invocation tracking
		r.cmdInvocations = append(r.cmdInvocations, cmdInvocation{
			cmd:     scriptCmd,
			usedPtr: usedPtr,
			args:    cmdArgs,
		})
	}

	return nil
}

func readSource(scriptPath string) (string, error) {
	source, err := os.ReadFile(scriptPath)
	return string(source), err
}

// runRepl starts the interactive REPL mode
func (r *RadRunner) runRepl() error {
	return RunRepl()
}

// Helper functions for inspection flags (--version, --src, --src-tree)
// These are used both when handling valid scripts and when checking flags before error exit

func printVersion() {
	RP.Printf(fmt.Sprintf("rad %s\n", Version))
}

func printSource(src string, prependNewline bool) {
	if prependNewline {
		RP.Printf("\n")
	}
	if !com.IsBlank(ScriptPath) && com.IsTty {
		RP.RadStderrf(com.YellowS("%s:\n", ScriptPath))
	}
	RP.Print(src + "\n")
}

func printTree(tree *rts.RadTree, prependNewline bool) {
	if prependNewline {
		RP.Printf("\n")
	}
	RP.Print(tree.Dump())
}

// handleGlobalInspectionFlagsOnInvalidSyntax checks os.Args for inspection flags and handles them.
// This is called from validateSyntax() when the script has errors, before showing the error.
// Returns true if a flag was handled (and the function exited), false otherwise.
func handleGlobalInspectionFlagsOnInvalidSyntax(src string, tree *rts.RadTree) {
	hasVersion := false
	hasSrc := false
	hasSrcTree := false

	for _, arg := range os.Args { // todo don't love the hardcoded string lookups
		if arg == "--version" || arg == "-v" {
			hasVersion = true
		}
		if arg == "--src" {
			hasSrc = true
		}
		if arg == "--src-tree" {
			hasSrcTree = true
		}
	}

	if hasVersion {
		printVersion()
		RExit.Exit(0)
	}

	if hasSrc {
		printSource(src, hasVersion)
		RExit.Exit(0)
	}

	if hasSrcTree {
		printTree(tree, hasSrc)
		RExit.Exit(0)
	}
}
