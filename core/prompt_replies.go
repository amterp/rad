package core

import (
	"fmt"
	"strings"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/prompts"
	"github.com/amterp/rad/rts/rl"
)

// takeReply looks up a command-line answer for the prompt at this call site.
//
// Sites are addressed by source position because that is the only identity that
// costs a script author nothing. The static walk and the interpreter convert the
// CST separately, so they hold different AST objects for the same code; both
// key off the call node so the positions line up. The awkward shapes - UFCS,
// interpolation, a call reached through a function value - are pinned
// end-to-end by "An answer reaches its prompt through every call shape" in
// core/testing/snapshots/functions/prompts_no_terminal.snap.
func takeReply(node rl.Node) (prompts.Answer, prompts.Outcome) {
	if node == nil || RReplies == nil {
		return prompts.Answer{}, prompts.NoReply
	}
	span := node.Span()
	answer, outcome := RReplies.Take(span.StartRow+1, span.StartCol+1)

	// Running out of answers only has to be fatal when there's nobody to ask.
	// Pre-answering the first few passes of a loop and typing the rest is a
	// reasonable thing to do, and stopping a user who is sitting right there
	// serves nothing. --reply-na stays fatal either way: that one is an
	// assertion about the script, which a terminal doesn't make true.
	if outcome == prompts.Exhausted && TerminalAvailable() {
		return prompts.Answer{}, prompts.NoReply
	}
	return answer, outcome
}

// replyPending reports whether an answer is still waiting for this call site,
// without consuming it. See Replies.Pending for why that matters.
func replyPending(node rl.Node) bool {
	if node == nil || RReplies == nil {
		return false
	}
	span := node.Span()
	return RReplies.Pending(span.StartRow+1, span.StartCol+1)
}

// unansweredPromptErr is what a prompt raises when it is reached and neither a
// terminal nor a usable answer is available. The remaining map goes with it, so
// one re-run can fix everything rather than uncovering the next problem.
func unansweredPromptErr(node rl.Node, outcome prompts.Outcome, what string) *RadError {
	var reason string
	switch outcome {
	case prompts.Unreachable:
		reason = fmt.Sprintf(
			"%s was reached, but --reply-na said it wouldn't be. "+
				"That's an assertion, not a fallback, so rad won't guess an answer", what)
	case prompts.Exhausted:
		// Two things land here and rad cannot tell them apart: the prompt runs
		// more times than the caller counted, or it rejected an answer and
		// re-asked. Asserting the first is what makes this unusable on the
		// second - an agent told to add a flag adds one, fails in exactly the
		// same place, and adds another. So name both and let the reader pick;
		// the pass count is what they check the script against.
		reason = fmt.Sprintf(
			"%s ran out of answers after %s. Either it runs more times than that, and needs "+
				"one --reply per pass, or it rejected an answer and re-asked, and needs a "+
				"value it accepts", what, com.Pluralize(passesAnswered(node), "pass"))
	default:
		reason = fmt.Sprintf(
			"%s needs an answer, and there's no terminal to ask at. "+
				"Supply one with --reply, or --reply-na if this run shouldn't reach it", what)
	}

	return NewErrorStrf("%s%s", reason, remainingPromptsHint()).SetCode(rl.ErrPromptUnanswerable)
}

// passesAnswered is how many answers the caller supplied for this site, which
// at exhaustion is how many passes got one.
func passesAnswered(node rl.Node) int {
	if node == nil || RReplies == nil {
		return 0
	}
	span := node.Span()
	return RReplies.Supplied(span.StartRow+1, span.StartCol+1)
}

// secretNeedsTerminalErr refuses a secret input that has no terminal to read
// from. A secret can never come off the command line - argv is readable by
// other processes on the machine - so this points at the only honest way out
// rather than at a --reply that would leak it.
func secretNeedsTerminalErr(prompt string) *RadError {
	return NewErrorStrf(
		"The secret input %q needs a terminal. It can't be answered with --reply, "+
			"because command-line arguments are visible to other processes. Run this "+
			"where a terminal is available, or use --reply-na if this run shouldn't reach it",
		prompt,
	).SetCode(rl.ErrPromptUnanswerable)
}

// remainingPromptsHint lists what else is still unanswered, so a caller fixing
// this error can fix the rest of the run in the same edit.
func remainingPromptsHint() string {
	remaining := RReplies.Unanswered()
	if len(remaining) == 0 {
		return ""
	}

	keys := make([]string, 0, len(remaining))
	for _, site := range remaining {
		keys = append(keys, fmt.Sprintf("%s (%s)", site.Key, site.Kind))
	}
	return ".\nStill unanswered on this path: " + strings.Join(keys, ", ")
}
