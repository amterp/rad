package core

const (
	COLOR_AUTO   = "auto"
	COLOR_ALWAYS = "always"
	COLOR_NEVER  = "never"
)

var (
	MODES = []string{COLOR_AUTO, COLOR_ALWAYS, COLOR_NEVER}

	FlagHelp     BoolRslArg
	FlagDebug    BoolRslArg
	FlagRadDebug BoolRslArg
	FlagColor    StringRslArg
	FlagQuiet    BoolRslArg
	FlagShell    BoolRslArg
	// todo allow scripts to override this global flag
	FlagVersion              BoolRslArg
	FlagStdinScriptName      StringRslArg
	FlagConfirmShellCommands BoolRslArg
	FlagSrc                  BoolRslArg
	FlagRslTree              BoolRslArg
	FlagMockResponse         MockResponseRslArg

	// ordering here matters -- it's the order in which they are printed in the usage string
	Flags = []RslArg{
		&FlagHelp,
		&FlagDebug,
		&FlagRadDebug,
		&FlagColor,
		&FlagQuiet,
		&FlagShell,
		&FlagVersion,
		&FlagStdinScriptName,
		&FlagConfirmShellCommands,
		&FlagSrc,
		&FlagRslTree,
		&FlagMockResponse,
	}
)

func CreateAndRegisterGlobalFlags() []RslArg {
	FlagHelp = NewBoolRadArg("help", "h", "Print usage string.", false)
	FlagDebug = NewBoolRadArg("DEBUG", "D", "Enables debug output. Intended for RSL script developers.", false)
	FlagRadDebug = NewBoolRadArg("RAD-DEBUG", "", "Enables Rad debug output. Intended for Rad developers.", false)
	FlagColor = NewStringRadArg("COLOR", "", "mode", "Control output colorization.", "auto", &MODES, nil)
	FlagQuiet = NewBoolRadArg("QUIET", "Q", "Suppresses some output.", false)
	FlagShell = NewBoolRadArg("SHELL", "", "Outputs shell/bash exports of variables, so they can be eval'd", false)
	FlagVersion = NewBoolRadArg("VERSION", "V", "Print rad version information.", false)
	FlagStdinScriptName = NewStringRadArg("STDIN", "", "script-name", "Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name'.", "", nil, nil)
	FlagConfirmShellCommands = NewBoolRadArg("CONFIRM-SHELL", "", "Confirm all shell commands before running them.", false)
	FlagSrc = NewBoolRadArg("SRC", "", "Instead of running the target script, just print it out.", false)
	FlagRslTree = NewBoolRadArg("RSL-TREE", "", "Instead of running the target script, print out its syntax tree.", false)
	FlagMockResponse = NewMockResponseRadArg("MOCK-RESPONSE", "", "Add mock response for json requests (pattern:filePath)")
	registerGlobalFlags()
	return Flags
}

func registerGlobalFlags() {
	for _, flag := range Flags {
		flag.Register()
	}
}
