package core

import (
	"github.com/spf13/pflag"
)

var (
	ShellFlag       bool
	StdinScriptName string
	QuietFlag       bool
	DebugFlag       bool
	RadDebugFlag    bool
	MockResponses   MockResponseSlice
	NoColorFlag     bool
)

func DefineGlobalFlags() {
	pflag.BoolVar(&ShellFlag, "SHELL", false, "Outputs shell/bash exports of variables, so they can be eval'd")
	pflag.StringVar(&StdinScriptName, "STDIN", "", "Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name', usually $0")
	pflag.BoolVar(&QuietFlag, "QUIET", false, "Suppresses some output.")
	pflag.BoolVar(&DebugFlag, "DEBUG", false, "Enables debug output. Intended for RSL script developers.")
	pflag.BoolVar(&RadDebugFlag, "RAD-DEBUG", false, "Enables Rad debug output. Intended for Rad developers.")
	// todo help prints as `--MOCK-RESPONSE mockResponse` which is not ideal
	pflag.Var(&MockResponses, "MOCK-RESPONSE", "Add mock response for json requests (pattern:filePath)")
	pflag.BoolVar(&NoColorFlag, "NO-COLOR", false, "Disable colorized output")
}

func HideGlobalFlags() {
	pflag.Lookup("version").Hidden = true
	pflag.Lookup("help").Hidden = true
	pflag.Lookup("SHELL").Hidden = true
	pflag.Lookup("STDIN").Hidden = true
	pflag.Lookup("QUIET").Hidden = true
	pflag.Lookup("DEBUG").Hidden = true
	pflag.Lookup("RAD-DEBUG").Hidden = true
	pflag.Lookup("MOCK-RESPONSE").Hidden = true
	pflag.Lookup("NO-COLOR").Hidden = true
}
