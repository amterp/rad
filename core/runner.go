package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/amterp/color"
	"github.com/amterp/ra"
	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

type InvocationType int

const (
	NoScript        InvocationType = iota // help, version, no args
	ScriptFile                            // existing file
	StdinScript                           // "rad -"
	EmbeddedCommand                       // built-in commands
)

type RadRunner struct {
	scriptData     *ScriptData
	invocationType InvocationType
	globalFlags    []RadArg
	scriptArgs     []RadArg
	cmdInvocations []cmdInvocation
	// usageCmd is the command whose usage a help or usage-error should render.
	// Nil means the script root. It matters once commands nest: failing on
	// `tool remote` and then printing the root's command list answers a
	// question the user didn't ask.
	usageCmd *ra.Cmd
}

type cmdInvocation struct {
	cmd     *ScriptCommand
	usedPtr *bool
	args    []RadArg // Command-specific arguments
	raCmd   *ra.Cmd
}

func NewRadRunner(runnerInput RunnerInput) *RadRunner {
	setGlobals(runnerInput)
	return &RadRunner{}
}

func (r *RadRunner) Run() error {
	RConfig = LoadRadConfig()

	// Handle `rad completion <shell> [scripts...]` early, before normal detection.
	// A Rad script named "completion" in CWD takes precedence (consistent with
	// embedded command behavior), but a non-Rad file named "completion" should not
	// block the completion command.
	if len(os.Args) > 1 && os.Args[1] == "completion" && !(com.IsRegularFile("completion") && isRadScript("completion")) {
		return r.handleCompletionCommand(os.Args[2:])
	}

	// Handle `rad repl` early too - the REPL is implemented in Go, not as an
	// embedded script. A regular file named "repl" in CWD takes precedence,
	// consistent with embedded command behavior.
	if len(os.Args) > 1 && os.Args[1] == "repl" && !com.IsRegularFile("repl") {
		return r.handleReplCommand(os.Args[2:])
	}

	// Phase 1: Detection & Setup
	invocationType, err := r.detectAndSetup(os.Args[1:])
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
		err := r.registerScript()
		if err != nil {
			RP.ErrorExit(err.Error())
		}
	} else if isCompletionRequest() {
		// Only parse and register embedded commands when actually serving a
		// completion request. This avoids the tree-sitter overhead on normal
		// invocations like `rad --help` or `rad --version`.
		r.registerEmbeddedCommandsForCompletion()
	}

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
		return StdinScript, NormalizeLineEndings(string(source)), nil
	}

	// Skip flags (anything starting with -)
	// todo don't think this correctly handles e.g. `rad -- myscript myarg` which should be equivalent to `rad myscript myarg`
	if strings.HasPrefix(firstArg, "-") {
		return NoScript, "", nil
	}

	// Check if it's an existing regular file (not a device, pipe, socket, etc.)
	if com.IsRegularFile(firstArg) {
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

	// Shell completion request for the rad CLI itself (e.g., "rad __complete docs ...").
	// Script completion (e.g., "rad script.rad __complete ...") is handled above as
	// ScriptFile - the __complete appears in argsToRead and is caught in parseAndExecute.
	if firstArg == "__complete" {
		return NoScript, "", nil
	}

	// Unknown file or command
	return NoScript, "", fmt.Errorf("Unknown file or command: %s", firstArg)
}

// isCompletionRequest returns true if the current invocation is a shell completion
// request. Checks both positions where __complete can appear:
//   - "rad __complete ..."         (CLI completion, __complete at position 1)
//   - "rad script.rad __complete ..." (script completion, __complete at position 2)
func isCompletionRequest() bool {
	return (len(os.Args) > 1 && os.Args[1] == "__complete") ||
		(len(os.Args) > 2 && os.Args[2] == "__complete")
}

// detectAndSetup analyzes args and sets up basic state
func (r *RadRunner) detectAndSetup(args []string) (InvocationType, error) {
	invocationType, sourceCode, err := r.detectInvocationType(args)
	if err != nil {
		return NoScript, err
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
	r.invocationType = invocationType
	SetScriptPath(scriptPath)

	if HasScript {
		r.scriptData = ExtractMetadata(sourceCode)
	}

	return invocationType, nil
}

// setupRootCommand creates the root command and registers global flags
func (r *RadRunner) setupRootCommand() {
	// Use script name as the command name if we have a script, otherwise use the binary name
	cmdName := os.Args[0]
	if r.invocationType == EmbeddedCommand {
		// Embedded commands present as rad subcommands ("rad fmt"), not as
		// standalone scripts. ScriptPath rather than ScriptName: tests override
		// ScriptName, but the command's name is fixed.
		cmdName = "rad " + ScriptPath
	} else if r.scriptData != nil && ScriptName != "" {
		cmdName = ScriptName
	}

	// In test mode, use a clean command name to match expected test output

	RRootCmd = ra.NewCmd(cmdName)

	argsHeader := "Script args:"
	if r.invocationType == EmbeddedCommand {
		// "Script args:" would leak that built-in commands are scripts under the hood.
		argsHeader = "Arguments:"
	}
	RRootCmd.SetUsageHeaders(ra.UsageHeaders{
		Usage:                 "Usage:",
		Commands:              "Commands:",
		Arguments:             argsHeader,
		GlobalOptions:         "Global options:",
		SubcommandPlaceholder: "command",
	})

	if r.scriptData == nil || !r.scriptData.DisableGlobalOpts {
		r.globalFlags = CreateAndRegisterGlobalFlags(r.invocationType)
	}

	if r.scriptData != nil && r.scriptData.Description != nil {
		RRootCmd.SetDescription(*r.scriptData.Description)
	}

	RRootCmd.SetHelpEnabled(false) // Disable help initially, enable after script registration
	RRootCmd.SetAutoHelpOnNoArgs(true)

	// Under --shell, stdout is reserved for eval-able output, so usage must go to
	// stderr, followed by an `exit 0` on stdout so an eval'ing wrapper stops after
	// showing help. Routing through RunUsage gives us that (via printHelpFromBuffer)
	// instead of Ra's unconditional print-to-stdout.
	RRootCmd.SetCustomUsage(func(isLongHelp bool) {
		r.RunUsage(!isLongHelp, false)
		emitShellExit(0)
	})

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

			// Parse failures are printed and exited inside Ra, bypassing our
			// printer, so shell mode emits its eval-able exit here. Help, dump,
			// and completion are non-failure exits handled by their own paths.
			if err != nil &&
				!errors.Is(err, ra.HelpInvokedErr) &&
				!errors.Is(err, ra.DumpInvokedErr) &&
				!errors.Is(err, ra.CompletionInvokedErr) {
				emitShellExit(1)
			}
		},
		ErrorHint: func(cmd *ra.Cmd, err error) string {
			return r.parseErrorHint(err)
		},
	})
}

// parseErrorHint attaches rad-specific guidance to ra parse errors, printed
// between the error and the usage string.
//
// Covers one case today: excess positionals while the script declares a
// non-variadic list-typed arg. `targets str[]` fills a single positional slot
// (lists are built by repeating the flag), so `script a b c` errors in a way
// that looks like a parser bug - and the fix, `*targets str`, is invisible
// from the error alone (issue #153).
func (r *RadRunner) parseErrorHint(err error) string {
	var tooMany *ra.TooManyPositionalArgsError
	if !errors.As(err, &tooMany) {
		return ""
	}
	for _, arg := range r.scriptArgs {
		if arg.IsVariadic() {
			continue
		}
		switch arg.GetType() {
		case ArgStrListT, ArgIntListT, ArgFloatListT, ArgBoolListT:
			return fmt.Sprintf(
				"Note: \"%s\" is a list arg, which takes a single positional value (lists are built by repeating the flag).\n"+
					"      To accept multiple positional values, declare it variadic: '*%s'",
				arg.GetIdentifier(), arg.GetIdentifier())
		}
	}
	return ""
}

// registerScript registers the script as a subcommand with its flags
func (r *RadRunner) registerScript() error {
	if r.scriptData == nil {
		return fmt.Errorf("Bug! Script data expected but not found")
	}

	if r.scriptData.HasArgsBlock && r.scriptData.DisableArgsBlock {
		return fmt.Errorf("Macro '%s' disabled, but args block found.\n", MACRO_ENABLE_ARGS_BLOCK)
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

	// Shell completion: bypass the double-parse entirely.
	// When the shell calls us for completion, we just need the registered args/commands
	// (already set up in Phase 2) and Ra's completion logic. No printer setup, no
	// version checks, no script execution needed.
	if len(argsToRead) > 0 && argsToRead[0] == "__complete" {
		RRootCmd.EnableCompletion()
		RRootCmd.ParseOrExit(argsToRead)
		return nil // ParseOrExit calls exit(0) on completion; we never reach here
	}

	// First parse: ignoreUnknown=true since script args aren't registered yet
	parseOpts := []ra.ParseOpt{ra.WithIgnoreUnknown(true), ra.WithVariadicUnknownFlags(true)}
	if FlagRadArgsDump.Value {
		parseOpts = append(parseOpts, ra.WithDump(true))
	}

	RRootCmd.ParseOrExit(argsToRead, parseOpts...)

	// Reject before anything acts on the parsed flags (printer routing,
	// --version, inspection flags) - silently honoring an
	// inapplicable flag is how `rad fmt --check --shell` lost its output.
	if invocationType == EmbeddedCommand {
		r.rejectOutOfScopeGlobalFlags()
	}

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

	if FlagTlsInsecure.Value {
		RReq.SetInsecure(true)
	}

	if FlagVersion.Value {
		printVersion()
		RExit.Exit(0)
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
		emitShellExit(0)
		RExit.Exit(0)
	}

	// Handle inspection flags for script output
	shouldExit := false
	if FlagSrc.Value {
		shouldExit = true
		printSource(r.scriptData.Src, FlagVersion.Value)
	}

	if FlagCstTree.Value {
		shouldExit = true
		printCstTree(r.scriptData.Tree, FlagSrc.Value)
	}

	if FlagAstTree.Value {
		shouldExit = true
		printAstTree(r.scriptData.Tree, r.scriptData.Src, r.scriptData.ScriptName, FlagSrc.Value || FlagCstTree.Value)
	}

	if shouldExit {
		RExit.Exit(0)
	}

	// Interactive pre-pass: prompt for args missing from the CLI and fold the
	// answers into the argv the second parse (below) will validate.
	if FlagInteractive.Value && r.scriptData != nil {
		argsToRead = r.runInteractivePrepass(argsToRead)
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

	// Determine the terminal (deepest) invoked command, if any.
	invokedCommand, commandArgs := r.resolveInvokedCommand()

	if len(r.cmdInvocations) > 0 {
		inspecting := FlagHelp.Value || FlagVersion.Value || FlagSrc.Value || FlagCstTree.Value || FlagAstTree.Value
		if invokedCommand == nil {
			if !inspecting {
				RP.UsageErrorExit("Must specify a command")
			}
		} else if invokedCommand.IsNamespace() {
			// A namespace groups and shares args; it has nothing of its own to
			// run, so stopping here means the user never named a leaf. Show
			// that namespace's commands, not the root's.
			if !inspecting {
				r.usageCmd = r.raCmdFor(invokedCommand)
				RP.UsageErrorExit(fmt.Sprintf("Must specify a command for '%s'", invokedCommand.ExternalName))
			}
			invokedCommand, commandArgs = nil, nil
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
		if err := r.registerCommandTree(RRootCmd, scriptCmd); err != nil {
			return err
		}
	}

	return nil
}

// raCmdFor returns the Ra command mirroring a script command, or nil if it was
// never registered.
func (r *RadRunner) raCmdFor(cmd *ScriptCommand) *ra.Cmd {
	for _, inv := range r.cmdInvocations {
		if inv.cmd == cmd {
			return inv.raCmd
		}
	}
	return nil
}

// resolveInvokedCommand finds the terminal (deepest) invoked command and the
// args in scope for its callback: the shared args inherited from every
// namespace on the path, then the command's own, with the deeper declaration
// winning a name collision - matching how Ra shadows them.
//
// Ra marks every command along the invoked path as used, not just the leaf, so
// this descends rather than taking the first match. A first-match loop compiles
// and quietly dispatches the namespace instead of the command the user asked
// for.
func (r *RadRunner) resolveInvokedCommand() (*ScriptCommand, []RadArg) {
	used := make(map[*ScriptCommand]bool, len(r.cmdInvocations))
	argsByCmd := make(map[*ScriptCommand][]RadArg, len(r.cmdInvocations))
	for _, inv := range r.cmdInvocations {
		used[inv.cmd] = *inv.usedPtr
		argsByCmd[inv.cmd] = inv.args
	}

	var descend func(cmd *ScriptCommand) *ScriptCommand
	descend = func(cmd *ScriptCommand) *ScriptCommand {
		for _, sub := range cmd.SubCmds {
			if used[sub] {
				return descend(sub)
			}
		}
		return cmd
	}

	var terminal *ScriptCommand
	for _, top := range r.scriptData.Commands {
		if used[top] {
			terminal = descend(top)
			break
		}
	}
	if terminal == nil {
		return nil, nil
	}

	var chain []*ScriptCommand
	for c := terminal; c != nil; c = c.Parent {
		chain = append([]*ScriptCommand{c}, chain...)
	}

	var pathArgs []RadArg
	idxByID := make(map[string]int)
	for _, c := range chain {
		for _, a := range argsByCmd[c] {
			id := a.GetIdentifier()
			if idx, ok := idxByID[id]; ok {
				pathArgs[idx] = a
			} else {
				idxByID[id] = len(pathArgs)
				pathArgs = append(pathArgs, a)
			}
		}
	}

	return terminal, pathArgs
}

func (r *RadRunner) registerCommandTree(parent *ra.Cmd, scriptCmd *ScriptCommand) error {
	return buildRaCmdTree(parent, scriptCmd, r.scriptArgs,
		func(cmd *ScriptCommand, raCmd *ra.Cmd, usedPtr *bool, cmdArgs []RadArg) {
			r.cmdInvocations = append(r.cmdInvocations, cmdInvocation{
				cmd:     cmd,
				usedPtr: usedPtr,
				args:    cmdArgs,
				raCmd:   raCmd,
			})
		})
}

// buildRaCmdTree mirrors a ScriptCommand and its descendants onto Ra's command
// tree, calling onNode for each once it is registered. Shell completion builds
// the same tree without the invocation tracking, and the two must agree: a
// completion that offers commands the parser won't accept is worse than one
// that offers nothing.
//
// Ordering is load-bearing. RegisterCmd copies the parent's globals into the
// child, so a namespace's shared args must be registered on it before its
// children are, or they never reach them.
func buildRaCmdTree(
	parent *ra.Cmd,
	scriptCmd *ScriptCommand,
	scriptArgs []RadArg,
	onNode func(cmd *ScriptCommand, raCmd *ra.Cmd, usedPtr *bool, cmdArgs []RadArg),
) error {
	raCmd := ra.NewCmd(scriptCmd.ExternalName)
	if scriptCmd.Description != nil {
		raCmd.SetDescription(*scriptCmd.Description)
	}

	raCmd.SetUsageHeaders(ra.UsageHeaders{
		Usage:                 "Usage:",
		Commands:              "Commands:",
		Arguments:             "Command args:",
		GlobalOptions:         "Global options:",
		SubcommandPlaceholder: "command",
	})
	raCmd.SetHelpEnabled(true)

	// A bare namespace should behave like a bare script: show what can be run
	// next, rather than complaining about args for a command nobody named.
	if scriptCmd.IsNamespace() {
		raCmd.SetAutoHelpOnNoArgs(true)
	}

	// Script args are shared across every command but flag-only, so register
	// them at each level - they stay valid wherever you are on the path.
	for _, scriptArg := range scriptArgs {
		scriptArg.Register(raCmd, AsScriptFlagOnly)
	}

	// This command's own args. A namespace's are shared with its descendants;
	// a leaf's are its own, and may be given positionally.
	argMode := AsCommandArg
	if scriptCmd.IsNamespace() {
		argMode = AsSharedNamespaceArg
	}
	cmdArgs := make([]RadArg, 0, len(scriptCmd.Args))
	for _, arg := range scriptCmd.Args {
		flag := CreateFlag(arg)
		flag.Register(raCmd, argMode)
		cmdArgs = append(cmdArgs, flag)
	}

	usedPtr, err := parent.RegisterCmd(raCmd)
	if err != nil {
		return fmt.Errorf("failed to register command '%s': %w", scriptCmd.ExternalName, err)
	}

	if onNode != nil {
		onNode(scriptCmd, raCmd, usedPtr, cmdArgs)
	}

	for _, sub := range scriptCmd.SubCmds {
		if err := buildRaCmdTree(raCmd, sub, scriptArgs, onNode); err != nil {
			return err
		}
	}

	return nil
}

// rejectOutOfScopeGlobalFlags exits with a usage error if an embedded command
// invocation set a global flag outside its scope, e.g. `rad fmt --shell`.
// The flags stay registered and are rejected *after* parsing: an unregistered
// flag would be consumed as a value by variadic positionals (e.g. fmt's
// *paths) instead of erroring.
func (r *RadRunner) rejectOutOfScopeGlobalFlags() {
	for _, scoped := range GlobalFlagScopes {
		if scoped.Scope == ScopeUniversal {
			continue
		}
		name := scoped.Arg.GetExternalName()
		if !RRootCmd.Configured(name) {
			continue
		}
		if scoped.Scope == ScopeRootOnly {
			RP.UsageErrorExit(fmt.Sprintf("--%s applies to rad itself. Run 'rad --%s' instead.", name, name))
		} else {
			RP.UsageErrorExit(
				fmt.Sprintf("--%s only applies when running a Rad script; %q is a built-in rad command.", name, ScriptPath),
			)
		}
		return
	}
}

func readSource(scriptPath string) (string, error) {
	source, err := os.ReadFile(scriptPath)
	// Normalize line endings for consistent behavior across platforms
	return NormalizeLineEndings(string(source)), err
}

// handleReplCommand handles `rad repl`, starting an interactive REPL session.
// Like `rad completion`, this is a Go-implemented command handled outside the
// embedded-script path. Arg parsing and help output are delegated to Ra.
func (r *RadRunner) handleReplCommand(args []string) error {
	// The REPL and Rad's exit handler both need the printer, which the normal
	// flow only sets up later in detectAndSetup.
	RP = NewPrinter(r, false, false, false, false)

	cmd := ra.NewCmd("rad repl")
	cmd.SetDescription(replDescription)
	cmd.SetHelpEnabled(true)
	cmd.ParseOrExit(args) // handles -h/--help, rejects unknown args

	return RunRepl()
}

// Helper functions for inspection flags (--version, --src, --cst-tree, --ast-tree)
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

func printCstTree(tree *rts.RadTree, prependNewline bool) {
	if prependNewline {
		RP.Printf("\n")
	}
	RP.Print(tree.Dump())
}

func printAstTree(tree *rts.RadTree, src string, file string, prependNewline bool) {
	if prependNewline {
		RP.Printf("\n")
	}
	astRoot := rts.ConvertCST(tree.Root(), src, file)
	RP.Print(rl.AstDump(astRoot))
}

// handleGlobalInspectionFlagsOnInvalidSyntax checks os.Args for inspection flags and handles them.
// This is called from validateSyntax() when the script has errors, before showing the error.
// Returns true if a flag was handled (and the function exited), false otherwise.
func handleGlobalInspectionFlagsOnInvalidSyntax(src string, tree *rts.RadTree) {
	// During shell completion, flags like "--src" or "--version" in the args are
	// completion context, not actual flag invocations. Skip inspection handling
	// to avoid corrupting completion output on stdout.
	if isCompletionRequest() {
		return
	}

	hasVersion := false
	hasSrc := false
	hasCstTree := false

	for _, arg := range os.Args { // todo don't love the hardcoded string lookups
		if arg == "--version" || arg == "-v" {
			hasVersion = true
		}
		if arg == "--src" {
			hasSrc = true
		}
		if arg == "--cst-tree" {
			hasCstTree = true
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

	if hasCstTree {
		printCstTree(tree, hasSrc)
		RExit.Exit(0)
	}
}
