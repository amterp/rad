package core

import (
	"github.com/spf13/pflag"
)

var (
	ShellFlag    bool
	QuietFlag    bool
	RadDebugFlag bool
	NoColorFlag  bool
)

func DefineGlobalFlags() {
	pflag.BoolVar(&ShellFlag, "SHELL", false, "Outputs shell/bash exports of variables, so they can be eval'd")
	pflag.BoolVar(&QuietFlag, "QUIET", false, "Suppresses some output.")
	pflag.BoolVar(&RadDebugFlag, "RAD-DEBUG", false, "Enables Rad debug output. Intended for Rad developers.")
	// todo help prints as `--MOCK-RESPONSE mockResponse` which is not ideal
	pflag.BoolVar(&NoColorFlag, "NO-COLOR", false, "Disable colorized output")
}
