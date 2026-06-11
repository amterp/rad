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
	stripped := stripInteractiveFlags(argsToRead)

	cmdToken := ""
	walkArgs := r.scriptData.Args
	if len(r.cmdInvocations) > 0 {
		invoked := r.invokedCmd()
		if invoked == nil {
			names := make([]string, len(r.cmdInvocations))
			for i, inv := range r.cmdInvocations {
				names[i] = inv.cmd.ExternalName
			}
			choice, err := prompter.Select("Choose a command", names)
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
		fmt.Fprintf(RIo.StdErr, format, a...)
	}
	tokens, err := walkInteractiveArgs(walkArgs, RRootCmd.Configured, prompter, notef)
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

func (r *RadRunner) interactiveErrorExit(err error) {
	if errors.Is(err, radish.ErrNotInteractive) {
		RP.ErrorExit("--interactive requires an interactive terminal\n")
	}
	if errors.Is(err, errPromptCanceled) {
		RP.ErrorExit("Interactive mode canceled.\n")
	}
	RP.ErrorExit(fmt.Sprintf("Error during interactive prompting: %v\n", err))
}

// stripInteractiveFlags drops -i/--interactive tokens, leaving everything after
// a literal "--" separator untouched (those are positional values, not flags).
func stripInteractiveFlags(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			out = append(out, args[i:]...)
			break
		}
		if a == "-i" || a == "--interactive" || strings.HasPrefix(a, "--interactive=") {
			continue
		}
		out = append(out, a)
	}
	return out
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
	fmt.Fprintf(RIo.StdErr, "Equivalent: %s\n", strings.Join(quoted, " "))
}

// ArgPrompter abstracts the two prompt shapes the --interactive walk needs, so
// the walk logic is unit-testable with a fake. The production implementation
// builds radish models and runs them through the RInteractive driver seam.
type ArgPrompter interface {
	Select(title string, options []string) (string, error)
	Input(title, placeholder string, validate func(string) error) (string, error)
}

type radishArgPrompter struct{}

func (radishArgPrompter) Select(title string, options []string) (string, error) {
	model := radish.NewSelect().Title(title).Options(options...).Width(GetTermWidth())
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

func (radishArgPrompter) Input(title, placeholder string, validate func(string) error) (string, error) {
	model := radish.NewInput().Title(title).Prompt("> ").Width(GetTermWidth())
	if placeholder != "" {
		model.Placeholder(placeholder)
	}
	if validate != nil {
		model.Validate(validate)
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

// walkInteractiveArgs prompts, in order, for each arg not already supplied on
// the CLI, and returns the argv tokens to append. Relational constraints are
// applied reactively as the walk progresses: an arg excluded by a supplied or
// answered arg is skipped (with a note), and an arg required by one is forced
// (no skipping). The final parse remains the backstop for anything the walk
// cannot see, e.g. an answer that requires an arg the walk already passed.
func walkInteractiveArgs(
	args []*ScriptArg,
	isConfigured func(externalName string) bool,
	prompter ArgPrompter,
	notef func(format string, a ...any),
) ([]string, error) {
	// External names supplied on the CLI or answered during the walk.
	active := make(map[string]bool)
	for _, arg := range args {
		if isConfigured(arg.ExternalName) {
			active[arg.ExternalName] = true
		}
	}

	var tokens []string
	for _, arg := range args {
		if active[arg.ExternalName] {
			continue
		}
		if excluder := excludedBy(arg, args, active); excluder != "" {
			notef("Skipping --%s (excluded by --%s)\n", arg.ExternalName, excluder)
			continue
		}
		required := !arg.IsNullable && !arg.HasDefaultValue
		if requiredBy(arg, args, active) != "" {
			required = true
		}

		argTokens, err := promptForArg(arg, required, prompter)
		if err != nil {
			return nil, err
		}
		if len(argTokens) > 0 {
			active[arg.ExternalName] = true
			tokens = append(tokens, argTokens...)
		}
	}
	return tokens, nil
}

// excludedBy returns the external name of an active arg that excludes (in
// either direction) the given arg, or "" if none does.
func excludedBy(arg *ScriptArg, all []*ScriptArg, active map[string]bool) string {
	for _, other := range all {
		if other == arg || !active[other.ExternalName] {
			continue
		}
		if lo.Contains(other.ExcludesConstraint, arg.ExternalName) ||
			lo.Contains(arg.ExcludesConstraint, other.ExternalName) {
			return other.ExternalName
		}
	}
	return ""
}

// requiredBy returns the external name of an active arg that requires the
// given arg, or "" if none does.
func requiredBy(arg *ScriptArg, all []*ScriptArg, active map[string]bool) string {
	for _, other := range all {
		if other == arg || !active[other.ExternalName] {
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
// final parse applies its default (or null).
func promptForArg(arg *ScriptArg, required bool, prompter ArgPrompter) ([]string, error) {
	flagToken := "--" + arg.ExternalName
	title := promptTitle(arg)

	if isListScriptArg(arg) {
		return promptList(arg, required, title, flagToken, prompter)
	}

	if arg.EnumConstraint != nil && arg.Type == ArgStringT {
		options := *arg.EnumConstraint
		skip := ""
		if !required {
			skip = enumSkipLabel(arg)
			options = append([]string{skip}, options...)
		}
		choice, err := prompter.Select(title, options)
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

	answer, err := prompter.Input(title, defaultPlaceholder(arg), validatorFor(arg, required))
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
	answer, err := prompter.Input(title+" "+hint, "", boolValidator)
	if err != nil {
		return nil, err
	}
	val := def
	switch strings.ToLower(answer) {
	case "":
		// keep default
	case "y", "yes", "true":
		val = true
	case "n", "no", "false":
		val = false
	}
	if val == def {
		return nil, nil
	}
	if val {
		return []string{flagToken}, nil
	}
	return []string{flagToken + "=false"}, nil
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
func promptList(arg *ScriptArg, required bool, title, flagToken string, prompter ArgPrompter) ([]string, error) {
	title += " (one per line, empty line to finish)"
	elemValidate := elementValidatorFor(arg)

	var values []string
	for {
		first := len(values) == 0
		validate := func(s string) error {
			if s == "" {
				if required && first {
					return errors.New("at least one value required")
				}
				return nil
			}
			if elemValidate != nil {
				return elemValidate(s)
			}
			return nil
		}
		placeholder := ""
		if first {
			placeholder = defaultPlaceholder(arg)
		}
		answer, err := prompter.Input(title, placeholder, validate)
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
		return append([]string{flagToken}, values...), nil
	}
	tokens := make([]string, 0, len(values)*2)
	for _, v := range values {
		tokens = append(tokens, flagToken, v)
	}
	return tokens, nil
}

// validatorFor wraps the arg's element validation with required/skip handling
// for scalar inputs: an empty answer means "skip" and is only rejected when
// the arg is required.
func validatorFor(arg *ScriptArg, required bool) func(string) error {
	elemValidate := elementValidatorFor(arg)
	return func(s string) error {
		if s == "" {
			if required {
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
