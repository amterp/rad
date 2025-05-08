package core

import (
	com "rad/core/common"

	"github.com/samber/lo"
)

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
	FLAG_RSL_TREE      = "rsl-tree"
	FLAG_MOCK_RESPONSE = "mock-response"
)

var (
	MODES          = []string{COLOR_AUTO, COLOR_ALWAYS, COLOR_NEVER}
	NO_CONSTRAINTS []string

	FlagsUsedInScript []string

	FlagHelp     BoolRslArg
	FlagDebug    BoolRslArg
	FlagRadDebug BoolRslArg
	FlagColor    StringRslArg
	FlagQuiet    BoolRslArg
	FlagShell    BoolRslArg
	// todo allow scripts to override this global flag
	FlagVersion              BoolRslArg
	FlagConfirmShellCommands BoolRslArg
	FlagSrc                  BoolRslArg
	FlagRslTree              BoolRslArg
	FlagMockResponse         MockResponseRslArg

	// ordering here matters -- it's the order in which they are printed in the usage string
)

func CreateAndRegisterGlobalFlags() []RslArg {
	flags := make([]RslArg, 0)

	if shouldAddFlag(FLAG_HELP, FLAG_H) {
		FlagHelp = NewBoolRadArg(flagOrEmpty(FLAG_HELP), flagOrEmpty(FLAG_H), "Printf usage string.", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		flags = append(flags, &FlagHelp)
	}

	if shouldAddFlag(FLAG_DEBUG, FLAG_D) {
		FlagDebug = NewBoolRadArg(flagOrEmpty(FLAG_DEBUG), flagOrEmpty(FLAG_D), "Enables debug output. Intended for RSL script developers.", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		flags = append(flags, &FlagDebug)
	}

	if shouldAddFlag(FLAG_RAD_DEBUG, "") {
		FlagRadDebug = NewBoolRadArg(flagOrEmpty(FLAG_RAD_DEBUG), flagOrEmpty(""), "Enables Rad debug output. Intended for Rad developers.", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		hideFromUsageIfHaveScript(&FlagRadDebug.hidden)
		flags = append(flags, &FlagRadDebug)
	}

	if shouldAddFlag(FLAG_COLOR, "") {
		FlagColor = NewStringRadArg(flagOrEmpty(FLAG_COLOR), flagOrEmpty(""), "mode", "Control output colorization.", false, "auto", &MODES, nil, NO_CONSTRAINTS, NO_CONSTRAINTS)
		flags = append(flags, &FlagColor)
	}

	if shouldAddFlag(FLAG_QUIET, FLAG_Q) {
		FlagQuiet = NewBoolRadArg(flagOrEmpty(FLAG_QUIET), flagOrEmpty(FLAG_Q), "Suppresses some output.", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		flags = append(flags, &FlagQuiet)
	}

	if shouldAddFlag(FLAG_SHELL, "") {
		FlagShell = NewBoolRadArg(flagOrEmpty(FLAG_SHELL), flagOrEmpty(""), "Outputs shell/bash exports of variables, so they can be eval'd", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		hideFromUsageIfHaveScript(&FlagShell.hidden)
		flags = append(flags, &FlagShell)
	}

	if shouldAddFlag(FLAG_VERSION, FLAG_V) {
		FlagVersion = NewBoolRadArg(flagOrEmpty(FLAG_VERSION), flagOrEmpty(FLAG_V), "Printf rad version information.", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		hideFromUsageIfHaveScript(&FlagVersion.hidden)
		flags = append(flags, &FlagVersion)
	}

	if shouldAddFlag(FLAG_CONFIRM_SHELL, "") {
		FlagConfirmShellCommands = NewBoolRadArg(flagOrEmpty(FLAG_CONFIRM_SHELL), flagOrEmpty(""), "Confirm all shell commands before running them.", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		flags = append(flags, &FlagConfirmShellCommands)
	}

	if shouldAddFlag(FLAG_SRC, "") {
		FlagSrc = NewBoolRadArg(flagOrEmpty(FLAG_SRC), flagOrEmpty(""), "Instead of running the target script, just print it out.", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		flags = append(flags, &FlagSrc)
	}

	if shouldAddFlag(FLAG_RSL_TREE, "") {
		FlagRslTree = NewBoolRadArg(flagOrEmpty(FLAG_RSL_TREE), flagOrEmpty(""), "Instead of running the target script, print out its syntax tree.", false, false, NO_CONSTRAINTS, NO_CONSTRAINTS)
		hideFromUsageIfHaveScript(&FlagRslTree.hidden)
		flags = append(flags, &FlagRslTree)
	}

	if shouldAddFlag(FLAG_MOCK_RESPONSE, "") {
		FlagMockResponse = NewMockResponseRadArg(flagOrEmpty(FLAG_MOCK_RESPONSE), flagOrEmpty(""), "Add mock response for json requests (pattern:filePath)")
		hideFromUsageIfHaveScript(&FlagMockResponse.hidden)
		flags = append(flags, &FlagMockResponse)
	}

	registerGlobalFlags(flags)
	return flags
}

func hideFromUsageIfHaveScript(hidden *bool) {
	*hidden = HasScript
}

func registerGlobalFlags(flags []RslArg) {
	for _, flag := range flags {
		flag.Register()
	}
}

func shouldAddFlag(flag string, short string) bool {
	if !com.IsBlank(flag) && !lo.Contains(FlagsUsedInScript, flag) {
		return true
	}

	if !com.IsBlank(short) && !lo.Contains(FlagsUsedInScript, short) {
		return true
	}

	return false
}

func flagOrEmpty(flag string) string {
	if lo.Contains(FlagsUsedInScript, flag) {
		return ""
	}
	return flag
}
