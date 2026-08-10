package check

import (
	"github.com/amterp/rad/rts/rl"
)

// Check 14: positional shell capture targets whose names promise a different
// stream than their slot delivers.
//
// Positional captures fill (stdout, stderr, code) in order. Named assignment -
// every target spelled exactly `code`, `stdout`, or `stderr` - binds by name
// instead and is never flagged here. The gap between those two rules is where
// this check lives: `code, output = $cmd` reads like it captures the exit code
// and stdout, but `output` isn't a reserved name, so the whole statement falls
// back to positional and `code` receives *stdout*.
//
// Two tiers share one mechanism, asking the same question of every target:
// does this name promise a different stream than its slot assigns?
//
//   - Reserved names are exact. `code` in slot 0 is confusing on its own terms,
//     independent of any release, so that half is permanent.
//   - Synonyms (`exit_code`, `out`, `err`, ...) and the leading-`_` idiom are
//     heuristics that exist because v0.12 reordered positional captures from
//     (code, stdout, stderr). A script written against the old order still
//     parses and still runs - it just binds different variables. Delete the
//     migrationHint half, and the synonym tables it reads, once v0.12 is a
//     couple of releases behind us.
//
// Deliberately quiet on positional captures whose names give no signal
// (`a, b = $cmd`). After the flip `out, err = $cmd` is the *recommended* form,
// so warning on positional captures as a class would fire on correct new code
// and train people to skim past the tier. See check_regex_migration.go for the
// same argument at more length.

// shellNameSignals maps a target name to the stream that name promises.
// Reserved names are exact; the rest are the migration heuristic.
var shellNameSignals = map[string]struct {
	stream    rl.ShellStream
	isMigrant bool
}{
	// Reserved - permanent half.
	rl.ShellCaptureStdout: {rl.ShellStdout, false},
	rl.ShellCaptureStderr: {rl.ShellStderr, false},
	rl.ShellCaptureCode:   {rl.ShellCode, false},

	// Synonyms - migration half. Every entry has to be a name that only makes
	// sense for one stream. `status` is the instructive exclusion: it reads as
	// "exit status" but `status = $`git status --porcelain`` is a perfectly
	// good post-flip capture of stdout, so flagging it would cry wolf on
	// correct new code. When in doubt, leave the name out - a missed migration
	// hit costs one script; a false one costs the tier's credibility.
	"out":         {rl.ShellStdout, true},
	"output":      {rl.ShellStdout, true},
	"err":         {rl.ShellStderr, true},
	"errout":      {rl.ShellStderr, true},
	"exit_code":   {rl.ShellCode, true},
	"exitcode":    {rl.ShellCode, true},
	"exit_status": {rl.ShellCode, true},
	"retcode":     {rl.ShellCode, true},
	"ret_code":    {rl.ShellCode, true},
	"rc":          {rl.ShellCode, true},
}

func (c *RadCheckerImpl) addMisleadingShellCaptureNameWarnings(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	walkAST(c.ast, func(node rl.Node) {
		shell, ok := node.(*rl.Shell)
		if !ok || len(shell.Targets) == 0 {
			return
		}
		// Named assignment binds by name, so a name can't be in the wrong slot.
		if rl.IsNamedShellAssignment(shell.Targets) {
			return
		}
		if diag, found := c.misleadingCaptureDiagnostic(shell); found {
			*d = append(*d, diag)
		}
	})
}

// misleadingCaptureDiagnostic returns at most one diagnostic per statement.
// Flagging every mismatched target would double up on the common
// `code, output = $cmd` shape, where both names are wrong for the same single
// reason - the reader only needs telling once.
func (c *RadCheckerImpl) misleadingCaptureDiagnostic(shell *rl.Shell) (Diagnostic, bool) {
	type mismatch struct {
		target rl.Node
		name   string
		slot   rl.ShellStream
	}
	var (
		reserved        *mismatch // exact signal, preferred when both tiers hit
		migrant         *mismatch
		anyNameConfirms bool
	)

	for i, target := range shell.Targets {
		name := rl.ShellTargetRootName(target)
		signal, known := shellNameSignals[name]
		if !known {
			continue
		}
		slot := rl.ShellStreamForTarget(shell.Targets, i, false)
		if signal.stream == slot {
			anyNameConfirms = true
			continue
		}
		hit := &mismatch{target: target, name: name, slot: slot}
		if signal.isMigrant {
			if migrant == nil {
				migrant = hit
			}
		} else if reserved == nil {
			reserved = hit
		}
	}

	if reserved != nil {
		return c.shellCaptureDiagnostic(reserved.target, reserved.name, reserved.slot, false), true
	}
	if migrant != nil {
		return c.shellCaptureDiagnostic(migrant.target, migrant.name, migrant.slot, true), true
	}

	// The `_, x = $cmd` shape, which no name signal can catch because `x` is
	// an ordinary name like `cpu` or `version`. A leading `_` was the old
	// order's idiom for "discard the exit code, take stdout"; under the new
	// one it discards stdout and shifts everything after it by one.
	//
	// Two guards keep this from firing on correct code. anyNameConfirms
	// exempts `_, err = $cmd` - swallow stdout, capture stderr - which is a
	// legitimate post-flip capture. And we flag the first target that actually
	// binds something, so the pure-silencing `_, _ = $cmd` and `_, _, _ = $cmd`
	// stay quiet: nothing receives a stream there, so nothing can receive the
	// wrong one.
	if anyNameConfirms || len(shell.Targets) < 2 {
		return Diagnostic{}, false
	}
	if rl.ShellTargetRootName(shell.Targets[0]) != "_" {
		return Diagnostic{}, false
	}
	for i := 1; i < len(shell.Targets); i++ {
		target := shell.Targets[i]
		name := rl.ShellTargetRootName(target)
		if name == "_" {
			continue
		}
		slot := rl.ShellStreamForTarget(shell.Targets, i, false)
		return c.shellCaptureDiagnostic(target, name, slot, true), true
	}
	return Diagnostic{}, false
}

func (c *RadCheckerImpl) shellCaptureDiagnostic(
	target rl.Node,
	name string,
	got rl.ShellStream,
	isMigrant bool,
) Diagnostic {
	msg := "'" + name + "' is bound to " + shellStreamName(got) +
		" here. Positional shell captures fill (stdout, stderr, code) in that order; " +
		"assignment goes by name only when every target is spelled exactly " +
		"'code', 'stdout' or 'stderr'."

	suggestion := "Reorder the targets to match, or name them all so assignment goes by " +
		"name - e.g. 'stdout, code = $`cmd`'."
	if isMigrant {
		msg += " Before v0.12 the order was (code, stdout, stderr), so this used to bind " +
			shellStreamName(oldOrderStream(got)) + " instead - same script, different variable, no error."
		suggestion += " See https://amterp.dev/rad/migrations/v0.12/"
	}

	diag := NewDiagnosticWarnFromSpan(target.Span(), c.src, msg, rl.ErrMisleadingShellCaptureName)
	diag.Suggestion = &suggestion
	return diag
}

// oldOrderStream reports what the v0.11 order (code, stdout, stderr) put in the
// slot that now holds `got`. Used only to phrase the migration half of the
// message.
func oldOrderStream(got rl.ShellStream) rl.ShellStream {
	switch got {
	case rl.ShellStdout:
		return rl.ShellCode
	case rl.ShellStderr:
		return rl.ShellStdout
	case rl.ShellCode:
		return rl.ShellStderr
	}
	return got
}

func shellStreamName(s rl.ShellStream) string {
	switch s {
	case rl.ShellStdout:
		return "stdout"
	case rl.ShellStderr:
		return "stderr"
	case rl.ShellCode:
		return "the exit code"
	}
	return "nothing"
}
