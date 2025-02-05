package core

import (
	"fmt"
	"io"
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

func NewCtx(src string, node *ts.Node) CodeCtx {
	return CodeCtx{
		Src:      src,
		RowStart: int(node.Range().StartPoint.Row) + 1,
		RowEnd:   int(node.Range().EndPoint.Row) + 1,
		ColStart: int(node.Range().StartPoint.Column) + 1,
		ColEnd:   int(node.Range().EndPoint.Column) + 1,
	}
}

func NewCtxFromRtsNode(node rts.Node) CodeCtx {
	return CodeCtx{
		Src:      node.Src(),
		RowStart: node.StartPos().Row + 1,
		RowEnd:   node.EndPos().Row + 1,
		ColStart: node.StartPos().Col + 1,
		ColEnd:   node.EndPos().Col + 1,
	}
}

// todo make global instance, rather than passing into everything
// For all output to the user, except perhaps pflag-handled help/parsing errors.
type Printer interface {
	// For RSL writers to debug their scripts. They input their debug logs with debug(). Enabled with --DEBUG.
	ScriptDebug(msg string)

	// For Rad (tool) developers to debug the Rad tool, not RSL scripts. Enabled with --RAD-DEBUG.
	// RSL writers should generally not need to use this.
	RadDebug(msg string)

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

	// For the parser and interpreter to print errors, with a token.
	// Exits.
	TokenErrorExit(token Token, msg string)

	// Like TokenErrorExit but takes an error code.
	TokenErrorCodeExit(token Token, msg string, errorCode int)

	// TODO
	CtxErrorExit(ctx CodeCtx, msg string)

	// TODO
	CtxErrorCodeExit(ctx CodeCtx, msg string, errorCode int)

	// For errors not related to the RSL script, but to rad itself and its usage (probably misuse or rad bugs).
	// Exits.
	RadErrorExit(msg string)

	// Similar to RadErrorExit, but where a token is available for context.
	RadTokenErrorExit(token Token, msg string)

	// TODO
	RadNodeErrorExit(node rts.Node, msg string)

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
		stdIn:         RIo.StdIn,
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
	stdIn         io.Reader
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

func (p *stdPrinter) RadDebug(msg string) {
	if p.isRadDebug {
		if p.isShellMode {
			fmt.Fprintf(p.stdErr, "RAD DEBUG: %s\n", msg)
		} else {
			fmt.Fprintf(p.stdOut, "RAD DEBUG: %s\n", msg)
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
	if !p.isQuiet || p.isScriptDebug {
		fmt.Fprint(p.stdErr, msg)
	}
	p.printShellExitIfEnabled()
	p.errorExit(errorCode)
}

func (p *stdPrinter) TokenErrorExit(token Token, msg string) {
	p.TokenErrorCodeExit(token, msg, 1)
}

func (p *stdPrinter) TokenErrorCodeExit(token Token, msg string, errorCode int) {
	if !p.isQuiet || p.isScriptDebug {
		if token == nil {
			fmt.Fprint(p.stdErr, msg)
		} else {
			lexeme := token.GetLexeme()
			lexeme = strings.ReplaceAll(lexeme, "\n", "\\n")
			lexeme = strings.ReplaceAll(lexeme, "\t", "\\t")
			fmt.Fprintf(p.stdErr, "RslError at L%d/%d on '%s': %s",
				token.GetLine(), token.GetCharLineStart(), lexeme, msg)
		}
	}
	p.printShellExitIfEnabled()
	p.errorExit(errorCode)
}

func (p *stdPrinter) CtxErrorExit(ctx CodeCtx, msg string) {
	p.CtxErrorCodeExit(ctx, msg, 1)
}

func (p *stdPrinter) CtxErrorCodeExit(ctx CodeCtx, msg string, errorCode int) {
	if !p.isQuiet || p.isScriptDebug {
		// todo do nice src code extraction + print + point
		fmt.Fprint(p.stdErr, color.YellowString(fmt.Sprintf("Error at L%d/%d\n\n",
			ctx.RowStart, ctx.ColStart)))
		lines := strings.Split(ctx.Src, "\n")
		relevantLine := lines[ctx.RowStart-1]
		errorLen := ctx.ColEnd - ctx.ColStart
		if ctx.RowStart != ctx.RowEnd {
			errorLen = 1
		}
		fmt.Fprintf(p.stdErr, "  %s\n", relevantLine)
		errorStartIndent := strings.Repeat(" ", ctx.ColStart-1)
		fmt.Fprintf(p.stdErr, "  %s%s\n", errorStartIndent, strings.Repeat("^", errorLen))

		if len(strings.Split(msg, "\n")) == 0 {
			fmt.Fprintf(p.stdErr, "  %s%s\n", errorStartIndent, color.RedString(msg))
		} else {
			fmt.Fprintf(p.stdErr, "\n%s\n", msg)
		}
	}
	p.printShellExitIfEnabled()
	p.errorExit(errorCode)
}

func (p *stdPrinter) RadErrorExit(msg string) {
	fmt.Fprint(p.stdErr, msg)
	p.printShellExitIfEnabled()
	p.errorExit(1)
}

func (p *stdPrinter) RadTokenErrorExit(token Token, msg string) {
	if token == nil {
		fmt.Fprint(p.stdErr, msg)
	} else {
		lexeme := token.GetLexeme()
		lexeme = strings.ReplaceAll(lexeme, "\n", "\\n")
		fmt.Fprintf(p.stdErr, "RadError at L%d/%d on '%s': %s",
			token.GetLine(), token.GetCharLineStart(), token.GetLexeme(), msg)
	}
}

func (p *stdPrinter) RadNodeErrorExit(node rts.Node, msg string) {
	fmt.Fprintf(p.stdErr, "RadError at L%d/%d in '%s': %s",
		node.StartPos().Row+1, node.StartPos().Col+1, node.Src(), msg)
}

func (p *stdPrinter) UsageErrorExit(msg string) {
	fmt.Fprint(p.stdErr, msg+"\n")
	p.runner.RunUsage()
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
		panic("Stacktrace because --RAD-DEBUG is enabled") // todo should do debug stack instead?
	} else {
		RExit(errorCode)
	}
}

type NullWriter struct{}

// Write implements the io.Writer interface for NullWriter.
// It accepts data but does nothing with it, effectively discarding it.
func (nw NullWriter) Write(p []byte) (int, error) {
	// Discard the data by doing nothing and returning the length of the data.
	return len(p), nil
}
