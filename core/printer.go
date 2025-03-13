package core

import (
	"fmt"
	"io"
	com "rad/core/common"
	"runtime/debug"
	"strings"

	"github.com/fatih/color"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/amterp/rts"
)

type CodeCtx struct {
	Src string // The whole source code script.
	// Human-friendly i.e. 1-indexed
	RowStart int
	RowEnd   int // inclusive
	ColStart int
	ColEnd   int // inclusive
}

// TODO require RslError code?
// TODO perhaps include a map[node]string for multiple explanations for different parts of the code?
type ErrorCtx struct {
	CodeCtx
	OneLiner string
	Details  string
}

func NewCtx(src string, node *ts.Node, oneLiner string, details string) ErrorCtx {
	return ErrorCtx{
		CodeCtx: CodeCtx{
			Src:      src,
			RowStart: int(node.Range().StartPoint.Row) + 1,
			RowEnd:   int(node.Range().EndPoint.Row) + 1,
			ColStart: int(node.Range().StartPoint.Column) + 1,
			ColEnd:   int(node.Range().EndPoint.Column) + 1,
		},
		OneLiner: oneLiner,
		Details:  details,
	}
}

func NewCtxFromRtsNode(node rts.Node, oneLiner string) ErrorCtx {
	return ErrorCtx{
		CodeCtx: CodeCtx{
			Src:      node.CompleteSrc(),
			RowStart: node.StartPos().Row + 1,
			RowEnd:   node.EndPos().Row + 1,
			ColStart: node.StartPos().Col + 1,
			ColEnd:   node.EndPos().Col + 1,
		},
		OneLiner: oneLiner,
	}
}

// todo make global instance, rather than passing into everything
// For all output to the user, except perhaps pflag-handled help/parsing errors.
type Printer interface {
	// For RSL writers to debug their scripts. They input their debug logs with debug(). Enabled with --DEBUG.
	ScriptDebug(msg string)

	// For Rad (tool) developers to debug the Rad tool, not RSL scripts. Enabled with --RAD-DEBUG.
	// RSL writers should generally not need to use this.
	RadDebugf(format string, args ...interface{})

	// For regular output to the script user
	Print(msg string)

	// For secondary output to the user from Rad, usually to give some feedback, for example querying a URL.
	// Goes to stderr.
	RadInfo(msg string)

	// For output that will be evaluated by the shell, used by --SHELL.
	PrintForShellEval(msg string)

	// For errors related to running the RSL script, with no token available for context.
	// Exits.
	ErrorExit(msg string)

	// Like ErrorExit but takes an error code.
	ErrorExitCode(msg string, errorCode int)

	// TODO
	CtxErrorExit(ctx ErrorCtx)

	// TODO
	CtxErrorCodeExit(ctx ErrorCtx, errorCode int)

	ErrorCodeExitf(errorCode int, msgFmt string, args ...interface{})

	// For errors not related to the RSL script, but to rad itself and its usage (probably misuse or rad bugs).
	// Exits.
	RadErrorExit(msg string)

	// Similar to RadErrorExit, but prints usage after errors, and before exiting.
	UsageErrorExit(msg string)

	// Returns the appropriate writer for regular/standard (non-error) output.
	GetStdWriter() io.Writer
}

// isShellMode directs all regular output to stderr, to avoid interfering with shell evals
//
// isQuiet
// suppresses all output except shell eval prints and rad usage errors, unless isDebug is true, in which
// case it will also print rsl errors stdout, and all debug messages
//
// isScriptDebug will enable script debug messages
// isRadDebug will enable rad debug messages, and include stack traces for errors
func NewPrinter(runner *RadRunner, isShellMode bool, isQuiet bool, isScriptDebug bool, isRadDebug bool) Printer {
	return &stdPrinter{
		stdOut:        RIo.StdOut,
		stdErr:        RIo.StdErr,
		runner:        runner,
		isShellMode:   isShellMode,
		isQuiet:       isQuiet,
		isScriptDebug: isScriptDebug,
		isRadDebug:    isRadDebug,
	}
}

type stdPrinter struct {
	stdOut        io.Writer
	stdErr        io.Writer
	runner        *RadRunner
	isShellMode   bool
	isQuiet       bool
	isScriptDebug bool
	isRadDebug    bool
}

func (p *stdPrinter) ScriptDebug(msg string) {
	if p.isScriptDebug {
		if p.isShellMode {
			fmt.Fprintf(p.stdErr, "DEBUG: %s", msg)
		} else {
			fmt.Fprintf(p.stdOut, "DEBUG: %s", msg)
		}
	}
}

func (p *stdPrinter) RadDebugf(format string, args ...interface{}) {
	if p.isRadDebug {
		if p.isShellMode {
			fmt.Fprintf(p.stdErr, "RAD DEBUG: %s\n", fmt.Sprintf(format, args...))
		} else {
			fmt.Fprintf(p.stdOut, "RAD DEBUG: %s\n", fmt.Sprintf(format, args...))
		}
	}
}

func (p *stdPrinter) Print(msg string) {
	if p.isQuiet {
		return
	} else if p.isShellMode {
		fmt.Fprint(p.stdErr, msg)
	} else {
		fmt.Fprint(p.stdOut, msg)
	}
}

func (p *stdPrinter) RadInfo(msg string) {
	if p.isQuiet {
		return
	}
	fmt.Fprint(p.stdErr, msg)
}

func (p *stdPrinter) PrintForShellEval(msg string) {
	fmt.Fprint(p.stdOut, msg)
}

func (p *stdPrinter) ErrorExit(msg string) {
	p.ErrorExitCode(msg, 1)
}

func (p *stdPrinter) ErrorExitCode(msg string, errorCode int) {
	if !com.IsBlank(msg) && (!p.isQuiet || p.isScriptDebug) {
		fmt.Fprint(p.stdErr, msg)
	}
	p.printShellExitIfEnabled()
	p.errorExit(errorCode)
}

func (p *stdPrinter) CtxErrorExit(ctx ErrorCtx) {
	p.CtxErrorCodeExit(ctx, 1)
}

// todo
//   - print two preceding lines for additional context? if NO, then let's trim the line we're printing
//   - include line numbers in msg
//   - properly point when using tab indentation (currently doesn't)
func (p *stdPrinter) CtxErrorCodeExit(ctx ErrorCtx, errorCode int) {
	if !p.isQuiet || p.isScriptDebug {
		fmt.Fprint(p.stdErr, color.YellowString(fmt.Sprintf("Error at L%d:%d\n\n",
			ctx.RowStart, ctx.ColStart)))
		lines := strings.Split(ctx.Src, "\n")
		relevantLine := lines[ctx.RowStart-1]
		errorLen := ctx.ColEnd - ctx.ColStart
		if ctx.RowStart != ctx.RowEnd {
			errorLen = 1
		}
		fmt.Fprintf(p.stdErr, "  %s\n", relevantLine)
		errorStartIndent := strings.Repeat(" ", ctx.ColStart-1)

		fmt.Fprintf(p.stdErr, "  %s%s", errorStartIndent, color.RedString(strings.Repeat("^", errorLen)))

		errStartIdx := com.StrLen(errorStartIndent) + errorLen
		if !com.IsBlank(ctx.OneLiner) {
			redOneLiner := color.RedString(ctx.OneLiner)
			if com.StrLen(ctx.OneLiner)+errStartIdx < 80 {
				// print next to pointing
				fmt.Fprintf(p.stdErr, " %s\n", redOneLiner)
				// todo some logic to print on the *left* of the arrow if space (but not space on right side)
			} else {
				// print below
				fmt.Fprintf(p.stdErr, "\n  %s%s\n", errorStartIndent, redOneLiner)
			}
		}

		if !com.IsBlank(ctx.Details) {
			fmt.Fprintf(p.stdErr, "\n%s\n", ctx.Details)
		}
	}
	p.printShellExitIfEnabled()
	p.errorExit(errorCode)
}

func (p *stdPrinter) ErrorCodeExitf(errorCode int, msgFmt string, args ...interface{}) {
	if !p.isQuiet || p.isScriptDebug {
		if !com.IsBlank(msgFmt) && !strings.HasSuffix(msgFmt, "\n") {
			msgFmt += "\n"
		}
		fmt.Fprintf(p.stdErr, msgFmt, args...)
	}
	p.printShellExitIfEnabled()
	p.errorExit(errorCode)
}

func (p *stdPrinter) RadErrorExit(msg string) {
	fmt.Fprint(p.stdErr, msg)
	p.printShellExitIfEnabled()
	p.errorExit(1)
}

func (p *stdPrinter) UsageErrorExit(msg string) {
	fmt.Fprint(p.stdErr, msg+"\n\n")
	p.runner.RunUsage(true)
	p.printShellExitIfEnabled()
	p.errorExit(1)
}

func (p *stdPrinter) GetStdWriter() io.Writer {
	if p.isQuiet {
		return NullWriter{}
	}
	if p.isShellMode {
		return p.stdErr
	}
	return p.stdOut
}

func (p *stdPrinter) printShellExitIfEnabled() {
	if p.isShellMode {
		fmt.Println("exit 1")
	}
}

func (p *stdPrinter) errorExit(errorCode int) {
	if p.isRadDebug {
		fmt.Fprintf(p.stdErr, "Stacktrace because --RAD-DEBUG is enabled:\n%s", debug.Stack())
	}
	RExit(errorCode)
}

type NullWriter struct{}

// Write implements the io.Writer interface for NullWriter.
// It accepts data but does nothing with it, effectively discarding it.
func (nw NullWriter) Write(p []byte) (int, error) {
	// Discard the data by doing nothing and returning the length of the data.
	return len(p), nil
}
