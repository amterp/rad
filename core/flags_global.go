package core

var (
	FlagHelp            = NewBoolRadFlag("help", "h", "Print usage string.", false)
	FlagDebug           = NewBoolRadFlag("DEBUG", "D", "Enables debug output. Intended for RSL script developers.", false)
	FlagStdinScriptName = NewStringRadFlag("STDIN", "", "script-name", "Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name'.", "")
	FlagMockResponse    = NewMockResponseRadFlag("MOCK-RESPONSE", "", "Add mock response for json requests (pattern:filePath)")
)

func RegisterGlobalFlags() []RadFlag {
	FlagHelp.Register()
	FlagDebug.Register()
	FlagStdinScriptName.Register()
	FlagMockResponse.Register()
	return []RadFlag{&FlagDebug, &FlagStdinScriptName, &FlagMockResponse, &FlagHelp}
}
