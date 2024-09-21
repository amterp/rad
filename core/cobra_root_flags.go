package core

import "github.com/spf13/cobra"

var (
	shellFlag       bool
	stdinScriptName string
	quietFlag       bool
	debugFlag       bool
	radDebugFlag    bool
	mockResponses   MockResponseSlice
	noColorFlag     bool
)

func defineGlobalFlags(cmd *cobra.Command) {
	// global flags
	// todo think more about bash vs. shell
	cmd.PersistentFlags().BoolVar(&shellFlag, "SHELL", false, "Outputs shell/bash exports of variables, so they can be eval'd")
	cmd.PersistentFlags().StringVar(&stdinScriptName, "STDIN", "", "Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name', usually $0")
	cmd.PersistentFlags().BoolVar(&quietFlag, "QUIET", false, "Suppresses some output.")
	cmd.PersistentFlags().BoolVar(&debugFlag, "DEBUG", false, "Enables debug output. Intended for RSL script developers.")
	cmd.PersistentFlags().BoolVar(&radDebugFlag, "RAD-DEBUG", false, "Enables Rad debug output. Intended for Rad developers.")
	// todo help prints as `--MOCK-RESPONSE mockResponse` which is not ideal
	cmd.PersistentFlags().Var(&mockResponses, "MOCK-RESPONSE", "Add mock response for json requests (pattern:filePath)")
	cmd.PersistentFlags().BoolVar(&noColorFlag, "NO-COLOR", false, "Disable colorized output")
}

func hideGlobalFlags(cmd *cobra.Command) {
	cmd.Flags().MarkHidden("version")
	cmd.Flags().MarkHidden("help")
	cmd.PersistentFlags().MarkHidden("SHELL")
	cmd.PersistentFlags().MarkHidden("STDIN")
	cmd.PersistentFlags().MarkHidden("QUIET")
	cmd.PersistentFlags().MarkHidden("DEBUG")
	cmd.PersistentFlags().MarkHidden("RAD-DEBUG")
	cmd.PersistentFlags().MarkHidden("MOCK-RESPONSE")
	cmd.PersistentFlags().MarkHidden("NO-COLOR")
}
