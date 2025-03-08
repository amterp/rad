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
	FLAG_DEBUG         = "DEBUG"
	FLAG_D             = "D"
	FLAG_RAD_DEBUG     = "RAD-DEBUG"
	FLAG_COLOR         = "COLOR"
	FLAG_QUIET         = "QUIET"
	FLAG_Q             = "Q"
	FLAG_SHELL         = "SHELL"
	FLAG_VERSION       = "VERSION"
	FLAG_V             = "V"
	FLAG_CONFIRM_SHELL = "CONFIRM-SHELL"
	FLAG_SRC           = "SRC"
	FLAG_RSL_TREE      = "RSL-TREE"
	FLAG_MOCK_RESPONSE = "MOCK-RESPONSE"
)

var (
	MODES = []string{COLOR_AUTO, COLOR_ALWAYS, COLOR_NEVER}

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
		FlagHelp = NewBoolRadArg(flagOrShort(FLAG_HELP, FLAG_H), flagOrEmpty(FLAG_H), "Print usage string.", false)
		flags = append(flags, &FlagHelp)
	}

	if shouldAddFlag(FLAG_DEBUG, FLAG_D) {
		FlagDebug = NewBoolRadArg(flagOrShort(FLAG_DEBUG, FLAG_D), flagOrEmpty(FLAG_D), "Enables debug output. Intended for RSL script developers.", false)
		flags = append(flags, &FlagDebug)
	}

	if shouldAddFlag(FLAG_RAD_DEBUG, "") {
		FlagRadDebug = NewBoolRadArg(flagOrShort(FLAG_RAD_DEBUG, ""), flagOrEmpty(""), "Enables Rad debug output. Intended for Rad developers.", false)
		flags = append(flags, &FlagRadDebug)
	}

	if shouldAddFlag(FLAG_COLOR, "") {
		FlagColor = NewStringRadArg(flagOrShort(FLAG_COLOR, ""), flagOrEmpty(""), "mode", "Control output colorization.", "auto", &MODES, nil)
		flags = append(flags, &FlagColor)
	}

	if shouldAddFlag(FLAG_QUIET, FLAG_Q) {
		FlagQuiet = NewBoolRadArg(flagOrShort(FLAG_QUIET, FLAG_Q), flagOrEmpty(FLAG_Q), "Suppresses some output.", false)
		flags = append(flags, &FlagQuiet)
	}

	if shouldAddFlag(FLAG_SHELL, "") {
		FlagShell = NewBoolRadArg(flagOrShort(FLAG_SHELL, ""), flagOrEmpty(""), "Outputs shell/bash exports of variables, so they can be eval'd", false)
		flags = append(flags, &FlagShell)
	}

	if shouldAddFlag(FLAG_VERSION, FLAG_V) {
		FlagVersion = NewBoolRadArg(flagOrShort(FLAG_VERSION, FLAG_V), flagOrEmpty(FLAG_V), "Print rad version information.", false)
		flags = append(flags, &FlagVersion)
	}

	if shouldAddFlag(FLAG_CONFIRM_SHELL, "") {
		FlagConfirmShellCommands = NewBoolRadArg(flagOrShort(FLAG_CONFIRM_SHELL, ""), flagOrEmpty(""), "Confirm all shell commands before running them.", false)
		flags = append(flags, &FlagConfirmShellCommands)
	}

	if shouldAddFlag(FLAG_SRC, "") {
		FlagSrc = NewBoolRadArg(flagOrShort(FLAG_SRC, ""), flagOrEmpty(""), "Instead of running the target script, just print it out.", false)
		flags = append(flags, &FlagSrc)
	}

	if shouldAddFlag(FLAG_RSL_TREE, "") {
		FlagRslTree = NewBoolRadArg(flagOrShort(FLAG_RSL_TREE, ""), flagOrEmpty(""), "Instead of running the target script, print out its syntax tree.", false)
		flags = append(flags, &FlagRslTree)
	}

	if shouldAddFlag(FLAG_MOCK_RESPONSE, "") {
		FlagMockResponse = NewMockResponseRadArg(flagOrShort(FLAG_MOCK_RESPONSE, ""), flagOrEmpty(""), "Add mock response for json requests (pattern:filePath)")
		flags = append(flags, &FlagMockResponse)
	}

	registerGlobalFlags(flags)
	return flags
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

func flagOrShort(flag string, short string) string {
	if lo.Contains(FlagsUsedInScript, flag) {
		return short
	}
	return flag
}

func flagOrEmpty(flag string) string {
	if lo.Contains(FlagsUsedInScript, flag) {
		return ""
	}
	return flag
}
