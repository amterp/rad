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
		emitPreflight(fmt.Sprintf("%s %v\n%s\n",
			color.RedString("error[RAD%s]:", rl.ErrPromptsNeedAnswers), err,
			com.CyanS(fmt.Sprintf("   = info: rad docs RAD%s", rl.ErrPromptsNeedAnswers))))
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
		emitPreflight(fmt.Sprintf(
			"%s --confirm-shell asks you to approve every shell command, but there's no terminal to ask at.\n\n"+
				"  Drop --confirm-shell, or run this where a terminal is available.\n",
			color.RedString("error:")))
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
		emitPreflight(fmt.Sprintf(
			"%s this script needs to ask you something, but there's no terminal\n\n"+
				"  %s prompts in %s, and it sets @%s = false.\n"+
				"  That removes the --reply flag which would otherwise answer them. Run\n"+
				"  this where a terminal is available, or re-enable global options in the\n"+
				"  script.\n",
			color.RedString("error[RAD%s]:", rl.ErrPromptsNeedAnswers),
			preflightScriptName(), com.Pluralize(len(sites), "place"),
			MACRO_ENABLE_GLOBAL_OPTIONS))
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

// preflightMessage explains the situation to a person first. The caller may
// well be an AI agent, but writing for a human keeps it readable for both, and
// the machine-facing part - the exact command to run - is unambiguous either way.
func preflightMessage(scriptName string, sites, unanswered []prompts.Site) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s this script needs to ask you something, but there's no terminal\n\n",
		color.RedString("error[RAD%s]:", rl.ErrPromptsNeedAnswers))

	fmt.Fprintf(&b, "  %s prompts in %s. rad found no terminal - stdin isn't one and\n",
		scriptName, com.Pluralize(len(sites), "place"))
	b.WriteString("  /dev/tty isn't available - so it can't ask. Supply the answers up front,\n")
	b.WriteString("  keyed by line number.\n\n")

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

	// Only worth saying when a placeholder actually made it into the command
	// above; copy-pasting it verbatim would otherwise answer with the literal
	// text "<value>" and run the script on it.
	if lo.SomeBy(unanswered, func(s prompts.Site) bool {
		return !s.Secret && !s.AsValue && placeholderFor(s) == ""
	}) {
		b.WriteString("  Replace each <...> with a real value first; rad takes them literally.\n")
	}
	b.WriteString("  Answers must match exactly; rad won't guess. For a prompt on a branch you\n")
	b.WriteString("  know won't run - or a filtered pick you expect to narrow to one option -\n")
	b.WriteString("  use --reply-na rather than inventing a value.\n")
	fmt.Fprintf(&b, "%s\n", com.CyanS(fmt.Sprintf("   = info: rad docs RAD%s", rl.ErrPromptsNeedAnswers)))

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
	if s.Filtered {
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

// suggestedCommand rebuilds the caller's own invocation with the missing
// answers appended, so the fix is a copy-paste rather than a reconstruction.
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
		// The two cases where no value can honestly go here: a secret must never
		// travel through argv, and a builtin used as a value has no call for an
		// answer to attach to. Everything else gets a --reply, including a
		// filtered pick - an answer the filter makes moot is dropped rather than
		// fought with, so offering one is safe either way.
		if s.Secret || s.AsValue {
			parts = append(parts, "--reply-na", s.Key)
			continue
		}
		value := placeholderFor(s)
		if value == "" {
			value = placeholderTextFor(s)
		}
		parts = append(parts, "--reply", s.Key+":"+value)
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

// placeholderFor is a value rad can honestly suggest, or "" when the caller has
// to supply one themselves.
func placeholderFor(s prompts.Site) string {
	switch s.Kind {
	case prompts.Confirm, prompts.ShellConfirm:
		return "yes"
	case prompts.Multipick:
		if len(s.Options) == 0 {
			return ""
		}
		// Suggesting every option overshoots a max the prompt sets, and an
		// option containing a comma has to survive the split it would land in.
		n := len(s.Options)
		if s.MaxSet && s.Max >= 0 && int64(n) > s.Max {
			n = int(s.Max)
		}
		escaped := lo.Map(s.Options[:n], func(o string, _ int) string { return escapeSelection(o) })
		return strings.Join(escaped, ",")
	default:
		if surviving := survivingOptions(s); len(surviving) > 0 {
			return surviving[0]
		}
		return ""
	}
}

// survivingOptions applies a pick's own filter to its own options, so the
// suggested answer is one the call will actually accept. Naming the first of
// the unfiltered list instead produces a command that fails on paste whenever
// the filter excludes it.
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

// placeholderTextFor is the stand-in shown when rad has no real value to offer.
// Deliberately angle-bracketed so it reads as something to replace.
func placeholderTextFor(s prompts.Site) string {
	if s.Kind == prompts.Multipick {
		return "<values>"
	}
	return "<value>"
}

// escapeSelection makes an option safe to put in a comma-separated multipick
// answer, mirroring what splitEscaped will undo.
func escapeSelection(opt string) string {
	return strings.NewReplacer(`\`, `\\`, `,`, `\,`).Replace(opt)
}

func commandPath(cmd *ScriptCommand) []string {
	var path []string
	for c := cmd; c != nil; c = c.Parent {
		path = append([]string{c.Name}, path...)
	}
	return path
}
