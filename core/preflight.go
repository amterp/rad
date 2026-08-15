package core

import (
	"fmt"
	"strings"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/prompts"
	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/color"
	"github.com/samber/lo"
)

const (
	// exitPromptsNeedAnswers is distinct from the generic 1 so a caller can tell
	// "this run needs input" apart from "this run failed" without parsing text.
	exitPromptsNeedAnswers = 7
)

// runPromptPreflight decides whether this run can proceed, given what the
// script is going to ask for and what the caller supplied.
//
// It runs before a single statement executes. That ordering is the whole point:
// discovering a prompt halfway through means the script has already fetched,
// written, or deleted things, and the caller has to reason about a partial run
// before they can retry. Here, nothing has happened yet.
//
// The replies it parses are stored for the runtime to consume, whether or not a
// terminal exists - a caller who supplies an answer gets it used either way.
func (r *RadRunner) runPromptPreflight(invoked *ScriptCommand) {
	sites := prompts.Find(r.scriptData.Ast, r.scriptData.Resolved, commandPath(invoked))

	replies, err := prompts.ParseReplies(FlagReply.Value, FlagReplyNa.Value, sites)
	if err != nil {
		var b strings.Builder
		writeDiagnosticHeader(&b, rl.ErrPromptsNeedAnswers, err.Error())
		b.WriteString(com.CyanS(fmt.Sprintf("  = info: rad docs RAD%s", rl.ErrPromptsNeedAnswers)) + "\n")
		emitPreflight(b.String())
		RExit.Exit(exitPromptsNeedAnswers)
		return
	}
	RReplies = replies

	if TerminalAvailable() {
		return // a human is here; anything unanswered just gets asked
	}

	// --confirm-shell gates every shell command, not only the ones the author
	// marked, so the static walk can't enumerate them. Rather than let the run
	// start and stop at the first one - after earlier commands already ran -
	// refuse the combination outright. A script with no shell commands at all
	// has nothing to gate, so it isn't caught by this.
	if FlagConfirmShellCommands.Value && hasShellCommand(r.scriptData.Ast) {
		var b strings.Builder
		writeDiagnosticHeader(&b, "",
			"--confirm-shell asks you to approve every shell command, but there's no terminal to ask at.")
		b.WriteString("\n")
		writePara(&b, "Drop --confirm-shell, or run this where a terminal is available.")
		emitPreflight(b.String())
		RExit.Exit(exitPromptsNeedAnswers)
		return
	}

	unanswered := replies.Unanswered()
	if len(unanswered) == 0 {
		return
	}

	// Without global flags there is no --reply to offer, so the usual message
	// would tell the caller to run a command the script itself rejects. Say what
	// is actually true instead. Stopping here still holds the line that nothing
	// runs before the caller knows what this needs.
	if r.scriptData.DisableGlobalOpts {
		var b strings.Builder
		writeDiagnosticHeader(&b, rl.ErrPromptsNeedAnswers,
			"this script needs to ask you something, but there's no terminal")
		b.WriteString("\n")
		writePara(&b, fmt.Sprintf(
			"%s prompts in %s, and it sets @%s = false. That removes the --reply flag "+
				"which would otherwise answer them. Run this where a terminal is available, "+
				"or re-enable global options in the script.",
			preflightScriptName(), com.Pluralize(len(sites), "place"),
			MACRO_ENABLE_GLOBAL_OPTIONS))
		emitPreflight(b.String())
		RExit.Exit(exitPromptsNeedAnswers)
		return
	}

	emitPreflight(preflightMessage(preflightScriptName(), sites, unanswered))
	RExit.Exit(exitPromptsNeedAnswers)
}

// preflightScriptName is what to call the script in the message. A script read
// from stdin has no name, and an empty one renders as a stray colon.
func preflightScriptName() string {
	if ScriptName == "" {
		return "<script>"
	}
	return ScriptName
}

func hasShellCommand(file *rl.SourceFile) bool {
	if file == nil {
		return true // can't tell; assume the worst rather than let the run start
	}
	found := false
	rl.Walk(file, func(n rl.Node) {
		if _, ok := n.(*rl.Shell); ok {
			found = true
		}
	})
	return found
}

// emitPreflight writes past --quiet on purpose. Quiet suppresses chatter; this
// is the only thing standing between the caller and a run that cannot work, and
// silently withholding it would leave them with a bare non-zero exit.
func emitPreflight(msg string) {
	fmt.Fprint(RIo.StdErr, msg)
	emitShellExit(exitPromptsNeedAnswers)
}

// writeDiagnosticHeader writes "error[RADxxxxx]: message", wrapping the message
// to the terminal with a 2-column hanging indent - the same shape
// DiagnosticRenderer produces, which these messages have to match by hand
// because they build their own bodies rather than going through the renderer.
// Pass an empty code for a bare "error:".
func writeDiagnosticHeader(b *strings.Builder, code rl.Error, message string) {
	prefix := "error:"
	if code != "" {
		// code.String() carries the "RAD" prefix already.
		prefix = fmt.Sprintf("error[%s]:", code.String())
	}

	// Wrap against the uncolored prefix, then swap the color in: measuring an
	// escape-laden string against a column budget would come out far too wide.
	for i, line := range com.WrapPrefixed(message, prefix+" ", "  ", DiagnosticProseWidth()) {
		if i == 0 {
			b.WriteString(color.RedString(prefix) + strings.TrimPrefix(line, prefix) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
}

// writePara wraps a paragraph to the terminal and indents it into the message
// body. The site table and the suggested command are written directly instead:
// one is a column layout and the other is meant to be copy-pasted, so neither
// survives being reflowed.
func writePara(b *strings.Builder, text string) {
	for _, line := range com.Wrap(text, DiagnosticProseWidth()-2) {
		b.WriteString("  " + line + "\n")
	}
}

// preflightMessage explains the situation to a person first. The caller may
// well be an AI agent, but writing for a human keeps it readable for both, and
// the machine-facing part - the exact command to run - is unambiguous either way.
func preflightMessage(scriptName string, sites, unanswered []prompts.Site) string {
	var b strings.Builder

	writeDiagnosticHeader(&b, rl.ErrPromptsNeedAnswers,
		"this script needs to ask you something, but there's no terminal")
	b.WriteString("\n")

	writePara(&b, fmt.Sprintf(
		"%s prompts in %s. rad found no terminal - stdin isn't one and /dev/tty isn't "+
			"available - so it can't ask. Supply the answers up front, keyed by line number.",
		scriptName, com.Pluralize(len(sites), "place")))
	b.WriteString("\n")

	missing := make(map[string]bool, len(unanswered))
	for _, u := range unanswered {
		missing[u.Key] = true
	}
	// Only worth a marker column once some sites are already covered; on the
	// first run everything is missing and the column would be dead space.
	showMarkers := len(unanswered) < len(sites)

	keyWidth, labelWidth := 0, 0
	for _, s := range sites {
		if n := len(s.Key); n > keyWidth {
			keyWidth = n
		}
		if n := len(s.Label()); n > labelWidth {
			labelWidth = n
		}
	}

	for _, s := range sites {
		var marker string
		if showMarkers {
			if missing[s.Key] {
				marker = "   "
			} else {
				marker = color.GreenString("ok") + " "
			}
		}
		line := fmt.Sprintf("  %s%s:%-*s  %-*s %s",
			marker, scriptName, keyWidth, s.Key, labelWidth, s.Label(), describeSite(s))
		b.WriteString(strings.TrimRight(line, " ") + "\n")
	}

	b.WriteString("\n")
	fmt.Fprintf(&b, "    %s\n\n", suggestedCommand(unanswered))

	// All of this is about filling in a blank, so it only belongs here when the
	// command above has one. A run whose every site takes --reply-na is already
	// complete, and telling that caller to fill something in sends them looking
	// for it.
	if lo.SomeBy(unanswered, func(s prompts.Site) bool { return !answeredByRad(s) }) {
		writePara(&b, "Every <...> is a blank - rad writes the shape of the command, never "+
			"the answers. Fill each one in; rad takes the text literally and matches it "+
			"exactly. For a prompt on a branch you know won't run, or a filtered pick you "+
			"expect to settle itself, use --reply-na instead of choosing for it.")
	}

	// Only where something above repeats. This run was harmless - nothing has
	// executed - so the advice is about the run the caller is about to start,
	// and on a script with no repeating prompt there is nothing to count.
	if lo.SomeBy(unanswered, func(s prompts.Site) bool { return s.Repeats }) {
		b.WriteString("\n")
		writePara(&b, "One --reply answers one pass, and rad can't count the passes - that "+
			"depends on the script's own data. Too few stops the run partway, after it has "+
			"already done whatever came before, so read the script and count them where "+
			"that would cost you.")
	}
	fmt.Fprintf(&b, "%s\n", com.CyanS(fmt.Sprintf("  = info: rad docs RAD%s", rl.ErrPromptsNeedAnswers)))

	return b.String()
}

func describeSite(s prompts.Site) string {
	var parts []string

	if s.Prompt != "" {
		parts = append(parts, fmt.Sprintf("%q", s.Prompt))
	}
	if s.AsValue {
		// Listed rather than silently blocking: a script that merely names one
		// on a path this run won't take has to stay runnable.
		parts = append(parts, "used as a value - rad can't see where it runs; assert with --reply-na")
		return strings.Join(parts, "  ")
	}
	if s.Secret {
		parts = append(parts, "secret - cannot be answered on the command line")
	}
	if len(s.Options) > 0 {
		parts = append(parts, "options: "+strings.Join(s.Options, ", "))
	} else if s.Kind == prompts.Pick || s.Kind == prompts.Multipick {
		parts = append(parts, "options computed at runtime")
	}
	if settlesItself(s) {
		// Naming the survivor is a fact about the code, not a suggested answer:
		// this call cannot land anywhere else, and saying so beats leaving the
		// caller to apply the filter by eye.
		parts = append(parts, fmt.Sprintf("settles on %q - won't ask", survivingOptions(s)[0]))
	} else if s.Filtered {
		// The second clause matters: an answer for a call the filter settles on
		// its own is consumed and dropped, which is surprising unless said.
		parts = append(parts, "filtered - may not prompt, and then ignores its answer")
	}
	if s.Repeats {
		// Without this the suggested command runs one pass, does whatever that
		// pass does, and then stops for want of a second answer.
		parts = append(parts, "may run more than once - repeat --reply per run")
	}
	if len(parts) == 0 {
		// Everything about this call is built at runtime; say so rather than
		// leaving a bare line the reader has to go look up in the source.
		parts = append(parts, "prompt built at runtime")
	}

	return strings.Join(parts, "  ")
}

// suggestedCommand rebuilds the caller's own invocation with a slot for every
// missing answer appended, so the fix is filling in blanks rather than a
// reconstruction.
//
// Quoting matters more than it looks: unquoted, a multipick's `\,` is eaten by
// the shell and rad silently sees two selections instead of one.
func suggestedCommand(unanswered []prompts.Site) string {
	scriptRef := ScriptPath
	if scriptRef == "" {
		scriptRef = ScriptName
	}
	if scriptRef == "" {
		scriptRef = "-"
	}

	parts := append([]string{"rad", scriptRef}, originalScriptArgs()...)

	for _, s := range unanswered {
		if answeredByRad(s) {
			parts = append(parts, "--reply-na", s.Key)
			continue
		}
		parts = append(parts, "--reply", s.Key+":"+blankFor(s))
	}

	quoted := lo.Map(parts, func(s string, _ int) string { return shellQuoteIfNeeded(s) })
	return com.GreenS(strings.Join(quoted, " "))
}

// originalScriptArgs is the caller's own arguments, minus the rad binary and the
// script path, so the suggestion reproduces their invocation rather than a
// generic one.
//
// It reads RawArgs rather than os.Args: a script piped in on stdin has its `-`
// stripped from os.Args before this runs, which would otherwise shift the slice
// and silently drop one of the caller's own flags.
func originalScriptArgs() []string {
	args := RawArgs
	if len(args) < 2 {
		return nil
	}
	return args[1:] // drop the script path / `-`
}

// answeredByRad reports whether rad can fill this site in itself rather than
// leave the caller a blank. Three cases, and all three are things rad knows
// rather than things it guesses: a secret must never travel through argv, a
// builtin used as a value has no call for an answer to attach to, and a pick
// that settles on one option never asks. --reply-na states each of them.
func answeredByRad(s prompts.Site) bool {
	return s.Secret || s.AsValue || settlesItself(s)
}

// settlesItself reports whether a pick can only land on one option, which it
// takes without asking. Options and any filter both have to be literal for
// that to be knowable here; a filter built at runtime leaves it open.
func settlesItself(s prompts.Site) bool {
	if s.Kind != prompts.Pick || len(s.Options) == 0 {
		return false
	}
	if s.Filtered && len(s.Filter) == 0 {
		return false
	}
	return len(survivingOptions(s)) == 1
}

// blankFor is the slot rad prints in place of an answer. It names the shape the
// answer takes and stops there: which option to take, or whether to say yes to
// "this is a one-way trip, proceed?", is the caller's to decide, and a
// plausible value filled in for them reads as advice rad has no standing to
// give - then gets pasted as one. The prompt's own options sit a few lines
// above, so the slot doesn't have to repeat them.
func blankFor(s prompts.Site) string {
	switch s.Kind {
	case prompts.Confirm, prompts.ShellConfirm:
		return "<yes|no>"
	case prompts.Pick:
		return "<option>"
	case prompts.Multipick:
		return "<option,...>"
	default:
		return "<value>"
	}
}

// survivingOptions applies a pick's own filter to its own options, so rad can
// tell whether the call has a choice left to make.
func survivingOptions(s prompts.Site) []string {
	if len(s.Filter) == 0 {
		return s.Options
	}
	var kept []string
	for _, opt := range s.Options {
		matchesAll := true
		for _, f := range s.Filter {
			if !FuzzyMatchFold(f, opt) {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			kept = append(kept, opt)
		}
	}
	return kept
}

func commandPath(cmd *ScriptCommand) []string {
	var path []string
	for c := cmd; c != nil; c = c.Parent {
		path = append([]string{c.Name}, path...)
	}
	return path
}
