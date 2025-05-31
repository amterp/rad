package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	com "rad/core/common"
	"strings"

	"github.com/samber/lo"

	"github.com/amterp/rad/rts"

	"github.com/amterp/color"

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
	rawArgs := os.Args[1:]

	// before we can parse any flags, we need to read the script ASAP so we can shadow global flags, then
	// parse flags with pflag, so that we can set up globals like printer, debug logger, etc.

	scriptPath := ""
	sourceCode := ""
	errMsg := ""

	if len(rawArgs) > 0 {
		scriptPath = rawArgs[0]

		if scriptPath == "-" {
			// remove the '-' from the args so that pflag doesn't try to parse it as a flag
			os.Args = append([]string{os.Args[0]}, rawArgs[1:]...)

			// reading script from stdin has been requested
			if RIo.StdIn.HasContent() {
				source, err := io.ReadAll(RIo.StdIn)
				if err != nil {
					errMsg = fmt.Sprintf("Could not read from stdin: %v\n", err)
				} else {
					sourceCode = string(source)
					scriptPath = ""
				}
			} else {
				errMsg = "Requested reading from stdin ('-'), but found no input"
			}
		} else if com.FileExists(scriptPath) {
			// there's a file, read its code
			source, err := readSource(scriptPath)
			if err != nil {
				errMsg = fmt.Sprintf("Could not read script: %v\n", err)
			} else {
				sourceCode = source
			}
		} else if !strings.HasPrefix(scriptPath, "-") {
			// no file, but also not a flag, maybe a command?
			cmdSource := GetEmbeddedCommandSrc(scriptPath)
			if cmdSource != nil {
				AddInternalFuncs()
				sourceCode = *cmdSource
			} else {
				// was not a file, not a flag, not a command, so error
				errMsg = fmt.Sprintf("Unknown file or command: %s", scriptPath)
			}
		}
	}

	// set up a best-effort printer temporarily. May get recreated with global flags later.
	RP = NewPrinter(r, false, false, false, false)

	if !com.IsBlank(errMsg) {
		RP.ErrorExit(errMsg)
	}

	HasScript = !com.IsBlank(sourceCode)
	SetScriptPath(scriptPath)

	// three outcomes so far:
	// 1. errMsg is populated with an error (we won't have a script)
	// 2. sourceCode is populated with a script, no error
	// 3. both sourceCode and errMsg are empty, meaning no script and no error, so print usage

	if HasScript {
		r.scriptData = ExtractMetadata(sourceCode)
	}

	if HasScript {
		// non-blank source implies no error, let's try parsing it so we can remove shadowed global flags
		rslParser, err := rts.NewRslParser()
		if err != nil {
			RP.ErrorExit(fmt.Sprintf("Failed to load RSL parser: %v", err))
		}
		tree := rslParser.Parse(sourceCode)
		argBlock, ok := tree.FindArgBlock()
		if ok {
			if r.scriptData != nil && r.scriptData.DisableArgsBlock {
				RP.ErrorExit(fmt.Sprintf("Macro '%s' disabled, but args block found.\n", MACRO_ENABLE_ARGS_BLOCK))
			}

			for _, argDecl := range argBlock.Args {
				FlagsUsedInScript = append(FlagsUsedInScript, argDecl.ExternalName())

				shorthand := argDecl.ShorthandStr()
				if shorthand != nil {
					FlagsUsedInScript = append(FlagsUsedInScript, *shorthand)
				}
			}
		}
	}

	r.setUpGlobals()
	args := RFlagSet.Args()

	scriptArgs := r.createRslArgsFromScript()
	r.scriptArgs = scriptArgs

	// determine if we should run help/version or not

	if FlagHelp.Value {
		shortHelp := !lo.Contains(os.Args[1:], "--help")
		r.RunUsageExit(shortHelp)
	}

	if FlagVersion.Value {
		RP.Printf(fmt.Sprintf("rad version %s\n", Version))
		RExit(0)
	}

	if com.IsBlank(sourceCode) {
		// re-enable erroring on unknown flags, so we can check if any unknown global flags were given.
		// seems like a limitation of pflag that you cannot just 'get unknown flags' after the earlier parse
		RFlagSet.ParseErrorsWhitelist.UnknownFlags = false

		err := RFlagSet.Parse(os.Args[1:])
		if err != nil {
			// unknown global flag
			RP.UsageErrorExit(err.Error())
		}

		// no flags, effectively, just print the basic usage
		r.RunUsageExit(false)
	}

	// from now on, assume we have a script name (or command)

	shouldExit := false
	if FlagSrc.Value {
		shouldExit = true
		if FlagVersion.Value {
			RP.Printf("\n")
		}
		if !com.IsBlank(ScriptPath) && com.IsTty {
			// print to stderr, since we wouldn't want to include it in e.g. redirects
			RP.RadInfo(com.YellowS("%s:\n", ScriptPath))
		}
		RP.Printf(r.scriptData.Src + "\n")
	}

	if FlagRslTree.Value {
		shouldExit = true
		if FlagSrc.Value {
			RP.Printf("\n")
		}
		RP.Printf(r.scriptData.Tree.Dump())
	}

	if shouldExit {
		RExit(0)
	}

	r.scriptData.ValidateNoErrors()

	// help not explicitly invoked and script has no errors, so let's try parsing other args and maybe run the script

	if !r.scriptData.DisableArgsBlock || !r.scriptData.DisableGlobalFlags {
		// re-enable erroring on unknown flags. note: maybe remove for 'catchall' args?
		RFlagSet.ParseErrorsWhitelist.UnknownFlags = false
	}

	// technically re-using the flagset is apparently discouraged, but i've yet to see where it goes wrong
	err := RFlagSet.Parse(os.Args[1:])
	if err != nil {
		RP.UsageErrorExit(err.Error())
	}

	posArgsIndex := 0
	if !com.IsBlank(scriptPath) {
		// We're invoked on an actual string path, which will be the first arg. Cut it out.
		args = args[1:]
	}

	atLeastOneFlagConfigured := false
	var missingArgs []RslArg
	for _, scriptArg := range scriptArgs {
		if !scriptArg.Configured() {
			// flag has not been explicitly set by the user
			if posArgsIndex < len(args) {
				// there's a positional arg to fill it
				scriptArg.SetValue(args[posArgsIndex])
				posArgsIndex++
			} else if scriptArg.IsOptional() {
				// there's no positional arg to fill it, but that's okay because it's optional, so continue
				continue
			} else if _, ok := scriptArg.(*BoolRslArg); ok {
				// all bools are implicitly optional and default false, unless explicitly defaulted to true
				// this branch implies it was not defaulted to true
			} else {
				missingArgs = append(missingArgs, scriptArg)
				continue // don't validate constraints if it's missing
			}
		} else {
			// arg was given via flag (not positional)
			atLeastOneFlagConfigured = true
		}

		err := scriptArg.ValidateConstraints()
		if err != nil {
			RP.UsageErrorExit(err.Error())
		}
	}

	// finished with our custom additional parsing

	if len(missingArgs) > 0 && len(args) == 0 && !atLeastOneFlagConfigured {
		// if no args were passed but some are required, treat that as the user not really trying to use the script
		// but instead just asking for help
		r.RunUsageExit(true)
	}

	// error if not all positional args were used
	if posArgsIndex < len(args) {
		RP.UsageErrorExit(fmt.Sprintf("Too many positional arguments. Unused: %v", args[posArgsIndex:]))
	}

	constraintCtx := NewConstraintCtx(scriptArgs)
	for _, scriptArg := range scriptArgs {
		err := scriptArg.ValidateRelationalConstraints(constraintCtx)
		if err != nil {
			RP.UsageErrorExit(fmt.Sprintf("Invalid args: %v", err))
		}
	}

	missingArgs = removeMissingIfExcludedByOtherDefinedArg(missingArgs, scriptArgs)
	if len(missingArgs) > 0 {
		RP.UsageErrorExit(fmt.Sprintf("Missing required arguments: %s", TransformRslArgs(missingArgs, RslArg.GetExternalName)))
	}

	// at this point, we'll assume we've been given a script to run, and we should do that now

	if r.scriptData == nil {
		RP.RadErrorExit("Bug! Script expected by this point, but found none")
	}

	interpreter := NewInterpreter(r.scriptData)
	interpreter.InitBuiltIns()
	interpreter.InitArgs(scriptArgs)
	interpreter.RegisterWithExit()
	interpreter.Run()

	if FlagShell.Value {
		interpreter.env.PrintShellExports()
	}

	RExit(0) // explicit exit to trigger deferred statements
	return nil
}

func (r *RadRunner) setUpGlobals() {
	// don't fail on unknown flags. they may be intended for the script, which we won't have parsed initially
	RFlagSet = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	RFlagSet.ParseErrorsWhitelist.UnknownFlags = true

	RFlagSet.Usage = func() {
		r.RunUsage(false, false)
	}

	if r.scriptData == nil || !r.scriptData.DisableGlobalFlags {
		r.globalFlags = CreateAndRegisterGlobalFlags()
	} else {
		// If we don't define our own help flag, pflag intercepts and runs its own usage.
		// If global flags disabled, that includes help, so we define this throwaway flag to
		// absorb the --help and prevent pflag from doing something undesirable.
		help := false
		RFlagSet.BoolVarP(&help, "help", "h", false, "")
	}

	err := RFlagSet.Parse(os.Args[1:])

	// immediately make use of global flags to control behavior for the rest of the program
	RP = NewPrinter(r, FlagShell.Value, FlagQuiet.Value, FlagDebug.Value, FlagRadDebug.Value)

	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			RExit(0)
		}
		RP.UsageErrorExit(err.Error())
	}

	RP.RadDebugf(fmt.Sprintf("Args passed: %v", RFlagSet.Args()))
	if FlagRadDebug.Value {
		RFlagSet.VisitAll(func(flag *pflag.Flag) {
			RP.RadDebugf(fmt.Sprintf("Flag %s: %v", flag.Name, flag.Value))
		})
	}

	switch FlagColor.Value {
	case COLOR_NEVER:
		color.NoColor = true
	case COLOR_ALWAYS:
		color.NoColor = false
	}

	for _, mockResponse := range FlagMockResponse.Value {
		RReq.AddMockedResponse(mockResponse.Pattern, mockResponse.FilePath)
		RP.RadDebugf(fmt.Sprintf("Mock response added: %q -> %q", mockResponse.Pattern, mockResponse.FilePath))
	}
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

func readSource(scriptPath string) (string, error) {
	source, err := os.ReadFile(scriptPath)
	return string(source), err
}

// an argument is only *missing* for error purposes if it is not excluded by another arg, which *is* defined.
// otherwise, this is just a valid constraint working as expected.
func removeMissingIfExcludedByOtherDefinedArg(missingArgs []RslArg, args []RslArg) []RslArg {
	missingIdentifiers := TransformRslArgs(missingArgs, RslArg.GetIdentifier)

	filteredMissingArgs := make([]RslArg, 0)
	for _, missingArg := range missingArgs {
		isMissing := true
		for _, potentialExcluder := range args {
			if potentialExcluder.Excludes(missingArg) && !lo.Contains(missingIdentifiers, potentialExcluder.GetIdentifier()) {
				isMissing = false
				break
			}
		}
		if isMissing {
			filteredMissingArgs = append(filteredMissingArgs, missingArg)
		}
	}
	return filteredMissingArgs
}
