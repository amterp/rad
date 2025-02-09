package core

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

type RadRunner struct {
	scriptData  *ScriptData
	globalFlags []RslArg
	scriptArgs  []RslArg
}

func NewRadRunner(runnerInput RunnerInput) *RadRunner {
	setGlobals(runnerInput)
	return &RadRunner{}
}

func (r *RadRunner) Run() error {
	// don't fail on unknown flags. they may be intended for the script, which we won't have parsed initially
	RFlagSet = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	RFlagSet.ParseErrorsWhitelist.UnknownFlags = true

	RFlagSet.Usage = func() {
		r.RunUsage()
	}

	r.globalFlags = CreateAndRegisterGlobalFlags()

	err := RFlagSet.Parse(os.Args[1:])
	if err != nil {
		RP.UsageErrorExit(err.Error())
	}

	// immediately make use of global flags to control behavior for the rest of the program

	RP = NewPrinter(r, FlagShell.Value, FlagQuiet.Value, FlagDebug.Value, FlagRadDebug.Value)

	RP.RadDebugf(fmt.Sprintf("Args passed: %v", RFlagSet.Args()))
	if FlagRadDebug.Value {
		RFlagSet.VisitAll(func(flag *pflag.Flag) {
			RP.RadDebugf(fmt.Sprintf("Flag %s: %v", flag.Name, flag.Value))
		})
	}

	color.NoColor = FlagNoColor.Value
	for _, mockResponse := range FlagMockResponse.Value {
		RReq.AddMockedResponse(mockResponse.Pattern, mockResponse.FilePath)
		RP.RadDebugf(fmt.Sprintf("Mock response added: %q -> %q", mockResponse.Pattern, mockResponse.FilePath))
	}

	// now let's see if we were given a script to run.

	args := RFlagSet.Args()

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

		RP.RadDebugf(fmt.Sprintf("Read src code (%d chars), parsing...", len(rslSourceCode)))
		r.scriptData = ExtractMetadata(rslSourceCode)
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
	RFlagSet.ParseErrorsWhitelist.UnknownFlags = false

	// technically re-using the flagset is apparently discouraged, but i've yet to see where it goes wrong
	err = RFlagSet.Parse(os.Args[1:])
	if err != nil {
		RP.UsageErrorExit(err.Error())
	}

	posArgsIndex := 0
	if FlagStdinScriptName.Value == "" {
		// We're invoked on an actual string path, which will be the first arg. Cut it out.
		args = args[1:]
	}
	var missingArgs []string
	for _, scriptArg := range scriptArgs {
		argName := scriptArg.GetExternalName()
		if !scriptArg.Configured() {
			// flag has not been explicitly set by the user
			if posArgsIndex < len(args) {
				// there's a positional arg to fill it
				scriptArg.SetValue(args[posArgsIndex])
				posArgsIndex++
			} else if scriptArg.IsOptional() {
				// there's no positional arg to fill it, but that's okay because it's optional, so continue
				// but first, fill in the optional's default value if it exists
				//scriptArg.InitializeOptional() // todo this is currently already done i think by pflag, remove?
				continue
			} else if _, ok := scriptArg.(*BoolRslArg); ok {
				// all bools are implicitly optional and default false, unless explicitly defaulted to true
				// this branch implies it was not defaulted to true
			} else {
				missingArgs = append(missingArgs, argName)
			}
		}
		err := scriptArg.ValidateConstraints()
		if err != nil {
			RP.UsageErrorExit(err.Error())
		}
	}

	// finished with our custom additional parsing

	if len(missingArgs) > 0 && len(args) == 0 {
		// if no args were passed but some are required, treat that as the user not really trying to use the script
		// but instead just asking for help
		r.RunUsageExit()
	}

	if len(missingArgs) > 0 {
		RP.UsageErrorExit(fmt.Sprintf("Missing required arguments: %s", missingArgs))
	}

	// error if not all positional args were used
	if posArgsIndex < len(args) {
		RP.UsageErrorExit(fmt.Sprintf("Too many positional arguments. Unused: %v", args[posArgsIndex:]))
	}

	// at this point, we'll assume we've been given a script to run, and we should do that now

	if r.scriptData == nil {
		RP.RadErrorExit("Bug! Script expected by this point, but found none")
	}

	interpreter := NewInterpreter(r.scriptData)
	interpreter.CheckForErrors()
	interpreter.InitArgs(scriptArgs)
	interpreter.RegisterWithExit()
	interpreter.Run()

	//if FlagShell.Value { todo
	//	env := interpreter.env
	//	env.PrintShellExports()
	//}

	RExit(0) // explicit exit to trigger deferred statements
	return nil
}

func (r *RadRunner) createRslArgsFromScript() []RslArg {
	if r.scriptData == nil {
		return nil
	}

	flags := make([]RslArg, 0, len(r.scriptData.Args))
	for _, arg := range r.scriptData.Args {
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
