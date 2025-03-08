package core

import (
	"bytes"
	"fmt"
	com "rad/core/common"
	"strings"

	"github.com/fatih/color"
)

var (
	plain     = color.New(color.Reset).FprintfFunc()
	green     = color.New(color.FgGreen).FprintfFunc()
	greenBold = color.New(color.FgGreen, color.Bold).FprintfFunc()
	yellow    = color.New(color.FgYellow).FprintfFunc()
	cyan      = color.New(color.FgCyan).FprintfFunc()
	bold      = color.New(color.Bold).FprintfFunc()
)

func (r *RadRunner) RunUsage(isErr bool) {
	if r.scriptData == nil {
		r.printScriptlessUsage(isErr)
	} else {
		r.printScriptUsage(isErr)
	}
}

func (r *RadRunner) RunUsageExit() {
	r.RunUsage(false)
	if FlagShell.Value {
		RP.PrintForShellEval("exit 0")
	}
	RExit(0)
}

func (r *RadRunner) printScriptlessUsage(isErr bool) {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "rad: A tool for writing user-friendly command line scripts.\n")
	fmt.Fprintf(buf, "GitHub: https://github.com/amterp/rad\n")
	fmt.Fprintf(buf, "Documentation: https://amterp.github.io/rad/\n\n")

	greenBold(buf, "Usage:\n")
	bold(buf, "  rad")
	cyan(buf, " [script path | command] [flags]\n\n")

	greenBold(buf, "Commands:\n")
	commandUsage(buf, CmdsByName)

	greenBold(buf, "Global flags:\n")
	flagUsage(buf, r.globalFlags)

	basicTips(buf)

	r.printHelpFromBuffer(buf, isErr)
}

func (r *RadRunner) printScriptUsage(isErr bool) {
	buf := new(bytes.Buffer)

	if r.scriptData.Description != nil {
		fmt.Fprintf(buf, *r.scriptData.Description+"\n")
	}

	greenBold(buf, "Usage:\n ")
	if !com.IsBlank(r.scriptData.ScriptName) {
		bold(buf, fmt.Sprintf(" %s", r.scriptData.ScriptName))
	}

	for _, arg := range r.scriptData.Args {
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

	// todo probably don't print these if there's a script? Or only minimal ones if --help is passed (not -h) ?
	greenBold(buf, "Global flags:\n")
	flagUsage(buf, r.globalFlags)

	r.printHelpFromBuffer(buf, isErr)
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
		if f.GetShort() != "" && f.GetExternalName() != "" {
			line = fmt.Sprintf("  -%s, --%s", f.GetShort(), f.GetExternalName())
		} else if f.GetShort() == "" {
			line = fmt.Sprintf("      --%s", f.GetExternalName())
		} else if f.GetExternalName() == "" {
			line = fmt.Sprintf("  -%s", f.GetShort())
		}

		argUsage := f.GetArgUsage()
		if argUsage != "" {
			line += " " + argUsage
		}

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += USAGE_ALIGNMENT_CHAR
		if com.StrLen(line) > maxlen {
			maxlen = com.StrLen(line)
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

func commandUsage(buf *bytes.Buffer, cmds map[string]EmbeddedCmd) {
	var sb strings.Builder

	for _, cmd := range cmds {
		sb.WriteString("  ")
		sb.WriteString(fmt.Sprintf("%-12s", cmd.Name))
		sb.WriteString("  ")
		sb.WriteString(cmd.Description + "\n")
	}

	sb.WriteString("\nTo see help for a specific command, run `rad <command> -h`.\n\n")

	fmt.Fprintf(buf, sb.String())
}

func basicTips(buf *bytes.Buffer) {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("To execute an RSL script:\n")
	sb.WriteString("  rad path/to/script.rsl [args]\n")
	sb.WriteString("\n")
	sb.WriteString("To execute a command:\n")
	sb.WriteString("  rad <command> [args]\n")
	sb.WriteString("\n")
	sb.WriteString("If you're new, check out the Getting Started guide: https://amterp.github.io/rad/guide/getting-started/\n")

	fmt.Fprintf(buf, sb.String())
}

func (r *RadRunner) printHelpFromBuffer(buf *bytes.Buffer, isErr bool) {
	ioWriter := RIo.StdOut
	if FlagShell.Value || isErr {
		ioWriter = RIo.StdErr
	}
	fmt.Fprintf(ioWriter, buf.String())
}
