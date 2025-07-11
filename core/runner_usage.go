package core

import (
	"bytes"
	"fmt"
	com "rad/core/common"
	"strings"
)

func (r *RadRunner) RunUsage(shortHelp, isErr bool) {
	if r.scriptData == nil {
		r.printScriptlessUsage(isErr)
	} else {
		r.printScriptUsage(shortHelp, isErr)
	}
}

func (r *RadRunner) RunUsageExit(shortHelp bool) {
	r.RunUsage(shortHelp, false)
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

	com.GreenBoldF(buf, "Usage:\n")
	com.BoldF(buf, "  rad")
	com.CyanF(buf, " [script path | command] [flags]\n\n")

	com.GreenBoldF(buf, "Commands:\n")
	commandUsage(buf, Cmds)

	com.GreenBoldF(buf, "Global options:\n")
	flagUsage(buf, r.globalFlags)

	basicTips(buf)

	r.printHelpFromBuffer(buf, isErr)
}

func (r *RadRunner) printScriptUsage(shortHelp, isErr bool) {
	buf := new(bytes.Buffer)

	if r.scriptData.Description != nil {
		fmt.Fprintf(buf, *r.scriptData.Description+"\n\n")
	}

	com.GreenBoldF(buf, "Usage:\n ")
	if !com.IsBlank(r.scriptData.ScriptName) {
		com.BoldF(buf, fmt.Sprintf(" %s", r.scriptData.ScriptName))
	}

	// separate out positionals from options, to print positionals first
	scriptArgs := make([]RadArg, 0)
	scriptOptions := make([]RadArg, 0)

	for _, arg := range r.scriptArgs {
		if arg.GetType() == ArgBoolT {
			// booleans are not positionally available
			scriptOptions = append(scriptOptions, arg)
			continue
		}
		scriptArgs = append(scriptArgs, arg)

		if arg.IsOptional() {
			com.CyanF(buf, fmt.Sprintf(" [%s]", arg.GetExternalName()))
		} else {
			com.CyanF(buf, fmt.Sprintf(" <%s>", arg.GetExternalName()))
		}
	}

	if !(r.scriptData.DisableGlobalOpts && len(scriptOptions) == 0) {
		fmt.Fprintf(buf, " [OPTIONS]")
	}

	fmt.Fprintf(buf, "\n")

	if len(r.scriptArgs) > 0 {
		fmt.Fprintf(buf, "\n")
		com.GreenBoldF(buf, "Script args:\n")
		flagUsage(buf, append(scriptArgs, scriptOptions...))
	}

	if !shortHelp && !r.scriptData.DisableGlobalOpts {
		fmt.Fprintf(buf, "\n")
		com.GreenBoldF(buf, "Global options:\n")
		flagUsage(buf, r.globalFlags)
	}

	r.printHelpFromBuffer(buf, isErr)
}

// does not handle gracefully/adjusting for cutting down lines if not enough width in terminal
func flagUsage(buf *bytes.Buffer, flags []RadArg) {
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
		if f.HasNonZeroDefault() { // todo should we just always print default if it has one?
			line += fmt.Sprintf(" (default %s)", f.DefaultAsString())
		}

		lines = append(lines, line)
	}

	for _, line := range lines {
		sidx := strings.Index(line, USAGE_ALIGNMENT_CHAR)
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		fmt.Fprintln(
			buf,
			line[:sidx],
			spacing,
			strings.Replace(line[sidx+1:], "\n", "\n"+strings.Repeat(" ", maxlen+2), -1),
		)
	}
}

func commandUsage(buf *bytes.Buffer, cmds []EmbeddedCmd) {
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
	sb.WriteString("To execute a Rad script:\n")
	sb.WriteString("  rad path/to/script.rad [args]\n")
	sb.WriteString("\n")
	sb.WriteString("To execute a command:\n")
	sb.WriteString("  rad <command> [args]\n")
	sb.WriteString("\n")
	sb.WriteString(
		"If you're new, check out the Getting Started guide: https://amterp.github.io/rad/guide/getting-started/\n",
	)

	fmt.Fprintf(buf, sb.String())
}

func (r *RadRunner) printHelpFromBuffer(buf *bytes.Buffer, isErr bool) {
	ioWriter := RIo.StdOut
	if FlagShell.Value || isErr {
		ioWriter = RIo.StdErr
	}
	fmt.Fprintf(ioWriter, buf.String())
}
