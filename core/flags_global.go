package core

const (
	COLOR_AUTO   = "auto"
	COLOR_ALWAYS = "always"
	COLOR_NEVER  = "never"

	FLAG_HELP          = "help"
	FLAG_H             = "h"
	FLAG_DEBUG         = "debug"
	FLAG_D             = "d"
	FLAG_RAD_DEBUG     = "rad-debug"
	FLAG_COLOR         = "color"
	FLAG_QUIET         = "quiet"
	FLAG_Q             = "q"
	FLAG_SHELL         = "shell"
	FLAG_VERSION       = "version"
	FLAG_V             = "v"
	FLAG_CONFIRM_SHELL = "confirm-shell"
	FLAG_SRC           = "src"
	FLAG_CST_TREE      = "cst-tree"
	FLAG_AST_TREE      = "ast-tree"
	FLAG_RAD_ARGS_DUMP = "rad-args-dump"
	FLAG_MOCK_RESPONSE = "mock-response"
	FLAG_REPL          = "repl"
	FLAG_R             = "r"
	FLAG_TLS_INSECURE  = "tls-insecure"
	FLAG_INTERACTIVE   = "interactive"
	FLAG_I             = "i"
)

// GlobalFlagScope classifies which invocations a global flag applies to.
// Script-execution flags (e.g. --shell) and rad-level flags (e.g. --version)
// make no sense on embedded commands like `rad fmt`; those invocations reject
// them rather than silently changing behavior (issue #149).
type GlobalFlagScope int

const (
	ScopeUniversal  GlobalFlagScope = iota // any invocation, including embedded commands
	ScopeScriptOnly                        // configures Rad script execution
	ScopeRootOnly                          // rad itself, e.g. `rad --version`
)

type ScopedGlobalFlag struct {
	Arg   RadArg
	Scope GlobalFlagScope
}

var (
	MODES          = []string{COLOR_AUTO, COLOR_ALWAYS, COLOR_NEVER}
	NO_CONSTRAINTS []string

	FlagHelp                 BoolRadArg
	FlagDebug                BoolRadArg
	FlagRadDebug             BoolRadArg
	FlagColor                StringRadArg
	FlagQuiet                BoolRadArg
	FlagShell                BoolRadArg
	FlagVersion              BoolRadArg
	FlagConfirmShellCommands BoolRadArg
	FlagSrc                  BoolRadArg
	FlagCstTree              BoolRadArg
	FlagAstTree              BoolRadArg
	FlagRadArgsDump          BoolRadArg
	FlagMockResponse         StringRadArg
	FlagRepl                 BoolRadArg
	FlagTlsInsecure          BoolRadArg
	FlagInteractive          BoolRadArg
	// ^ when adding more, update ResetGlobals and the GlobalFlagScopes table below

	// GlobalFlagScopes is the single source of truth for global flag ordering
	// (as printed in usage) and scope. Rebuilt by CreateAndRegisterGlobalFlags.
	GlobalFlagScopes []ScopedGlobalFlag
)

func CreateAndRegisterGlobalFlags(invocationType InvocationType) []RadArg {
	FlagHelp = NewBoolRadArg(
		FLAG_HELP,
		FLAG_H,
		"Print usage string.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)

	FlagRepl = NewBoolRadArg(
		FLAG_REPL,
		FLAG_R,
		"Start interactive REPL mode.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)

	FlagInteractive = NewBoolRadArg(
		FLAG_INTERACTIVE,
		FLAG_I,
		"Interactively prompt for script args not already provided, then run.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	// Bypass so the first parse doesn't die on missing required args before we
	// get the chance to prompt for them. The pre-pass strips -i from the argv it
	// synthesizes, so the final parse still enforces everything.
	FlagInteractive.SetBypassValidation(true)

	FlagDebug = NewBoolRadArg(
		FLAG_DEBUG,
		FLAG_D,
		"Enables debug output. Intended for Rad script developers.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)

	FlagRadDebug = NewBoolRadArg(
		FLAG_RAD_DEBUG,
		"",
		"Enables Rad debug output. Intended for Rad developers.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	hideFromUsageIfHaveScript(&FlagRadDebug.hidden)

	FlagColor = NewStringRadArg(
		FLAG_COLOR,
		"",
		"mode",
		"Control output colorization.",
		true,
		"auto",
		&MODES,
		nil,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)

	FlagQuiet = NewBoolRadArg(
		FLAG_QUIET,
		FLAG_Q,
		"Suppresses some output.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)

	FlagShell = NewBoolRadArg(
		FLAG_SHELL,
		"",
		"Outputs shell/bash exports of variables, so they can be eval'd",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	hideFromUsageIfHaveScript(&FlagShell.hidden)

	FlagVersion = NewBoolRadArg(
		FLAG_VERSION,
		FLAG_V,
		"Print rad version information.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	FlagVersion.SetBypassValidation(true)
	hideFromUsageIfHaveScript(&FlagVersion.hidden)

	FlagConfirmShellCommands = NewBoolRadArg(
		FLAG_CONFIRM_SHELL,
		"",
		"Confirm all shell commands before running them.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)

	FlagTlsInsecure = NewBoolRadArg(
		FLAG_TLS_INSECURE,
		"",
		"Skip TLS certificate verification for all HTTP requests.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)

	FlagSrc = NewBoolRadArg(
		FLAG_SRC,
		"",
		"Instead of running the target script, just print it out.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	FlagSrc.SetBypassValidation(true)

	FlagCstTree = NewBoolRadArg(
		FLAG_CST_TREE,
		"",
		"Instead of running the target script, print out its CST (concrete syntax tree).",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	FlagCstTree.SetBypassValidation(true)
	hideFromUsageIfHaveScript(&FlagCstTree.hidden)

	FlagAstTree = NewBoolRadArg(
		FLAG_AST_TREE,
		"",
		"Instead of running the target script, print out its AST (abstract syntax tree).",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	FlagAstTree.SetBypassValidation(true)
	hideFromUsageIfHaveScript(&FlagAstTree.hidden)

	FlagRadArgsDump = NewBoolRadArg(
		FLAG_RAD_ARGS_DUMP,
		"",
		"Instead of running the target script, print out an args dump for debugging argument parsing.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	FlagRadArgsDump.SetBypassValidation(true)
	hideFromUsageIfHaveScript(&FlagRadArgsDump.hidden)

	FlagMockResponse = NewStringRadArg(
		FLAG_MOCK_RESPONSE,
		"",
		"str",
		// "pattern:filePath", // todo more descriptive
		"Add mock response for json requests (pattern:filePath)",
		false,
		"",
		nil,
		nil,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	hideFromUsageIfHaveScript(&FlagMockResponse.hidden)

	// ordering of this table matters -- it's the order in which flags are printed in the usage string
	GlobalFlagScopes = []ScopedGlobalFlag{
		{&FlagHelp, ScopeUniversal},
		{&FlagRepl, ScopeRootOnly},
		{&FlagInteractive, ScopeUniversal},
		{&FlagDebug, ScopeUniversal},
		{&FlagRadDebug, ScopeUniversal},
		{&FlagColor, ScopeUniversal},
		{&FlagQuiet, ScopeUniversal},
		{&FlagShell, ScopeScriptOnly},
		{&FlagVersion, ScopeRootOnly},
		{&FlagConfirmShellCommands, ScopeScriptOnly},
		{&FlagTlsInsecure, ScopeScriptOnly},
		{&FlagSrc, ScopeScriptOnly},
		{&FlagCstTree, ScopeScriptOnly},
		{&FlagAstTree, ScopeScriptOnly},
		{&FlagRadArgsDump, ScopeUniversal},
		{&FlagMockResponse, ScopeScriptOnly},
	}

	flags := make([]RadArg, 0, len(GlobalFlagScopes))
	for _, scoped := range GlobalFlagScopes {
		flags = append(flags, scoped.Arg)
	}

	if invocationType == EmbeddedCommand {
		hideInapplicableFlagsForEmbedded()
	}

	registerGlobalFlags(flags)
	return flags
}

func hideFromUsageIfHaveScript(hidden *bool) {
	*hidden = HasScript
}

// Embedded commands keep every global flag registered - an unregistered flag
// would be consumed by variadic positionals (e.g. fmt's *paths) instead of
// erroring - but only the applicable ones belong in their help output.
// Out-of-scope flags are rejected post-parse; see rejectOutOfScopeGlobalFlags.
func hideInapplicableFlagsForEmbedded() {
	visible := map[string]bool{
		FLAG_HELP:        true,
		FLAG_INTERACTIVE: true,
		FLAG_COLOR:       true,
		FLAG_QUIET:       true,
	}
	for _, scoped := range GlobalFlagScopes {
		if !visible[scoped.Arg.GetExternalName()] {
			scoped.Arg.Hidden(true)
		}
	}
}

func registerGlobalFlags(flags []RadArg) {
	for _, flag := range flags {
		flag.Register(RRootCmd, AsGlobalFlag)
	}
}
