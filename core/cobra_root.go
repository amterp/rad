package core

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
	"os"
)

var (
	rootModified bool
)

func NewRootCmd(cmdInput CmdInput) *cobra.Command {
	setGlobals(cmdInput)
	rootModified = false

	return &cobra.Command{
		Use:     "",
		Short:   "Request And Display (RAD)",
		Long:    `Request And Display (RAD): A tool for making HTTP requests, extracting details, and displaying the result.`,
		Version: "0.3.12",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if RP == nil {
				RP = NewPrinter(cmd, shellFlag, quietFlag, debugFlag, radDebugFlag)
			}
			if !rootModified {
				RP.RadDebug(fmt.Sprintf("Args passed: %v", args))
				if radDebugFlag {
					cmd.Flags().VisitAll(func(flag *pflag.Flag) {
						RP.RadDebug(fmt.Sprintf("Flag %s: %v", flag.Name, flag.Value))
					})
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			for _, mockResponse := range mockResponses {
				RReq.AddMockedResponse(mockResponse.Pattern, mockResponse.FilePath)
				RP.RadDebug(fmt.Sprintf("Mock response added: %q -> %q", mockResponse.Pattern, mockResponse.FilePath))
			}

			var rslSourceCode string
			if stdinScriptName != "" {
				// we're in stdin mode
				SetScriptPath(stdinScriptName)
				source, err := io.ReadAll(RIo.StdIn)
				if err != nil {
					RP.RadErrorExit(fmt.Sprintf("Could not read from stdin: %v\n", err))
				}
				rslSourceCode = string(source)
			} else if len(args) == 0 {
				cmd.Help()
				return
			} else {
				SetScriptPath(args[0])
				rslSourceCode = readSource(ScriptPath)
			}

			if rootModified {
				return
			}

			extractMetadataAndModifyCmd(cmd, rslSourceCode)
			cmd.Execute()
		},
	}
}

func extractMetadataAndModifyCmd(cmd *cobra.Command, rslSourceCode string) {
	l := NewLexer(RP, rslSourceCode)
	l.Lex()

	p := NewParser(RP, l.Tokens)
	instructions := p.Parse()

	scriptMetadata := ExtractMetadata(instructions)
	modifyCmd(cmd, ScriptName, scriptMetadata, instructions)
	rootModified = true
}

func modifyCmd(cmd *cobra.Command, scriptName string, scriptMetadata ScriptMetadata, instructions []Stmt) {
	useString := GenerateUseString(scriptName, scriptMetadata.Args)
	var cobraArgs []*CobraArg
	cmd.Use = useString
	cmd.Short = ShortDescription(scriptMetadata)
	cmd.Long = LongDescription(scriptMetadata)
	cmd.FParseErrWhitelist = cobra.FParseErrWhitelist{} // re-enable erroring on unknown flags. note: maybe remove for 'catchall' args?
	cmd.Run = func(cmd *cobra.Command, args []string) {
		// fill in positional args, and
		// error if required args are missing
		posArgsIndex := 0
		if stdinScriptName == "" {
			// We're invoked on an actual string path, which will be the first arg. Cut it out.
			args = args[1:]
		}
		var missingArgs []string
		for _, cobraArg := range cobraArgs {
			argName := cobraArg.Arg.ApiName
			cobraFlag := cmd.Flags().Lookup(argName)
			if !cobraFlag.Changed {
				// flag has not been explicitly set by the user
				if posArgsIndex < len(args) {
					// there's a positional arg to fill it
					cobraArg.SetValue(args[posArgsIndex])
					posArgsIndex++
				} else if cobraArg.Arg.IsOptional {
					// there's no positional arg to fill it, but that's okay because it's optional, so continue
					// but first, fill in the optional's default value if it exists
					cobraArg.InitializeOptional()
					continue
				} else if cobraArg.IsBool() {
					// all bools are implicitly optional and default false, unless explicitly defaulted to true
					// this branch implies it was not defaulted to true
					cobraArg.SetValue("false")
				} else {
					missingArgs = append(missingArgs, argName)
				}
			}
		}

		if len(missingArgs) > 0 && len(args) == 0 {
			cmd.Help()
			return
		}

		if len(missingArgs) > 0 {
			RP.UsageErrorExit(fmt.Sprintf("Missing required arguments: %s\n", missingArgs))
		}

		// error if not all positional args were used
		if posArgsIndex < len(args) {
			RP.UsageErrorExit(fmt.Sprintf("Too many positional arguments. Unused: %v\n", args[posArgsIndex:]))
		}

		color.NoColor = noColorFlag
		interpreter := NewInterpreter(instructions)
		interpreter.InitArgs(cobraArgs)
		interpreter.Run()

		if shellFlag {
			env := interpreter.env
			for varName, val := range env.Vars {
				// todo handle different data types specifically
				// todo avoid *dangerous* exports like PATH!!
				RP.PrintForShellEval(fmt.Sprintf("export %s=\"%v\"\n", varName, val))
			}
		}
	}

	for _, arg := range scriptMetadata.Args {
		cobraArg := CreateCobraArg(cmd, arg)
		cobraArgs = append(cobraArgs, &cobraArg)
	}

	// hide global flags, that distract from the particular script
	hideGlobalFlags(cmd)
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

		if RP == nil {
			RP = NewPrinter(cmd, shellFlag, quietFlag, debugFlag, radDebugFlag)
		}

		// try to detect if help has been called on either a script or with --STDIN flag
		if len(args) >= 2 {
			if lo.Some(args[1:], []string{"-h", "--help"}) && stdinScriptName == "" {
				// it has, and with a rsl file source, so let's modify the cmd and re-run the root again
				SetScriptPath(stdinScriptName)
				rslSourceCode := readSource(ScriptPath)
				extractMetadataAndModifyCmd(cmd, rslSourceCode)
			} else if stdinScriptName != "" {
				// it has, and with reading rsl from stdin, so let's modify the cmd and re-run the root again
				source, err := io.ReadAll(RIo.StdIn)
				if err == nil {
					extractMetadataAndModifyCmd(cmd, string(source))
				} else {
					RP.RadErrorExit(fmt.Sprintf("Could not read from stdin: %v\n", err))
				}
			}
		}

		cmd.Help()

		if shellFlag && stdinScriptName != "" {
			// if both these flags are set, we're likely being invoked from within a bash script, so let's
			// output an exit 0 for bash to eval and exit, so it doesn't continue
			RP.PrintForShellEval("exit 0")
		}
	})

	defineGlobalFlags(cmd)
	cmd.SetOut(RIo.StdErr)
}

func Execute() {
	rootCmd := NewRootCmd(CmdInput{})
	InitCmd(rootCmd)
	err := rootCmd.Execute()
	if err != nil {
		RExit(1)
	}
}
