package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amterp/ra"
	com "github.com/amterp/rad/core/common"
	"github.com/amterp/radish"
	"github.com/samber/lo"
)

// errPromptCanceled marks a user abort (Esc/Ctrl-C) during the --interactive
// walk, distinguishing it from I/O failures so the runner can exit cleanly.
var errPromptCanceled = errors.New("interactive prompt canceled")

// runInteractivePrepass runs between the runner's two parses: by now the first
// parse has told us which args the CLI already supplied (and which command, if
// any, was invoked), and the second parse will validate whatever argv we return
// here - it stays the single source of truth, this pre-pass only fills gaps.
//
// Returns the equivalent non-interactive argv: the original tokens minus
// -i/--interactive, a selected command token at the front if one was prompted
// for, and the answered args appended as flags. The same argv is printed to
// stderr so the user can rerun the invocation directly.
func (r *RadRunner) runInteractivePrepass(argsToRead []string) []string {
	prompter := radishArgPrompter{}

	// The synthesized argv must not carry -i: its BypassValidation would neuter
	// the second parse's required/relational checks, and the printed invocation
	// must reproduce the run *without* prompting.
	stripped := stripInteractiveFlags(argsToRead, r.flagTokenInfo())

	cmdToken := ""
	walkArgs := r.scriptData.Args
	if len(r.cmdInvocations) > 0 {
		invoked := r.invokedCmd()
		if invoked == nil {
			names := make([]string, len(r.cmdInvocations))
			for i, inv := range r.cmdInvocations {
				names[i] = inv.cmd.ExternalName
			}
			choice, err := prompter.Select("Choose a command", names,
				func(choice string) string { return com.BoldS("Command:") + " " + com.GreenS(choice) })
			if err != nil {
				r.interactiveErrorExit(err)
			}
			cmdToken = choice
			// Re-parse with the command at the front: flags are registered on the
			// subcommand, so without a matched command anything the user already
			// typed sat in root's unknown args and Configured() would miss it. The
			// original (unstripped) tokens keep -i's validation bypass in effect.
			RRootCmd.ResetParseState()
			RRootCmd.ParseOrExit(
				append([]string{cmdToken}, argsToRead...),
				ra.WithIgnoreUnknown(true), ra.WithVariadicUnknownFlags(true),
			)
			invoked = r.invokedCmd()
		}
		if invoked != nil {
			// Command args first: they're the context the user just chose; shared
			// script args follow, matching the subcommand usage layout.
			walkArgs = append(append([]*ScriptArg{}, invoked.cmd.Args...), r.scriptData.Args...)
		}
	}

	notef := func(format string, a ...any) {
		fmt.Fprint(RIo.StdErr, com.YellowS(format, a...))
	}
	tokens, err := walkInteractiveArgs(walkArgs, RRootCmd.Configured, r.cliBoolLookup(), prompter, notef)
	if err != nil {
		r.interactiveErrorExit(err)
	}

	finalArgs := make([]string, 0, len(stripped)+len(tokens)+1)
	if cmdToken != "" {
		// Ra matches a subcommand on the first non-flag token, so it must lead.
		finalArgs = append(finalArgs, cmdToken)
	}
	finalArgs = append(finalArgs, stripped...)
	finalArgs = append(finalArgs, tokens...)

	printEquivalentInvocation(finalArgs)
	return finalArgs
}

func (r *RadRunner) invokedCmd() *cmdInvocation {
	for i := range r.cmdInvocations {
		if *r.cmdInvocations[i].usedPtr {
			return &r.cmdInvocations[i]
		}
	}
	return nil
}

// cliBoolLookup returns the parsed value of a bool flag the CLI supplied, so
// the walk knows whether --no-cache=false style input counts for exclusion.
// Covers shared script args and, when a command was invoked, its args too.
func (r *RadRunner) cliBoolLookup() func(externalName string) bool {
	radArgs := r.scriptArgs
	if invoked := r.invokedCmd(); invoked != nil {
		radArgs = append(append([]RadArg{}, invoked.args...), r.scriptArgs...)
	}
	return func(externalName string) bool {
		for _, a := range radArgs {
			if a.GetExternalName() == externalName {
				if b, ok := a.(*BoolRadArg); ok {
					return b.Value
				}
				return false
			}
		}
		return false
	}
}

func (r *RadRunner) interactiveErrorExit(err error) {
	if errors.Is(err, radish.ErrNotInteractive) {
		RP.ErrorExit("--interactive requires an interactive terminal\n")
	}
	if errors.Is(err, errPromptCanceled) {
		RP.ErrorExit("Interactive mode canceled\n")
	}
	RP.ErrorExit(fmt.Sprintf("Error during interactive prompting: %v\n", err))
}

// flagTokenInfo is what stripInteractiveFlags needs to walk an argv the way ra
// does: which flag tokens blindly consume the next token as their value (so a
// "-i" in that position is data, not the interactive flag), and which short
// runes are registered (so "-di" can be recognized as a flag cluster).
type flagTokenInfo struct {
	valueTaking map[string]bool
	shorts      map[rune]bool
}

// flagTokenInfo aggregates flag shapes across global flags, script args, and
// every command's args (at strip time we don't yet know which command, if any,
// applies). A flag consumes the next token unless it's a scalar bool (set by
// presence) or variadic (ra's greedy collection stops at registered flags
// rather than consuming one blindly).
func (r *RadRunner) flagTokenInfo() flagTokenInfo {
	info := flagTokenInfo{
		valueTaking: make(map[string]bool),
		shorts:      make(map[rune]bool),
	}
	add := func(externalName, short string, takesValue bool) {
		if short != "" {
			info.shorts[rune(short[0])] = true
		}
		if takesValue {
			info.valueTaking["--"+externalName] = true
			if short != "" {
				info.valueTaking["-"+short] = true
			}
		}
	}
	for _, g := range r.globalFlags {
		add(g.GetExternalName(), g.GetShort(), g.GetType() != ArgBoolT)
	}
	collect := func(args []*ScriptArg) {
		for _, a := range args {
			short := ""
			if a.Short != nil {
				short = *a.Short
			}
			takesValue := !a.IsVariadic && (a.Type != ArgBoolT || isListScriptArg(a))
			add(a.ExternalName, short, takesValue)
		}
	}
	if r.scriptData != nil {
		collect(r.scriptData.Args)
	}
	for _, inv := range r.cmdInvocations {
		collect(inv.cmd.Args)
	}
	return info
}

// stripInteractiveFlags drops the tokens that switched interactive mode on:
// -i, --interactive, and their =value and short-cluster (-di) forms. Tokens in
// a value position (right after a value-consuming flag) are data and survive,
// as does everything after a literal "--" separator.
//
// Known gap: a cluster whose trailing short consumes a value (-dc never)
// doesn't mark the next token as a value position; only exact -c/--color
// tokens do.
func stripInteractiveFlags(args []string, info flagTokenInfo) []string {
	out := make([]string, 0, len(args))
	inValue := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			out = append(out, args[i:]...)
			break
		}
		if inValue {
			out = append(out, a)
			inValue = false
			continue
		}
		if a == "-i" || strings.HasPrefix(a, "-i=") ||
			a == "--interactive" || strings.HasPrefix(a, "--interactive=") {
			continue
		}
		if rebuilt, ok := stripShortCluster(a, info.shorts); ok {
			if rebuilt != "" {
				out = append(out, rebuilt)
			}
			continue
		}
		out = append(out, a)
		inValue = info.valueTaking[a]
	}
	return out
}

// stripShortCluster removes the 'i' rune from a short-flag cluster like -di,
// returning the rebuilt token and whether the input was such a cluster. It
// only treats a token as a cluster when every rune is a registered short, so
// values that merely look dashed pass through untouched. An =value belongs to
// the cluster's last short; if that short is the 'i' being removed, the value
// goes with it.
func stripShortCluster(a string, shorts map[rune]bool) (string, bool) {
	if !strings.HasPrefix(a, "-") || strings.HasPrefix(a, "--") {
		return "", false
	}
	body := a[1:]
	eqVal := ""
	if idx := strings.Index(body, "="); idx != -1 {
		eqVal = body[idx:]
		body = body[:idx]
	}
	if len(body) < 2 || !strings.ContainsRune(body, 'i') {
		return "", false
	}
	for _, r := range body {
		if !shorts[r] {
			return "", false
		}
	}
	if eqVal != "" && strings.HasSuffix(body, "i") {
		eqVal = ""
	}
	rebuilt := strings.ReplaceAll(body, "i", "")
	if rebuilt == "" {
		return "", true
	}
	return "-" + rebuilt + eqVal, true
}

// printEquivalentInvocation tells the user how to rerun this exact invocation
// non-interactively. Stderr keeps it out of the script's own stdout.
func printEquivalentInvocation(args []string) {
	scriptRef := ScriptPath
	if scriptRef == "" {
		scriptRef = ScriptName
	}
	if scriptRef == "" {
		scriptRef = "-"
	}
	parts := append([]string{"rad", scriptRef}, args...)
	quoted := lo.Map(parts, func(s string, _ int) string { return shellQuoteIfNeeded(s) })
	fmt.Fprintf(RIo.StdErr, "%s %s\n", com.BoldS("Equivalent:"), com.GreenS(strings.Join(quoted, " ")))
}

// ArgPrompter abstracts the prompt shapes the --interactive walk needs, so the
// walk logic is unit-testable with a fake. The production implementation builds
// radish models and runs them through the RInteractive driver seam. Each prompt
// takes a summarize renderer producing the collapsed line left in the transcript
// after the prompt ends (compact flag-form, previewing the equivalent
// invocation); an empty result collapses to nothing.
type ArgPrompter interface {
	Select(title string, options []string, summarize func(choice string) string) (string, error)
	MultiSelect(title string, options, preselected []string, summarize func(chosen []string) string) ([]string, error)
	Input(title, placeholder string, validate func(string) error, summarize func(value string) string) (string, error)
}

type radishArgPrompter struct{}

func (radishArgPrompter) Select(title string, options []string, summarize func(string) string) (string, error) {
	model := radish.NewSelect().Title(title).Options(options...).Width(GetTermWidth())
	if summarize != nil {
		model.SummaryFunc(summarize)
	}
	res, _, err := RInteractive.Run(model)
	if err != nil {
		return "", err
	}
	if res.Canceled {
		return "", errPromptCanceled
	}
	selected, _ := model.Selected()
	return selected, nil
}

func (radishArgPrompter) MultiSelect(title string, options, preselected []string, summarize func([]string) string) ([]string, error) {
	model := radish.NewMultiSelect().Title(title).Options(options...).Preselect(preselected...).
		Hint("space to toggle, enter to confirm").Width(GetTermWidth())
	if summarize != nil {
		model.SummaryFunc(summarize)
	}
	res, _, err := RInteractive.Run(model)
	if err != nil {
		return nil, err
	}
	if res.Canceled {
		return nil, errPromptCanceled
	}
	return model.Selected(), nil
}

func (radishArgPrompter) Input(title, placeholder string, validate func(string) error, summarize func(string) string) (string, error) {
	model := radish.NewInput().Title(title).Prompt("> ").Width(GetTermWidth())
	if placeholder != "" {
		model.Placeholder(placeholder)
	}
	if validate != nil {
		model.Validate(validate)
	}
	if summarize != nil {
		model.SummaryFunc(summarize)
	}
	res, _, err := RInteractive.Run(model)
	if err != nil {
		return "", err
	}
	if res.Canceled {
		return "", errPromptCanceled
	}
	value, _ := model.Value()
	return value, nil
}

// walkState tracks, during the --interactive walk, what the final parse will
// see for each arg - mirroring ra's two distinct relational-constraint
// semantics. Excludes only count explicitly-set args (bools: set AND true);
// requires has has-value semantics, so defaults participate (bools: counts iff
// the value is true, supplied or not).
type walkState struct {
	// explicit marks args supplied on the CLI or answered with emitted tokens.
	explicit map[string]bool
	// boolVal is each scalar bool's best-known final value: the CLI-parsed
	// value if supplied, the walk answer once given, else the default.
	boolVal map[string]bool
}

func newWalkState(args []*ScriptArg, isConfigured func(string) bool, cliBoolVal func(string) bool) *walkState {
	st := &walkState{
		explicit: make(map[string]bool),
		boolVal:  make(map[string]bool),
	}
	for _, arg := range args {
		configured := isConfigured(arg.ExternalName)
		if configured {
			st.explicit[arg.ExternalName] = true
		}
		if isGroupableBool(arg) {
			if configured {
				st.boolVal[arg.ExternalName] = cliBoolVal(arg.ExternalName)
			} else {
				st.boolVal[arg.ExternalName] = arg.DefaultBool != nil && *arg.DefaultBool
			}
		}
	}
	return st
}

// noteAnswered records a walk answer that emitted tokens. Scalar bools emit
// exactly one token: the bare flag (true) or the =false form.
func (st *walkState) noteAnswered(arg *ScriptArg, tokens []string) {
	if len(tokens) == 0 {
		return
	}
	st.explicit[arg.ExternalName] = true
	if isGroupableBool(arg) {
		st.boolVal[arg.ExternalName] = tokens[0] == "--"+arg.ExternalName
	}
}

// countsForExcludes mirrors ra's flagExplicitlySetForExclusion: a default is
// the author's fallback, not user intent, so only explicitly-set args (bools:
// set AND true) can exclude others.
func (st *walkState) countsForExcludes(arg *ScriptArg) bool {
	if !st.explicit[arg.ExternalName] {
		return false
	}
	return !isGroupableBool(arg) || st.boolVal[arg.ExternalName]
}

// countsForRequires mirrors ra's flagConfiguredForRelationalConstraints:
// requires triggers off any value, including defaults (bools: iff true).
func (st *walkState) countsForRequires(arg *ScriptArg) bool {
	if isGroupableBool(arg) {
		return st.boolVal[arg.ExternalName]
	}
	return st.explicit[arg.ExternalName] || arg.HasDefaultValue
}

// walkInteractiveArgs prompts, in order, for each arg not already supplied on
// the CLI, and returns the argv tokens to append. Relational constraints are
// applied reactively as the walk progresses: an arg excluded by an explicitly
// set arg is skipped (with a note), and an arg required by a valued arg is
// forced - unless its own default already satisfies the requirement. The final
// parse remains the backstop for anything the walk cannot see, e.g. an answer
// that requires an arg the walk already passed.
//
// Bool args are special-cased: when two or more are unsupplied, they collapse
// into one MultiSelect at the position of the first one (toggled = true), so a
// flag-heavy script is one prompt instead of a y/n per flag.
func walkInteractiveArgs(
	args []*ScriptArg,
	isConfigured func(externalName string) bool,
	cliBoolVal func(externalName string) bool,
	prompter ArgPrompter,
	notef func(format string, a ...any),
) ([]string, error) {
	st := newWalkState(args, isConfigured, cliBoolVal)

	var groupBools []*ScriptArg
	for _, arg := range args {
		if !st.explicit[arg.ExternalName] && isGroupableBool(arg) {
			groupBools = append(groupBools, arg)
		}
	}
	if len(groupBools) < 2 {
		groupBools = nil // a lone bool keeps its y/n prompt
	}
	boolsHandled := false

	var tokens []string
	for _, arg := range args {
		if st.explicit[arg.ExternalName] {
			continue
		}
		if groupBools != nil && isGroupableBool(arg) {
			if boolsHandled {
				continue
			}
			boolsHandled = true
			groupTokens, err := promptBoolGroup(groupBools, args, st, prompter, notef)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, groupTokens...)
			continue
		}
		if excluder := excludedBy(arg, args, st); excluder != "" {
			notef("Skipping --%s (excluded by --%s)\n", arg.ExternalName, excluder)
			continue
		}
		required := !arg.IsNullable && !arg.HasDefaultValue
		requiredReason := ""
		// An arg with a default satisfies a requires constraint by itself, so
		// only force ones that would otherwise stay valueless.
		if by := requiredBy(arg, args, st); by != "" && !arg.HasDefaultValue {
			required = true
			requiredReason = fmt.Sprintf("required by --%s", by)
		}

		argTokens, err := promptForArg(arg, required, requiredReason, prompter)
		if err != nil {
			return nil, err
		}
		st.noteAnswered(arg, argTokens)
		tokens = append(tokens, argTokens...)
	}
	return tokens, nil
}

func isGroupableBool(arg *ScriptArg) bool {
	return arg.Type == ArgBoolT && !isListScriptArg(arg)
}

// promptBoolGroup runs one MultiSelect over the unsupplied bool args: each
// opens checked iff its default is true, and the toggled state is its final
// value. Tokens are emitted only where that differs from the default, same as
// the individual y/n path. Bools excluded by explicitly-set args are dropped
// (with a note); if only one bool survives, it falls back to a y/n prompt.
// Exclusions and requirements BETWEEN grouped bools can't react within a
// single prompt - the final parse backstops those.
func promptBoolGroup(
	bools []*ScriptArg,
	all []*ScriptArg,
	st *walkState,
	prompter ArgPrompter,
	notef func(format string, a ...any),
) ([]string, error) {
	var eligible []*ScriptArg
	for _, b := range bools {
		if excluder := excludedBy(b, all, st); excluder != "" {
			notef("Skipping --%s (excluded by --%s)\n", b.ExternalName, excluder)
			continue
		}
		eligible = append(eligible, b)
	}
	if len(eligible) == 0 {
		return nil, nil
	}
	if len(eligible) == 1 {
		tokens, err := promptForArg(eligible[0], false, "", prompter)
		if err == nil {
			st.noteAnswered(eligible[0], tokens)
		}
		return tokens, err
	}

	labels := make([]string, len(eligible))
	var preselected []string
	for i, b := range eligible {
		labels[i] = promptTitle(b)
		if b.DefaultBool != nil && *b.DefaultBool {
			preselected = append(preselected, labels[i])
		}
	}

	// The collapsed line previews exactly what lands in the equivalent
	// invocation: the flags whose final state differs from their default.
	// "(all defaults)" - not "(none)" - because unchecking a default-true flag
	// is a change, and leaving everything alone isn't "nothing picked".
	summarize := func(chosen []string) string {
		chosenSet := make(map[string]bool, len(chosen))
		for _, label := range chosen {
			chosenSet[label] = true
		}
		var diffs []string
		for i, b := range eligible {
			final := chosenSet[labels[i]]
			def := b.DefaultBool != nil && *b.DefaultBool
			if final == def {
				continue
			}
			if final {
				diffs = append(diffs, com.CyanS("--"+b.ExternalName))
			} else {
				diffs = append(diffs, com.CyanS("--"+b.ExternalName+"=false"))
			}
		}
		if len(diffs) == 0 {
			return com.BoldS("Flags:") + " " + com.FaintS("(all defaults)")
		}
		return com.BoldS("Flags:") + " " + strings.Join(diffs, ", ")
	}

	chosen, err := prompter.MultiSelect("Flags", labels, preselected, summarize)
	if err != nil {
		return nil, err
	}
	chosenSet := make(map[string]bool, len(chosen))
	for _, label := range chosen {
		chosenSet[label] = true
	}

	var tokens []string
	for i, b := range eligible {
		final := chosenSet[labels[i]]
		def := b.DefaultBool != nil && *b.DefaultBool
		// The user decided every member's final value, so record it for later
		// requires checks even where no token is emitted; explicit (and thus
		// excludes participation) still tracks emitted tokens only, since a
		// kept default is not explicitly set in the final parse's eyes.
		st.boolVal[b.ExternalName] = final
		if final == def {
			continue
		}
		if final {
			tokens = append(tokens, "--"+b.ExternalName)
		} else {
			tokens = append(tokens, "--"+b.ExternalName+"=false")
		}
		st.explicit[b.ExternalName] = true
	}
	return tokens, nil
}

// excludedBy returns the external name of an arg that excludes (in either
// direction) the given arg and counts for exclusion, or "" if none does.
func excludedBy(arg *ScriptArg, all []*ScriptArg, st *walkState) string {
	for _, other := range all {
		if other == arg || !st.countsForExcludes(other) {
			continue
		}
		if lo.Contains(other.ExcludesConstraint, arg.ExternalName) ||
			lo.Contains(arg.ExcludesConstraint, other.ExternalName) {
			return other.ExternalName
		}
	}
	return ""
}

// requiredBy returns the external name of an arg that requires the given arg
// and counts for requires (has a value, defaults included), or "" if none.
func requiredBy(arg *ScriptArg, all []*ScriptArg, st *walkState) string {
	for _, other := range all {
		if other == arg || !st.countsForRequires(other) {
			continue
		}
		if lo.Contains(other.RequiresConstraint, arg.ExternalName) {
			return other.ExternalName
		}
	}
	return ""
}

// promptForArg runs the prompt matching the arg's type and returns the argv
// tokens encoding the answer. An empty slice means the arg was skipped, so the
// final parse applies its default (or null). requiredReason, when non-empty,
// replaces the generic "value required" message - e.g. when another arg's
// requires constraint forced this prompt.
func promptForArg(arg *ScriptArg, required bool, requiredReason string, prompter ArgPrompter) ([]string, error) {
	flagToken := "--" + arg.ExternalName
	title := promptTitle(arg)

	if isListScriptArg(arg) {
		return promptList(arg, required, requiredReason, title, flagToken, prompter)
	}

	if arg.EnumConstraint != nil && arg.Type == ArgStringT {
		options := *arg.EnumConstraint
		skip := ""
		if !required {
			skip = enumSkipLabel(arg)
			options = append([]string{skip}, options...)
		}
		summarize := func(choice string) string {
			if skip != "" && choice == skip {
				return com.CyanS(flagToken) + " " + skipSummary(arg)
			}
			return com.CyanS(flagToken) + " " + com.GreenS(shellQuoteIfNeeded(choice))
		}
		choice, err := prompter.Select(title, options, summarize)
		if err != nil {
			return nil, err
		}
		if skip != "" && choice == skip {
			return nil, nil
		}
		return []string{flagToken, choice}, nil
	}

	if arg.Type == ArgBoolT {
		return promptBool(arg, title, flagToken, prompter)
	}

	summarize := func(value string) string {
		if value == "" {
			return com.CyanS(flagToken) + " " + skipSummary(arg)
		}
		return com.CyanS(flagToken) + " " + com.GreenS(shellQuoteIfNeeded(value))
	}
	answer, err := prompter.Input(title, defaultPlaceholder(arg), validatorFor(arg, required, requiredReason), summarize)
	if err != nil {
		return nil, err
	}
	if answer == "" {
		return nil, nil
	}
	return []string{flagToken, answer}, nil
}

// promptBool asks y/n with Enter meaning the default, and only emits tokens
// when the answer differs from the default, keeping the equivalent invocation
// minimal. Script bools always have a default (false unless specified).
func promptBool(arg *ScriptArg, title, flagToken string, prompter ArgPrompter) ([]string, error) {
	def := arg.DefaultBool != nil && *arg.DefaultBool
	hint := lo.Ternary(def, "[Y/n]", "[y/N]")
	// An answer matching the default emits no tokens, so the transcript must
	// not claim it was set - faint "(default: X)" keeps it honest with the
	// equivalent invocation.
	summarize := func(answer string) string {
		val := parseBoolAnswer(answer, def)
		if val == def {
			return com.CyanS(flagToken) + " " + com.FaintS("(default: %s)", lo.Ternary(def, "yes", "no"))
		}
		return com.CyanS(flagToken) + " " + com.GreenS(lo.Ternary(val, "yes", "no"))
	}
	answer, err := prompter.Input(title+" "+hint, "", boolValidator, summarize)
	if err != nil {
		return nil, err
	}
	val := parseBoolAnswer(answer, def)
	if val == def {
		return nil, nil
	}
	if val {
		return []string{flagToken}, nil
	}
	return []string{flagToken + "=false"}, nil
}

// parseBoolAnswer maps a y/n prompt answer to its final value; empty (just
// Enter) keeps the default. boolValidator has already constrained the answer.
func parseBoolAnswer(answer string, def bool) bool {
	switch strings.ToLower(answer) {
	case "y", "yes", "true":
		return true
	case "n", "no", "false":
		return false
	}
	return def
}

// boolValidator gates the y/n prompt; Enter (empty) means the default.
func boolValidator(s string) error {
	switch strings.ToLower(s) {
	case "", "y", "yes", "true", "n", "no", "false":
		return nil
	}
	return errors.New("answer y or n")
}

// promptList collects values one per line until an empty line. Variadic args
// are encoded greedily (--name a b c), the form Ra parses for them; plain list
// args take one value per flag occurrence (--name a --name b). Either way each
// value stays its own argv token, unambiguous regardless of commas or spaces.
// A value starting with "-" rides the =-form (--name=-5) instead: as a bare
// token it could end a variadic collection or read as a flag.
func promptList(arg *ScriptArg, required bool, requiredReason, title, flagToken string, prompter ArgPrompter) ([]string, error) {
	title += " (one per line, empty line to finish)"
	elemValidate := elementValidatorFor(arg)

	var values []string
	for {
		first := len(values) == 0
		validate := func(s string) error {
			if s == "" {
				if required && first {
					if requiredReason != "" {
						return errors.New(requiredReason)
					}
					return errors.New("at least one value required")
				}
				return nil
			}
			if elemValidate != nil {
				return elemValidate(s)
			}
			return nil
		}
		// One transcript line per value; the empty terminator collapses silently,
		// except when it's the first entry (the whole arg was skipped). Dash
		// values show the =-form they ride in the equivalent invocation.
		summarize := func(value string) string {
			if value != "" {
				if strings.HasPrefix(value, "-") {
					return com.CyanS(flagToken+"=") + com.GreenS(shellQuoteIfNeeded(value))
				}
				return com.CyanS(flagToken) + " " + com.GreenS(shellQuoteIfNeeded(value))
			}
			if first {
				return com.CyanS(flagToken) + " " + skipSummary(arg)
			}
			return ""
		}
		placeholder := ""
		if first {
			placeholder = defaultPlaceholder(arg)
		}
		answer, err := prompter.Input(title, placeholder, validate, summarize)
		if err != nil {
			return nil, err
		}
		if answer == "" {
			break
		}
		values = append(values, answer)
	}

	if len(values) == 0 {
		return nil, nil
	}
	if arg.IsVariadic {
		tokens := make([]string, 0, len(values)+1)
		inGreedyRun := false
		for _, v := range values {
			if strings.HasPrefix(v, "-") {
				tokens = append(tokens, flagToken+"="+v)
				inGreedyRun = false
				continue
			}
			if !inGreedyRun {
				tokens = append(tokens, flagToken)
				inGreedyRun = true
			}
			tokens = append(tokens, v)
		}
		return tokens, nil
	}
	tokens := make([]string, 0, len(values)*2)
	for _, v := range values {
		if strings.HasPrefix(v, "-") {
			tokens = append(tokens, flagToken+"="+v)
		} else {
			tokens = append(tokens, flagToken, v)
		}
	}
	return tokens, nil
}

// validatorFor wraps the arg's element validation with required/skip handling
// for scalar inputs: an empty answer means "skip" and is only rejected when
// the arg is required. requiredReason, when non-empty, names what forced the
// requirement (e.g. "required by --username").
func validatorFor(arg *ScriptArg, required bool, requiredReason string) func(string) error {
	elemValidate := elementValidatorFor(arg)
	return func(s string) error {
		if s == "" {
			if required {
				if requiredReason != "" {
					return errors.New(requiredReason)
				}
				return errors.New("value required")
			}
			return nil
		}
		if elemValidate != nil {
			return elemValidate(s)
		}
		return nil
	}
}

// elementValidatorFor returns the per-value validator for the arg's element
// type, or nil when anything goes. The synthesized argv carries the raw string,
// so validators must only accept what the final Ra parse will: e.g. bool list
// elements are "true"/"false", not "y"/"n".
func elementValidatorFor(arg *ScriptArg) func(string) error {
	switch elementType(arg.Type) {
	case ArgIntT:
		return func(s string) error {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return errors.New("must be an integer")
			}
			return checkRange(float64(n), arg.RangeConstraint)
		}
	case ArgFloatT:
		return func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return errors.New("must be a number")
			}
			return checkRange(f, arg.RangeConstraint)
		}
	case ArgBoolT:
		return func(s string) error {
			if s != "true" && s != "false" {
				return errors.New("must be 'true' or 'false'")
			}
			return nil
		}
	default:
		if arg.RegexConstraint != nil {
			re := arg.RegexConstraint
			return func(s string) error {
				if !re.MatchString(s) {
					return fmt.Errorf("must match regex: %s", re.String())
				}
				return nil
			}
		}
		if arg.EnumConstraint != nil {
			values := *arg.EnumConstraint
			return func(s string) error {
				if !lo.Contains(values, s) {
					return fmt.Errorf("must be one of: %s", strings.Join(values, ", "))
				}
				return nil
			}
		}
		return nil
	}
}

func isListScriptArg(arg *ScriptArg) bool {
	if arg.IsVariadic {
		return true
	}
	switch arg.Type {
	case ArgStrListT, ArgIntListT, ArgFloatListT, ArgBoolListT:
		return true
	}
	return false
}

func elementType(t RadArgTypeT) RadArgTypeT {
	switch t {
	case ArgStrListT:
		return ArgStringT
	case ArgIntListT:
		return ArgIntT
	case ArgFloatListT:
		return ArgFloatT
	case ArgBoolListT:
		return ArgBoolT
	}
	return t
}

func checkRange(v float64, rc *ArgRangeConstraint) error {
	if rc == nil {
		return nil
	}
	belowMin := rc.Min != nil && (v < *rc.Min || (!rc.MinInclusive && v == *rc.Min))
	aboveMax := rc.Max != nil && (v > *rc.Max || (!rc.MaxInclusive && v == *rc.Max))
	if belowMin || aboveMax {
		return fmt.Errorf("must be in range %s", rangeText(rc))
	}
	return nil
}

func rangeText(rc *ArgRangeConstraint) string {
	var sb strings.Builder
	sb.WriteString(lo.Ternary(rc.MinInclusive, "[", "("))
	if rc.Min != nil {
		sb.WriteString(fmt.Sprintf("%v", *rc.Min))
	}
	sb.WriteString(", ")
	if rc.Max != nil {
		sb.WriteString(fmt.Sprintf("%v", *rc.Max))
	}
	sb.WriteString(lo.Ternary(rc.MaxInclusive, "]", ")"))
	return sb.String()
}

func promptTitle(arg *ScriptArg) string {
	title := "--" + arg.ExternalName
	if arg.Description != nil && !com.IsBlank(*arg.Description) {
		title += ": " + *arg.Description
	}
	return title
}

func defaultPlaceholder(arg *ScriptArg) string {
	if d := argDefaultDisplay(arg); d != "" {
		return "Default: " + d
	}
	return ""
}

func enumSkipLabel(arg *ScriptArg) string {
	if d := argDefaultDisplay(arg); d != "" {
		return fmt.Sprintf("(skip - use default: %s)", d)
	}
	return "(skip)"
}

// skipSummary is the transcript annotation for a skipped arg.
func skipSummary(arg *ScriptArg) string {
	if d := argDefaultDisplay(arg); d != "" {
		return com.FaintS("(skip - default: %s)", d)
	}
	return com.FaintS("(skip)")
}

func argDefaultDisplay(arg *ScriptArg) string {
	switch {
	case arg.DefaultString != nil:
		return *arg.DefaultString
	case arg.DefaultInt != nil:
		return ToPrintable(*arg.DefaultInt)
	case arg.DefaultFloat != nil:
		return ToPrintable(*arg.DefaultFloat)
	case arg.DefaultBool != nil:
		return ToPrintable(*arg.DefaultBool)
	case arg.DefaultStringList != nil:
		return ToPrintable(*arg.DefaultStringList)
	case arg.DefaultIntList != nil:
		return ToPrintable(*arg.DefaultIntList)
	case arg.DefaultFloatList != nil:
		return ToPrintable(*arg.DefaultFloatList)
	case arg.DefaultBoolList != nil:
		return ToPrintable(*arg.DefaultBoolList)
	}
	return ""
}

// shellQuoteIfNeeded leaves shell-inert tokens bare and single-quotes the rest,
// so the printed equivalent invocation is paste-safe without being noisy.
func shellQuoteIfNeeded(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@%+=:,./_-", r) {
			return shellQuote(s)
		}
	}
	return s
}
