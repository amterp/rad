package core

import (
	"testing"

	"github.com/amterp/radish"
	"github.com/stretchr/testify/assert"
)

// fakePrompter scripts answers for walkInteractiveArgs tests. Select answers
// and Input answers are consumed in prompt order; running out fails the test
// via the returned error. Validators are exercised against the answer so the
// tests prove the walk wires them up, not just that answers flow through.
type fakePrompter struct {
	t       *testing.T
	answers []string
	prompts []string // titles seen, in order
	err     error    // if set, returned by every prompt
}

func (f *fakePrompter) next(title string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	f.prompts = append(f.prompts, title)
	if len(f.answers) == 0 {
		f.t.Fatalf("prompt %q: no scripted answer left", title)
	}
	answer := f.answers[0]
	f.answers = f.answers[1:]
	return answer, nil
}

func (f *fakePrompter) Select(title string, options []string) (string, error) {
	answer, err := f.next(title)
	if err != nil {
		return "", err
	}
	assert.Contains(f.t, options, answer, "scripted Select answer must be an offered option")
	return answer, nil
}

func (f *fakePrompter) Input(title, placeholder string, validate func(string) error) (string, error) {
	answer, err := f.next(title)
	if err != nil {
		return "", err
	}
	if validate != nil {
		assert.NoError(f.t, validate(answer), "scripted Input answer must pass the arg's validator")
	}
	return answer, nil
}

func strArg(name string, opts ...func(*ScriptArg)) *ScriptArg {
	arg := &ScriptArg{Name: name, ExternalName: name, Type: ArgStringT}
	for _, opt := range opts {
		opt(arg)
	}
	return arg
}

func withDefaultStr(s string) func(*ScriptArg) {
	return func(a *ScriptArg) { a.DefaultString = &s; a.HasDefaultValue = true }
}

func notConfigured(string) bool { return false }

func noNotes(string, ...any) {}

func TestWalkPromptsOnlyUnconfiguredArgs(t *testing.T) {
	args := []*ScriptArg{strArg("alpha"), strArg("beta")}
	p := &fakePrompter{t: t, answers: []string{"b-val"}}

	tokens, err := walkInteractiveArgs(args,
		func(ext string) bool { return ext == "alpha" }, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--beta", "b-val"}, tokens)
	assert.Equal(t, []string{"--beta"}, p.prompts)
}

func TestWalkEnumUsesSelectAndOptionalGetsSkipRow(t *testing.T) {
	enum := []string{"dev", "prod"}
	required := strArg("env", func(a *ScriptArg) { a.EnumConstraint = &enum })
	optional := strArg("tier", withDefaultStr("small"), func(a *ScriptArg) { a.EnumConstraint = &enum })

	p := &fakePrompter{t: t, answers: []string{"prod", "(skip - use default: small)"}}
	tokens, err := walkInteractiveArgs([]*ScriptArg{required, optional}, notConfigured, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--env", "prod"}, tokens, "skipped optional enum emits no tokens")
}

func TestWalkBoolEmitsTokenOnlyWhenDifferingFromDefault(t *testing.T) {
	defTrue := true
	args := []*ScriptArg{
		{Name: "force", ExternalName: "force", Type: ArgBoolT, HasDefaultValue: true},
		{Name: "cache", ExternalName: "cache", Type: ArgBoolT, HasDefaultValue: true, DefaultBool: &defTrue},
		{Name: "loud", ExternalName: "loud", Type: ArgBoolT, HasDefaultValue: true},
	}
	// force: "y" flips default false -> --force; cache: "n" flips default true ->
	// --cache=false; loud: Enter keeps the default -> no token.
	p := &fakePrompter{t: t, answers: []string{"y", "n", ""}}
	tokens, err := walkInteractiveArgs(args, notConfigured, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--force", "--cache=false"}, tokens)
}

func TestWalkOptionalInputSkippedOnEmpty(t *testing.T) {
	args := []*ScriptArg{strArg("name", withDefaultStr("anon"))}
	p := &fakePrompter{t: t, answers: []string{""}}
	tokens, err := walkInteractiveArgs(args, notConfigured, p, noNotes)

	assert.NoError(t, err)
	assert.Empty(t, tokens, "empty answer on optional arg means use-default")
}

func TestWalkListCollectsRepeatedFlags(t *testing.T) {
	args := []*ScriptArg{{Name: "tags", ExternalName: "tags", Type: ArgStrListT}}
	p := &fakePrompter{t: t, answers: []string{"a", "b c", ""}}
	tokens, err := walkInteractiveArgs(args, notConfigured, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--tags", "a", "--tags", "b c"}, tokens)
}

func TestWalkVariadicUsesGreedyFlagForm(t *testing.T) {
	// Variadic flags take all their values after one flag token (--files a b);
	// the repeated-flag form is reserved for plain list args.
	args := []*ScriptArg{{Name: "files", ExternalName: "files", Type: ArgStringT, IsVariadic: true}}
	p := &fakePrompter{t: t, answers: []string{"x.txt", "y 2.txt", ""}}
	tokens, err := walkInteractiveArgs(args, notConfigured, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--files", "x.txt", "y 2.txt"}, tokens)
}

func TestWalkExcludesSkipsAndNotes(t *testing.T) {
	a := strArg("json")
	b := strArg("csv", func(arg *ScriptArg) { arg.ExcludesConstraint = []string{"json"} })

	var notes []string
	notef := func(format string, args ...any) { notes = append(notes, format) }

	// json supplied on CLI; csv excludes json (checked in both directions), so
	// csv's prompt is skipped entirely.
	p := &fakePrompter{t: t}
	tokens, err := walkInteractiveArgs([]*ScriptArg{a, b},
		func(ext string) bool { return ext == "json" }, p, notef)

	assert.NoError(t, err)
	assert.Empty(t, tokens)
	assert.Empty(t, p.prompts, "excluded arg must not be prompted")
	assert.Len(t, notes, 1)
}

func TestWalkAnsweredArgExcludesLaterArg(t *testing.T) {
	a := strArg("json", func(arg *ScriptArg) { arg.ExcludesConstraint = []string{"csv"} })
	b := strArg("csv", withDefaultStr("out.csv"))

	p := &fakePrompter{t: t, answers: []string{"out.json"}}
	tokens, err := walkInteractiveArgs([]*ScriptArg{a, b}, notConfigured, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--json", "out.json"}, tokens)
	assert.Equal(t, []string{"--json"}, p.prompts)
}

func TestWalkRequiresForcesOptionalArg(t *testing.T) {
	a := strArg("user", func(arg *ScriptArg) { arg.RequiresConstraint = []string{"token"} })
	b := strArg("token", withDefaultStr(""))

	p := &fakePrompter{t: t, answers: []string{"alice", "s3cret"}}
	tokens, err := walkInteractiveArgs([]*ScriptArg{a, b}, notConfigured, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--user", "alice", "--token", "s3cret"}, tokens)

	// The forced arg's validator must reject empty (no skipping).
	forced := validatorFor(b, true)
	assert.Error(t, forced(""))
}

func TestWalkPropagatesNotInteractive(t *testing.T) {
	args := []*ScriptArg{strArg("name")}
	p := &fakePrompter{t: t, err: radish.ErrNotInteractive}
	_, err := walkInteractiveArgs(args, notConfigured, p, noNotes)

	assert.ErrorIs(t, err, radish.ErrNotInteractive)
}

func TestValidators(t *testing.T) {
	min, max := 1.0, 10.0
	intArg := &ScriptArg{
		Name: "n", ExternalName: "n", Type: ArgIntT,
		RangeConstraint: &ArgRangeConstraint{Min: &min, MinInclusive: true, Max: &max, MaxInclusive: false},
	}
	v := validatorFor(intArg, true)
	assert.Error(t, v(""), "required rejects empty")
	assert.Error(t, v("abc"), "non-integer rejected")
	assert.Error(t, v("10"), "exclusive max rejected")
	assert.Error(t, v("0"), "below min rejected")
	assert.NoError(t, v("1"))
	assert.NoError(t, v("9"))

	floatArg := &ScriptArg{Name: "f", ExternalName: "f", Type: ArgFloatT}
	fv := validatorFor(floatArg, false)
	assert.NoError(t, fv(""), "optional accepts empty")
	assert.NoError(t, fv("1.5"))
	assert.Error(t, fv("x"))

	boolListArg := &ScriptArg{Name: "bs", ExternalName: "bs", Type: ArgBoolListT}
	bv := elementValidatorFor(boolListArg)
	assert.NoError(t, bv("true"))
	assert.Error(t, bv("y"), "bool list elements must be ra-parseable")
}

func TestShellQuoteIfNeeded(t *testing.T) {
	assert.Equal(t, "plain-token._/2", shellQuoteIfNeeded("plain-token._/2"))
	assert.Equal(t, "'has space'", shellQuoteIfNeeded("has space"))
	assert.Equal(t, `'it'\''s'`, shellQuoteIfNeeded("it's"))
	assert.Equal(t, "''", shellQuoteIfNeeded(""))
}

func TestStripInteractiveFlags(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, stripInteractiveFlags([]string{"-i", "a", "--interactive", "b"}))
	assert.Equal(t, []string{"a", "--", "-i"}, stripInteractiveFlags([]string{"a", "-i", "--", "-i"}),
		"tokens after -- are positional values, not flags")
}
