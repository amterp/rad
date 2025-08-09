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
	FLAG_SRC_TREE      = "src-tree"
	FLAG_MOCK_RESPONSE = "mock-response"
)

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
	FlagRadTree              BoolRadArg
	FlagMockResponse         StringRadArg
	// ^ when adding more, update ResetGlobals function
)

func CreateAndRegisterGlobalFlags() []RadArg {
	// ordering of this list matters -- it's the order in which they are printed in the usage string
	flags := make([]RadArg, 0)

	FlagHelp = NewBoolRadArg(
		FLAG_HELP,
		FLAG_H,
		"Print usage string.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	flags = append(flags, &FlagHelp)

	FlagDebug = NewBoolRadArg(
		FLAG_DEBUG,
		FLAG_D,
		"Enables debug output. Intended for Rad script developers.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	flags = append(flags, &FlagDebug)

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
	flags = append(flags, &FlagRadDebug)

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
	flags = append(flags, &FlagColor)

	FlagQuiet = NewBoolRadArg(
		FLAG_QUIET,
		FLAG_Q,
		"Suppresses some output.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	flags = append(flags, &FlagQuiet)

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
	flags = append(flags, &FlagShell)

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
	flags = append(flags, &FlagVersion)

	FlagConfirmShellCommands = NewBoolRadArg(
		FLAG_CONFIRM_SHELL,
		"",
		"Confirm all shell commands before running them.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	flags = append(flags, &FlagConfirmShellCommands)

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
	flags = append(flags, &FlagSrc)

	FlagRadTree = NewBoolRadArg(
		FLAG_SRC_TREE,
		"",
		"Instead of running the target script, print out its syntax tree.",
		false,
		false,
		NO_CONSTRAINTS,
		NO_CONSTRAINTS,
	)
	FlagRadTree.SetBypassValidation(true)
	hideFromUsageIfHaveScript(&FlagRadTree.hidden)
	flags = append(flags, &FlagRadTree)

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
	flags = append(flags, &FlagMockResponse)

	registerGlobalFlags(flags)
	return flags
}

func hideFromUsageIfHaveScript(hidden *bool) {
	*hidden = HasScript
}

func registerGlobalFlags(flags []RadArg) {
	for _, flag := range flags {
		flag.Register(RRootCmd, true)
	}
}
