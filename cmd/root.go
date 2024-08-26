package cmd

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"rad/core"
)

var (
	subCommandInitialized bool
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "Request And Display (RAD)",
	Long:  `Request And Display (RAD): A tool for making HTTP requests, extracting details, and displaying the result.`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		if subCommandInitialized {
			return
		}

		addScriptSubCommand(cmd, args)
		cmd.Execute()
	},
}

func addScriptSubCommand(cmd *cobra.Command, args []string) {
	scriptPath := args[0]
	source := readSource(scriptPath)
	l := core.NewLexer(source)
	l.Lex()

	p := core.NewParser(l.Tokens)
	statements := p.Parse()
	for _, stmt := range statements {
		fmt.Printf("%v\n", stmt)
	}

	// todo
	//scriptArgs := extractArgs()
	//scriptCmd := createCmd(scriptArgs)
	//cmd.AddCommand(scriptCmd)
	subCommandInitialized = true
}

func createCmd(args []core.ScriptArg) *cobra.Command {
	scriptName := "sdssym"
	useString := generateUseString(scriptName, args)
	var cobraArgs []*CobraArg
	scriptCmd := &cobra.Command{
		Use:   useString,
		Short: "short description", // todo use file header's short description
		Long:  generateLongDescription(),
		Run: func(cmd *cobra.Command, args []string) {

			// fill in positional args, and
			// error if required args are missing
			posArgsIndex := 0
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
						cobraArg.SetDefaultIfPresent()
						continue
					} else {
						errorExit(cmd, fmt.Sprintf("Missing required argument: %s", argName))
					}
				}
			}
			// error if not all positional args were used
			if posArgsIndex < len(args) {
				errorExit(cmd, fmt.Sprintf("Too many positional arguments. Unused: %v", args[posArgsIndex:]))
			}

			fmt.Println("Flags:")
			for _, arg := range cobraArgs {
				switch {
				case arg.IsString():
					fmt.Printf("%s: %s\n", arg.Arg.Name, arg.GetString())
				case arg.IsStringArray():
					fmt.Printf("%s: %v\n", arg.Arg.Name, arg.GetStringArray())
				case arg.IsInt():
					fmt.Printf("%s: %d\n", arg.Arg.Name, arg.GetInt())
				case arg.IsIntArray():
					fmt.Printf("%s: %v\n", arg.Arg.Name, arg.GetIntArray())
				case arg.IsBool():
					fmt.Printf("%s: %t\n", arg.Arg.Name, arg.GetBool())
				}
			}
		},
	}
	for _, arg := range args {
		name, argType, flag, description := arg.Name, arg.Type, "", ""
		if arg.Flag != nil {
			flag = *arg.Flag
		}
		if arg.Description != nil {
			description = *arg.Description
		}

		var cobraArgValue interface{}
		switch argType {
		case core.RslString:
			cobraArgValue = scriptCmd.Flags().StringP(name, flag, "", description)
		case core.RslStringArray:
			cobraArgValue = scriptCmd.Flags().StringSliceP(name, flag, []string{}, description)
		case core.RslInt:
			cobraArgValue = scriptCmd.Flags().IntP(name, flag, 0, description)
		case core.RslIntArray:
			cobraArgValue = scriptCmd.Flags().IntSliceP(name, flag, []int{}, description)
		case core.RslBool:
			cobraArgValue = scriptCmd.Flags().BoolP(name, flag, false, description)
		default:
			// todo better error handling
			panic(fmt.Sprintf("Unknown arg type: %v", argType))
		}
		cobraArgs = append(cobraArgs, &CobraArg{Arg: arg, value: cobraArgValue})
	}
	return scriptCmd
}

func generateUseString(scriptName string, args []core.ScriptArg) string {
	useString := scriptName
	for _, arg := range args {
		if arg.IsOptional {
			useString += fmt.Sprintf(" [%s]", arg.Name)
		} else {
			useString += fmt.Sprintf(" <%s>", arg.Name)
		}
	}
	return useString
}

func generateLongDescription() string {
	// todo use file header's long description
	return "loooooong description"
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
	// this is a little crazy, bear with me!
	// we use a hack required to allow help flags intended for subcommands to correctly
	// apply to the subcommand. We need to create & register the subcommand with
	// cobra before *cobra* is able to properly understand that the help flag is
	// intended for the subcommand.
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// immediately reset the help func, as we only want this hacked
		// version to run once
		rootCmd.SetHelpFunc(nil)

		// try to detect if help has been called on a subcommand
		if len(args) >= 2 {
			if lo.Some(args[1:], []string{"-h", "--help"}) {
				// it has! so let's add the subcommand and re-run the root again
				addScriptSubCommand(rootCmd, args)
				rootCmd.Execute()
				return
			}
		}

		// does not look like we are trying to get help on a subcommand,
		// so just print the root/normal help
		rootCmd.Help()
	})
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
