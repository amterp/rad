package core

var (
	FlagHelp            = NewBoolRadFlag("help", "h", "Print usage string.", false)
	FlagDebug           = NewBoolRadFlag("DEBUG", "D", "Enables debug output. Intended for RSL script developers.", false)
	FlagRadDebug        = NewBoolRadFlag("RAD-DEBUG", "", "Enables Rad debug output. Intended for Rad developers.", false)
	FlagNoColor         = NewBoolRadFlag("NO-COLOR", "", "Disable colorized output.", false)
	FlagQuiet           = NewBoolRadFlag("QUIET", "Q", "Suppresses some output.", false)
	FlagShell           = NewBoolRadFlag("SHELL", "", "Outputs shell/bash exports of variables, so they can be eval'd", false)
	FlagStdinScriptName = NewStringRadFlag("STDIN", "", "script-name", "Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name'.", "")
	FlagMockResponse    = NewMockResponseRadFlag("MOCK-RESPONSE", "", "Add mock response for json requests (pattern:filePath)")

	Flags = []RslArg{
		&FlagHelp,
		&FlagDebug,
		&FlagRadDebug,
		&FlagNoColor,
		&FlagQuiet,
		&FlagShell,
		&FlagStdinScriptName,
		&FlagMockResponse,
	}
)

func RegisterGlobalFlags() []RslArg {
	for _, flag := range Flags {
		flag.Register()
	}
	return Flags
}
