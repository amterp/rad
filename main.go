package main

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"os"
	"rad/core"
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

func main() {
	helpFlag := core.NewBoolRadFlag("help", "h", "Print usage string.", false)
	helpFlag.Register()

	debugFlag := core.NewBoolRadFlag("DEBUG", "D", "Enables debug output. Intended for RSL script developers.", false)
	debugFlag.Register()

	stdinFlag := core.NewStringRadFlag("STDIN", "", "script-name", "Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name'.", "")
	stdinFlag.Register()

	mockRespFlag := core.NewMockResponseRadFlag("MOCK-RESPONSE", "", "Add mock response for json requests (pattern:filePath)")
	mockRespFlag.Register()

	globalFlags := []core.RadFlag{&debugFlag, &stdinFlag, &mockRespFlag, &helpFlag}

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

	fmt.Println(debugFlag.Value)
	fmt.Println(stdinFlag.Value)
	fmt.Println(mockRespFlag.Value)
}

// does not handle gracefully/adjusting for cutting down lines if not enough width in terminal
func flagUsage(buf *bytes.Buffer, flags []core.RadFlag) {
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
