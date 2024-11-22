package core

var (
	FlagHelp     BoolRslArg
	FlagDebug    BoolRslArg
	FlagRadDebug BoolRslArg
	FlagNoColor  BoolRslArg
	FlagQuiet    BoolRslArg
	FlagShell    BoolRslArg
	// todo allow scripts to override this global flag
	FlagVersion         BoolRslArg
	FlagStdinScriptName StringRslArg
	FlagMockResponse    MockResponseRslArg

	// ordering here matters -- it's the order in which they are printed in the usage string
	Flags = []RslArg{
		&FlagHelp,
		&FlagDebug,
		&FlagRadDebug,
		&FlagNoColor,
		&FlagQuiet,
		&FlagShell,
		&FlagVersion,
		&FlagStdinScriptName,
		&FlagMockResponse,
	}
)

func CreateAndRegisterGlobalFlags() []RslArg {
	FlagHelp = NewBoolRadArg("help", "h", "Print usage string.", false)
	FlagDebug = NewBoolRadArg("DEBUG", "D", "Enables debug output. Intended for RSL script developers.", false)
	FlagRadDebug = NewBoolRadArg("RAD-DEBUG", "", "Enables Rad debug output. Intended for Rad developers.", false)
	FlagNoColor = NewBoolRadArg("NO-COLOR", "", "Disable colorized output.", false)
	FlagQuiet = NewBoolRadArg("QUIET", "Q", "Suppresses some output.", false)
	FlagShell = NewBoolRadArg("SHELL", "", "Outputs shell/bash exports of variables, so they can be eval'd", false)
	FlagVersion = NewBoolRadArg("VERSION", "V", "Print rad version information.", false)
	FlagStdinScriptName = NewStringRadArg("STDIN", "", "script-name", "Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name'.", "")
	FlagMockResponse = NewMockResponseRadArg("MOCK-RESPONSE", "", "Add mock response for json requests (pattern:filePath)")
	registerGlobalFlags()
	return Flags
}

func registerGlobalFlags() {
	for _, flag := range Flags {
		flag.Register()
	}
}
