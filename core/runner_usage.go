package core

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	com "github.com/amterp/rad/core/common"
)

func (r *RadRunner) RunUsage(shortHelp, isErr bool) {
	if r.scriptData == nil {
		r.printScriptlessUsage(isErr)
	} else {
		r.printScriptUsage(shortHelp, isErr)
	}
}

func (r *RadRunner) printScriptlessUsage(isErr bool) {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "rad: A tool for writing user-friendly command line scripts.\n")
	fmt.Fprintf(buf, "GitHub: https://github.com/amterp/rad\n")
	fmt.Fprintf(buf, "Documentation: https://amterp.dev/rad/\n\n")

	com.GreenBoldF(buf, "Usage:\n")
	com.BoldF(buf, "  rad")
	com.CyanF(buf, " [script path | command] [flags]\n\n")

	com.GreenBoldF(buf, "Commands:\n")
	commandUsage(buf, Cmds, completionDescription)

	// Use Ra's GenerateLongGlobalOptionsSection for superior flag formatting
	buf.WriteString("\n")
	globalOptionsContent := RRootCmd.GenerateLongGlobalOptionsSection()
	buf.WriteString(globalOptionsContent)

	basicTips(buf)

	r.printHelpFromBuffer(buf, isErr)
}

func (r *RadRunner) printScriptUsage(shortHelp, isErr bool) {
	// Delegate to Ra for consistent help formatting
	usageText := RRootCmd.GenerateUsage(!shortHelp)

	buf := new(bytes.Buffer)
	buf.WriteString(usageText)
	r.printHelpFromBuffer(buf, isErr)
}

func commandUsage(buf *bytes.Buffer, cmds []EmbeddedCmd, completionDesc string) {
	type cmdEntry struct {
		name string
		desc string
	}

	// Collect all commands (embedded + Go-implemented) and sort alphabetically
	entries := make([]cmdEntry, 0, len(cmds)+1)
	for _, cmd := range cmds {
		entries = append(entries, cmdEntry{cmd.Name, cmd.Description})
	}
	entries = append(entries, cmdEntry{"completion", completionDesc})
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString("  ")
		sb.WriteString(fmt.Sprintf("%-12s", e.name))
		sb.WriteString("  ")
		sb.WriteString(e.desc + "\n")
	}

	sb.WriteString("\nTo see help for a specific command, run `rad <command> -h`.")

	buf.WriteString(sb.String())
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
		"If you're new, check out the Getting Started guide: https://amterp.dev/rad/guide/getting-started/\n",
	)

	buf.WriteString(sb.String())
}

func (r *RadRunner) printHelpFromBuffer(buf *bytes.Buffer, isErr bool) {
	ioWriter := RIo.StdOut
	if FlagShell.Value || isErr {
		ioWriter = RIo.StdErr
	}
	fmt.Fprint(ioWriter, buf.String())
}
