package core

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
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

func (r *RadRunner) RunUsage() {
	if r.scriptMetadata == nil {
		r.printScriptlessUsage()
	} else {
		r.printScriptUsage()
	}
}

func (r *RadRunner) RunUsageExit() {
	r.RunUsage()
	if FlagShell.Value {
		RP.PrintForShellEval("exit 0")
	}
	RExit(0)
}

func (r *RadRunner) printScriptlessUsage() {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "rad: A tool for writing user-friendly command line scripts.\n\n")
	greenBold(buf, "Usage:\n")
	bold(buf, "  rad")
	cyan(buf, " [script path] [flags]\n\n")

	greenBold(buf, "Global flags:\n")
	flagUsage(buf, r.globalFlags)

	fmt.Fprintf(RIo.StdErr, buf.String())
}

func (r *RadRunner) printScriptUsage() {
	buf := new(bytes.Buffer)

	if r.scriptMetadata.BlockDescription != nil {
		fmt.Fprintf(buf, *r.scriptMetadata.BlockDescription+"\n\n")
	} else if r.scriptMetadata.OneLineDescription != nil {
		fmt.Fprintf(buf, *r.scriptMetadata.OneLineDescription+"\n\n")
	}

	greenBold(buf, "Usage:\n")
	// todo need to prefix with 'rad' if that's how this got invoked.
	bold(buf, fmt.Sprintf("  %s", r.scriptMetadata.ScriptName))

	for _, arg := range r.scriptMetadata.Args {
		if arg.IsOptional {
			cyan(buf, fmt.Sprintf(" [%s]", arg.ApiName))
		} else if arg.Type == ArgBoolT {
			if arg.Short == nil {
				cyan(buf, fmt.Sprintf(" [--%s]", arg.ApiName))
			} else {
				cyan(buf, fmt.Sprintf(" [-%s, --%s]", *arg.Short, arg.ApiName))
			}
		} else {
			cyan(buf, fmt.Sprintf(" <%s>", arg.ApiName))
		}
	}
	fmt.Fprintf(buf, "\n\n")

	greenBold(buf, "Script args:\n")
	flagUsage(buf, r.scriptArgs)

	fmt.Fprintf(buf, "\n")

	if !FlagStdinScriptName.Configured() {
		FlagStdinScriptName.Hidden(true)
	}
	// todo probably don't print these if there's a script? Or only minimal ones if --help is passed (not -h) ?
	greenBold(buf, "Global flags:\n")
	flagUsage(buf, r.globalFlags)

	fmt.Fprintf(RIo.StdErr, buf.String())
}

// does not handle gracefully/adjusting for cutting down lines if not enough width in terminal
func flagUsage(buf *bytes.Buffer, flags []RslArg) {
	lines := make([]string, 0, len(flags))

	maxlen := 0
	for _, f := range flags {
		if f.IsHidden() {
			continue
		}

		line := ""
		if f.GetShort() != "" && f.GetName() != "" {
			line = fmt.Sprintf("  -%s, --%s", f.GetShort(), f.GetName())
		} else if f.GetShort() == "" {
			line = fmt.Sprintf("      --%s", f.GetName())
		} else if f.GetName() == "" {
			line = fmt.Sprintf("  -%s", f.GetShort())
		}

		argUsage := f.GetArgUsage()
		if argUsage != "" {
			line += " " + argUsage
		}

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += USAGE_ALIGNMENT_CHAR
		if StrLen(line) > maxlen {
			maxlen = StrLen(line)
		}

		line += f.GetDescription()
		if f.HasNonZeroDefault() {
			line += fmt.Sprintf(" (default %s)", f.DefaultAsString())
		}

		lines = append(lines, line)
	}

	for _, line := range lines {
		sidx := strings.Index(line, USAGE_ALIGNMENT_CHAR)
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		fmt.Fprintln(buf, line[:sidx], spacing, strings.Replace(line[sidx+1:], "\n", "\n"+strings.Repeat(" ", maxlen+2), -1))
	}
}
