package core

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	rootModified bool
)

func NewRootCmd(runnerInput RunnerInput) *cobra.Command {
	setGlobals(runnerInput)
	rootModified = false

	return &cobra.Command{
		Use:     "",
		Short:   "Request And Display (RAD)",
		Long:    `Request And Display (RAD): A tool for making HTTP requests, extracting details, and displaying the result.`,
		Version: "0.4.12",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !rootModified {

			}
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}

func registerInterpreterWithExit(interpreter *MainInterpreter) {
	existing := RExit
	exiting := false
	codeToExitWith := 0
	RExit = func(code int) {
		if exiting {
			// we're already exiting. if we're here again, it's probably because one of the deferred
			// statements is calling exit again (perhaps because it failed). we should keep running
			// all the deferred statements, however, and *then* exit.
			// therefore, we panic here in order to send the stack back up to where the deferred statement is being
			// invoked in the interpreter, which should be wrapped in a recover() block to catch, maybe log, and move on.
			if codeToExitWith == 0 {
				codeToExitWith = code
			}
			panic(code)
		}
		exiting = true
		codeToExitWith = code
		interpreter.ExecuteDeferredStmts(code)
		existing(codeToExitWith)
	}
}

func readSource(scriptPath string) string {
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		RP.RadErrorExit(fmt.Sprintf("Could not read script '%s': %v\n", scriptPath, err))
		RExit(1)
	}
	return string(source)
}

func InitCmd(cmd *cobra.Command) {
	// this is a bit hacky, but bear with me!
	// we intercept the very first call to help, it implies that the user has set the help flag
	// however, we won't yet have read the RSL script if it was also provided, so we may first
	// try to read that, register the args, etc, that's relevant to help, and then re-run the help
	// command, so that it can print help for the script
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// immediately reset the help func, as we only want this hacked
		// version to run once
		cmd.SetHelpFunc(nil)

		// try to detect if help has been called on either a script or with --STDIN flag
		if len(args) >= 2 {

		}

		cmd.Help()

		if FlagShell.Value && FlagStdinScriptName.Value != "" {
			// if both these flags are set, we're likely being invoked from within a bash script, so let's
			// output an exit 0 for bash to eval and exit, so it doesn't continue
			RP.PrintForShellEval("exit 0")
		}
	})

	cmd.SetOut(RIo.StdErr)
}

func Execute() {
	rootCmd := NewRootCmd(RunnerInput{})
	InitCmd(rootCmd)
	err := rootCmd.Execute()
	if err != nil {
		RExit(1)
	}
}
