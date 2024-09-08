package core

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var (
	rootModified bool
	bashFlag     bool
	stdinFlag    bool
)

var rootCmd = &cobra.Command{
	Use:     "",
	Short:   "Request And Display (RAD)",
	Long:    `Request And Display (RAD): A tool for making HTTP requests, extracting details, and displaying the result.`,
	Version: "0.1.5",
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		scriptPath := ""
		var rslSourceCode string
		if stdinFlag {
			source, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not read from stdin: %v\n", err)
				os.Exit(1)
			}
			rslSourceCode = string(source)
		} else if len(args) == 0 {
			cmd.Help()
			return
		} else {
			scriptPath = args[0]
			rslSourceCode = readSource(scriptPath)
		}

		if rootModified {
			return
		}

		extractMetadataAndModifyCmd(cmd, scriptPath, rslSourceCode)
		cmd.Execute()
	},
}

func extractMetadataAndModifyCmd(cmd *cobra.Command, rslSourcePath string, rslSourceCode string) {
	l := NewLexer(rslSourceCode)
	l.Lex()

	p := NewParser(l.Tokens)
	instructions := p.Parse()

	scriptMetadata := ExtractMetadata(instructions)
	modifyCmd(cmd, rslSourcePath, scriptMetadata, instructions)
	rootModified = true
}

func modifyCmd(cmd *cobra.Command, scriptPath string, scriptMetadata ScriptMetadata, instructions []Stmt) {
	useString := GenerateUseString(scriptPath, scriptMetadata.Args)
	var cobraArgs []*CobraArg
	cmd.Use = useString
	cmd.Short = ShortDescription(scriptMetadata)
	cmd.Long = LongDescription(scriptMetadata)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		// fill in positional args, and
		// error if required args are missing
		posArgsIndex := 1
		if stdinFlag {
			posArgsIndex = 0
		}
		for _, cobraArg := range cobraArgs {
			argName := cobraArg.Arg.Name
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
					errorExit(cmd, fmt.Sprintf("Missing required argument: %s", argName))
				}
			}
		}
		// error if not all positional args were used
		if posArgsIndex < len(args) {
			errorExit(cmd, fmt.Sprintf("Too many positional arguments. Unused: %v", args[posArgsIndex:]))
		}

		interpreter := NewInterpreter(instructions)
		interpreter.InitArgs(cobraArgs)
		interpreter.Run()

		if bashFlag {
			env := interpreter.env
			for varName, val := range env.Vars {
				// todo handle different types specifically
				fmt.Printf("export %s=\"%v\"\n", varName, val.value)
			}
		}
	}

	for _, arg := range scriptMetadata.Args {
		cobraArg := CreateCobraArg(cmd, arg)
		cobraArgs = append(cobraArgs, &cobraArg)
	}
}

func readSource(scriptPath string) string {
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read script '%s': %v\n", scriptPath, err)
		os.Exit(1)
	}
	return string(source)
}

func errorExit(cmd *cobra.Command, message string) {
	fmt.Println(message)
	cmd.Usage()
	os.Exit(1)
}

func init() {
	// this is a bit hacky, but bear with me!
	// we intercept the very first call to help, it implies that the user has set the help flag
	// however, we won't yet have read the RSL script if it was also provided, so we may first
	// try to read that, register the args, etc, that's relevant to help, and then re-run the help
	// command, so that it can print help for the script
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// immediately reset the help func, as we only want this hacked
		// version to run once
		rootCmd.SetHelpFunc(nil)

		// try to detect if help has been called on either a script or with --STDIN flag
		if len(args) >= 2 {
			if lo.Some(args[1:], []string{"-h", "--help"}) && !stdinFlag {
				// it has, and with a rsl file source, so let's modify the rootCmd and re-run the root again
				scriptPath := args[0]
				rslSourceCode := readSource(scriptPath)
				extractMetadataAndModifyCmd(rootCmd, scriptPath, rslSourceCode)
			} else if stdinFlag {
				// it has, and with reading rsl from stdin, so let's modify the rootCmd and re-run the root again
				source, err := io.ReadAll(os.Stdin)
				if err == nil {
					extractMetadataAndModifyCmd(rootCmd, "", string(source))
				} else {
					fmt.Fprintf(os.Stderr, "Could not read from stdin: %v\n", err)
				}
			}
		}

		rootCmd.Help()

		if bashFlag && stdinFlag {
			// if both these flags are set, we're likely being invoked from within a bash script, so let's
			// output an exit 0 for bash to eval and exit, so it doesn't continue
			fmt.Println("exit 0")
		}
	})

	// global flags
	// todo think more about bash vs. shell
	// todo these flags should be hidden (probably) when --help called on script
	rootCmd.PersistentFlags().BoolVar(&bashFlag, "BASH", false, "Outputs bash exports of variables, so they can be eval'd")
	rootCmd.PersistentFlags().BoolVar(&stdinFlag, "STDIN", false, "Reads RSL script from stdin")
	rootCmd.SetOut(os.Stderr)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
