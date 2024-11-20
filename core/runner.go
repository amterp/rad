package core

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"io"
	"os"
	"strings"
)

var (
	plain     = color.New(color.Reset).FprintfFunc()
	green     = color.New(color.FgGreen).FprintfFunc()
	greenBold = color.New(color.FgGreen, color.Bold).FprintfFunc()
	yellow    = color.New(color.FgYellow).FprintfFunc()
	cyan      = color.New(color.FgCyan).FprintfFunc()
	bold      = color.New(color.Bold).FprintfFunc()
)

type RadRunner struct {
	scriptMetadata *ScriptMetadata
}

func NewRadRunner(cmdInput CmdInput) *RadRunner {
	setGlobals(cmdInput)
	return &RadRunner{}
}

func (r *RadRunner) Run() {
	// don't fail on unknown flags. they may be intended for the script, which we won't have parsed initially
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true

	//initialFlagSet := pflag.NewFlagSet("initial", pflag.ExitOnError)
	globalFlags := RegisterGlobalFlags()

	pflag.Usage = func() {
		buf := new(bytes.Buffer)

		fmt.Fprintf(buf, "A tool for writing user-friendly command line scripts.\n\n")
		greenBold(buf, "Usage:\n")
		bold(buf, "  rad")
		cyan(buf, " [script path] [flags]\n\n")

		greenBold(buf, "Global flags:\n")
		flagUsage(buf, globalFlags)

		fmt.Fprintf(os.Stderr, buf.String())
	}

	pflag.Parse()
	RP = NewPrinter(r, ShellFlag, QuietFlag, FlagDebug.Value, RadDebugFlag)

	// flags parsed by pflag. now let's see if we were given a script to run.

	args := pflag.Args()

	var scriptPath string
	if FlagStdinScriptName.Configured() {
		scriptPath = FlagStdinScriptName.Value
		// todo need to read from stdin
	} else if len(args) > 0 {
		scriptPath = args[0]
	}

	if scriptPath != "" {
		// we've been given a script -- parse through it and extract metadata
		SetScriptPath(scriptPath)
		var rslSourceCode string

		if FlagStdinScriptName.Value == "" {
			rslSourceCode = readSource(ScriptPath)
		} else {
			source, err := io.ReadAll(RIo.StdIn)
			if err == nil {
				rslSourceCode = string(source)
			} else {
				RP.RadErrorExit(fmt.Sprintf("Could not read from stdin: %v\n", err))
			}
		}

		l := NewLexer(rslSourceCode)
		l.Lex()

		p := NewParser(l.Tokens)
		instructions := p.Parse()

		r.scriptMetadata = ExtractMetadata(instructions)
	}

	// determine if we should run help or not

	if FlagHelp.Value {
		r.PrintUsage()
		return
	}

	// help not explicitly invoked, so let's try parsing other args

	scriptFlags := r.createFlagsFromScript()

	// re-enable erroring on unknown flags. note: maybe remove for 'catchall' args?
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true

	pflag.Parse()

	fmt.Printf("scriptFlags: %v\n", scriptFlags)

	// todo finish the below stuff, handle usage

	//posArgsIndex := 0
	//if FlagStdinScriptName.Value == "" {
	//	// We're invoked on an actual string path, which will be the first arg. Cut it out.
	//	args = args[1:]
	//}
	//var missingArgs []string
	//for _, cobraArg := range cobraArgs {
	//	argName := cobraArg.Arg.ApiName
	//	cobraFlag := cmd.Flags().Lookup(argName)
	//	if !cobraFlag.Changed {
	//		// flag has not been explicitly set by the user
	//		if posArgsIndex < len(args) {
	//			// there's a positional arg to fill it
	//			cobraArg.SetValue(args[posArgsIndex])
	//			posArgsIndex++
	//		} else if cobraArg.Arg.IsOptional {
	//			// there's no positional arg to fill it, but that's okay because it's optional, so continue
	//			// but first, fill in the optional's default value if it exists
	//			cobraArg.InitializeOptional()
	//			continue
	//		} else if cobraArg.IsBool() {
	//			// all bools are implicitly optional and default false, unless explicitly defaulted to true
	//			// this branch implies it was not defaulted to true
	//			cobraArg.SetValue("false")
	//		} else {
	//			missingArgs = append(missingArgs, argName)
	//		}
	//	}
	//}
	//
	//if len(missingArgs) > 0 && len(args) == 0 {
	//	cmd.Help()
	//	return
	//}
	//
	//if len(missingArgs) > 0 {
	//	RP.UsageErrorExit(fmt.Sprintf("Missing required arguments: %s\n", missingArgs))
	//}
	//
	//// error if not all positional args were used
	//if posArgsIndex < len(args) {
	//	RP.UsageErrorExit(fmt.Sprintf("Too many positional arguments. Unused: %v\n", args[posArgsIndex:]))
	//}
	//
	//color.NoColor = NoColorFlag
	//interpreter := NewInterpreter(instructions)
	//interpreter.InitArgs(cobraArgs)
	//registerInterpreterWithExit(interpreter)
	//interpreter.Run()
	//
	//if ShellFlag {
	//	env := interpreter.env
	//	env.PrintShellExports()
	//}
	//
	//RExit(0) // explicit exit to trigger deferred statements
}

func (r *RadRunner) PrintUsage() {
}

func (r *RadRunner) createFlagsFromScript() []RadFlag {
	if r.scriptMetadata == nil {
		return nil
	}

	flags := make([]RadFlag, 0, len(r.scriptMetadata.Args))
	for _, arg := range r.scriptMetadata.Args {
		flag := CreateFlag(arg)
		flag.Register()
		flags = append(flags, flag)
	}

	return flags
}

// does not handle gracefully/adjusting for cutting down lines if not enough width in terminal
func flagUsage(buf *bytes.Buffer, flags []RadFlag) {
	lines := make([]string, 0, len(flags))

	maxlen := 0
	for _, f := range flags {
		line := ""
		if f.GetShort() != "" {
			line = fmt.Sprintf("  -%s, --%s", f.GetShort(), f.GetName())
		} else {
			line = fmt.Sprintf("      --%s", f.GetName())
		}

		argUsage := f.GetArgUsage()
		if argUsage != "" {
			line += " " + argUsage
		}

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += f.GetDescription()
		if f.HasNonZeroDefault() {
			line += fmt.Sprintf(" (default %s)", f.DefaultAsString())
		}

		lines = append(lines, line)
	}

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		fmt.Fprintln(buf, line[:sidx], spacing, strings.Replace(line[sidx+1:], "\n", "\n"+strings.Repeat(" ", maxlen+2), -1))
	}
}
