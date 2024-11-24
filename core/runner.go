package core

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"io"
	"os"
)

type RadRunner struct {
	scriptMetadata *ScriptData
	globalFlags    []RslArg
	scriptArgs     []RslArg
}

func NewRadRunner(runnerInput RunnerInput) *RadRunner {
	setGlobals(runnerInput)
	return &RadRunner{}
}

func (r *RadRunner) Run() error {
	// don't fail on unknown flags. they may be intended for the script, which we won't have parsed initially
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true

	pflag.Usage = func() {
		r.RunUsageExit()
	}

	r.globalFlags = CreateAndRegisterGlobalFlags()

	pflag.Parse()

	// immediately make use of global flags to control behavior for the rest of the program

	RP = NewPrinter(r, FlagShell.Value, FlagQuiet.Value, FlagDebug.Value, FlagRadDebug.Value)

	RP.RadDebug(fmt.Sprintf("Args passed: %v", pflag.Args()))
	if FlagRadDebug.Value {
		pflag.VisitAll(func(flag *pflag.Flag) {
			RP.RadDebug(fmt.Sprintf("Flag %s: %v", flag.Name, flag.Value))
		})
	}

	color.NoColor = FlagNoColor.Value
	for _, mockResponse := range FlagMockResponse.Value {
		RReq.AddMockedResponse(mockResponse.Pattern, mockResponse.FilePath)
		RP.RadDebug(fmt.Sprintf("Mock response added: %q -> %q", mockResponse.Pattern, mockResponse.FilePath))
	}

	// now let's see if we were given a script to run.

	args := pflag.Args()

	var scriptName string
	if FlagStdinScriptName.Configured() {
		scriptName = FlagStdinScriptName.Value
	} else if len(args) > 0 {
		scriptName = args[0]
	}

	if scriptName != "" || FlagStdinScriptName.Configured() {
		// we've been given a script either via file or stdin -- parse through it and extract metadata
		SetScriptPath(scriptName)
		var rslSourceCode string

		if FlagStdinScriptName.Configured() {
			// script is given via stdin
			source, err := io.ReadAll(RIo.StdIn)
			if err == nil {
				rslSourceCode = string(source)
			} else {
				RP.RadErrorExit(fmt.Sprintf("Could not read from stdin: %v\n", err))
			}
		} else {
			// script is given via file
			rslSourceCode = readSource(ScriptPath)
		}

		l := NewLexer(rslSourceCode)
		l.Lex()

		p := NewParser(l.Tokens)
		instructions := p.Parse()

		r.scriptMetadata = ExtractMetadata(instructions)
	}

	scriptArgs := r.createRslArgsFromScript()
	r.scriptArgs = scriptArgs

	// determine if we should run help/version or not

	if FlagHelp.Value {
		r.RunUsageExit()
	} else if FlagVersion.Value {
		RP.RadInfo(fmt.Sprintf("rad version %s\n", Version))
		RExit(0)
	}

	// help not explicitly invoked, so let's try parsing other args

	// re-enable erroring on unknown flags. note: maybe remove for 'catchall' args?
	// todo if unknown flag passed, pflag handles the error & prints a kinda ugly msg (twice, bug).
	//  continue allowing unknown flags and then detect ourselves?
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = false

	// todo apparently this is not recommended, I should be using flagsets? I THINK I DO, FOR TESTS?
	pflag.Parse()

	posArgsIndex := 0
	if FlagStdinScriptName.Value == "" {
		// We're invoked on an actual string path, which will be the first arg. Cut it out.
		args = args[1:]
	}
	var missingArgs []string
	for _, scriptArg := range scriptArgs {
		argName := scriptArg.GetName()
		if !scriptArg.Configured() {
			// flag has not been explicitly set by the user
			if posArgsIndex < len(args) {
				// there's a positional arg to fill it
				scriptArg.SetValue(args[posArgsIndex])
				posArgsIndex++
			} else if scriptArg.IsOptional() {
				// there's no positional arg to fill it, but that's okay because it's optional, so continue
				// but first, fill in the optional's default value if it exists
				//scriptArg.InitializeOptional() // todo this is currently already done i think by pflag
				continue
				//} else if _, ok := scriptArg.(*BoolRslArg); ok {
				//	// all bools are implicitly optional and default false, unless explicitly defaulted to true
				//	// this branch implies it was not defaulted to true // todo i think also not needed ditto above
				//	scriptArg.SetValue("false")
			} else {
				missingArgs = append(missingArgs, argName)
			}
		}
	}

	// finished with our custom additional parsing

	if len(missingArgs) > 0 && len(args) == 0 {
		// if no args were passed but some are required, treat that as the user not really trying to use the script
		// but instead just asking for help
		r.RunUsageExit()
	}

	if len(missingArgs) > 0 {
		RP.UsageErrorExit(fmt.Sprintf("Missing required arguments: %s\n", missingArgs))
	}

	// error if not all positional args were used
	if posArgsIndex < len(args) {
		RP.UsageErrorExit(fmt.Sprintf("Too many positional arguments. Unused: %v\n", args[posArgsIndex:]))
	}

	// at this point, we'll assume we've been given a script to run, and we should do that now

	if r.scriptMetadata == nil {
		RP.RadErrorExit("Bug! Script expected by this point, but found none")
	}

	interpreter := NewInterpreter(r.scriptMetadata.Instructions)
	interpreter.InitArgs(scriptArgs)
	registerInterpreterWithExit(interpreter)
	interpreter.Run()

	if FlagShell.Value {
		env := interpreter.env
		env.PrintShellExports()
	}

	RExit(0) // explicit exit to trigger deferred statements
	return nil
}

func (r *RadRunner) createRslArgsFromScript() []RslArg {
	if r.scriptMetadata == nil {
		return nil
	}

	flags := make([]RslArg, 0, len(r.scriptMetadata.Args))
	for _, arg := range r.scriptMetadata.Args {
		flag := CreateFlag(arg)
		flag.Register()
		flags = append(flags, flag)
	}

	return flags
}
func readSource(scriptPath string) string {
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		RP.RadErrorExit(fmt.Sprintf("Could not read script '%s': %v\n", scriptPath, err))
		RExit(1)
	}
	return string(source)
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
