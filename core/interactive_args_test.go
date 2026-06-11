package core

import (
	"testing"

	"github.com/amterp/radish"
	"github.com/stretchr/testify/assert"
)

// fakePrompter scripts answers for walkInteractiveArgs tests. Select and Input
// answers are consumed in prompt order from answers; MultiSelect answers from
// multiAnswers. Running out fails the test. Validators are exercised against
// the answer, and each prompt's summarize renderer is applied and recorded, so
// the tests prove the walk wires both up - not just that answers flow through.
type fakePrompter struct {
	t            *testing.T
	answers      []string
	multiAnswers [][]string
	prompts      []string   // titles seen, in order
	summaries    []string   // rendered collapse lines, in order
	preselects   [][]string // MultiSelect preselected labels, per prompt
	err          error      // if set, returned by every prompt
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

func (f *fakePrompter) recordSummary(line string) {
	if line != "" {
		f.summaries = append(f.summaries, line)
	}
}

func (f *fakePrompter) Select(title string, options []string, summarize func(string) string) (string, error) {
	answer, err := f.next(title)
	if err != nil {
		return "", err
	}
	assert.Contains(f.t, options, answer, "scripted Select answer must be an offered option")
	if summarize != nil {
		f.recordSummary(summarize(answer))
	}
	return answer, nil
}

func (f *fakePrompter) MultiSelect(title string, options, preselected []string, summarize func([]string) string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.prompts = append(f.prompts, title)
	f.preselects = append(f.preselects, preselected)
	if len(f.multiAnswers) == 0 {
		f.t.Fatalf("multiselect %q: no scripted answer left", title)
	}
	chosen := f.multiAnswers[0]
	f.multiAnswers = f.multiAnswers[1:]
	for _, c := range chosen {
		assert.Contains(f.t, options, c, "scripted MultiSelect answer must be an offered option")
	}
	if summarize != nil {
		f.recordSummary(summarize(chosen))
	}
	return chosen, nil
}

func (f *fakePrompter) Input(title, placeholder string, validate func(string) error, summarize func(string) string) (string, error) {
	answer, err := f.next(title)
	if err != nil {
		return "", err
	}
	if validate != nil {
		assert.NoError(f.t, validate(answer), "scripted Input answer must pass the arg's validator")
	}
	if summarize != nil {
		f.recordSummary(summarize(answer))
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

// noBoolVals stands in for the CLI bool-value lookup when no bools were
// supplied on the CLI (the walk then falls back to declared defaults).
func noBoolVals(string) bool { return false }

func noNotes(string, ...any) {}

func TestWalkPromptsOnlyUnconfiguredArgs(t *testing.T) {
	args := []*ScriptArg{strArg("alpha"), strArg("beta")}
	p := &fakePrompter{t: t, answers: []string{"b-val"}}

	tokens, err := walkInteractiveArgs(args,
		func(ext string) bool { return ext == "alpha" }, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--beta", "b-val"}, tokens)
	assert.Equal(t, []string{"--beta"}, p.prompts)
}

func TestWalkEnumUsesSelectAndOptionalGetsSkipRow(t *testing.T) {
	enum := []string{"dev", "prod"}
	required := strArg("env", func(a *ScriptArg) { a.EnumConstraint = &enum })
	optional := strArg("tier", withDefaultStr("small"), func(a *ScriptArg) { a.EnumConstraint = &enum })

	p := &fakePrompter{t: t, answers: []string{"prod", "(skip - use default: small)"}}
	tokens, err := walkInteractiveArgs([]*ScriptArg{required, optional}, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--env", "prod"}, tokens, "skipped optional enum emits no tokens")
}

func TestWalkSingleBoolUsesYesNoInput(t *testing.T) {
	args := []*ScriptArg{
		{Name: "force", ExternalName: "force", Type: ArgBoolT, HasDefaultValue: true},
	}
	p := &fakePrompter{t: t, answers: []string{"y"}}
	tokens, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--force"}, tokens)
	assert.Equal(t, []string{"--force [y/N]"}, p.prompts)
	assert.Equal(t, []string{"--force yes"}, p.summaries)
}

func TestWalkSingleBoolEnterKeepsDefault(t *testing.T) {
	defTrue := true
	args := []*ScriptArg{
		{Name: "cache", ExternalName: "cache", Type: ArgBoolT, HasDefaultValue: true, DefaultBool: &defTrue},
	}
	p := &fakePrompter{t: t, answers: []string{""}}
	tokens, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Empty(t, tokens)
	assert.Equal(t, []string{"--cache (default: yes)"}, p.summaries,
		"accepting the default must not read as an explicit answer - no token is emitted")
}

func TestWalkGroupsMultipleBoolsIntoMultiSelect(t *testing.T) {
	defTrue := true
	desc := "Build the project."
	args := []*ScriptArg{
		strArg("name"), // non-bool: stays an individual prompt
		{Name: "build", ExternalName: "build", Type: ArgBoolT, HasDefaultValue: true, Description: &desc},
		{Name: "cache", ExternalName: "cache", Type: ArgBoolT, HasDefaultValue: true, DefaultBool: &defTrue},
		{Name: "loud", ExternalName: "loud", Type: ArgBoolT, HasDefaultValue: true},
	}
	// MultiSelect result: build toggled on (default false -> --build), cache
	// toggled off (default true -> --cache=false), loud untouched (no token).
	p := &fakePrompter{
		t:            t,
		answers:      []string{"alice"},
		multiAnswers: [][]string{{"--build: Build the project."}},
	}
	tokens, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--name", "alice", "--build", "--cache=false"}, tokens)
	assert.Equal(t, []string{"--name", "Flags"}, p.prompts, "one multiselect, at the first bool's position")
	assert.Equal(t, [][]string{{"--cache"}}, p.preselects, "defaults-true bools open checked")
	assert.Equal(t, []string{"--name alice", "Flags: --build, --cache=false"}, p.summaries,
		"the group summary previews the equivalent invocation: every diff from defaults, including untoggles")
}

func TestWalkGroupAllDefaultsSummary(t *testing.T) {
	defTrue := true
	args := []*ScriptArg{
		{Name: "build", ExternalName: "build", Type: ArgBoolT, HasDefaultValue: true},
		{Name: "cache", ExternalName: "cache", Type: ArgBoolT, HasDefaultValue: true, DefaultBool: &defTrue},
	}
	// cache stays checked (its default) and build stays unchecked: no changes.
	p := &fakePrompter{t: t, multiAnswers: [][]string{{"--cache"}}}
	tokens, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Empty(t, tokens)
	assert.Equal(t, []string{"Flags: (all defaults)"}, p.summaries,
		"keeping every default is not '(none)' - nothing changed, nothing is emitted")
}

func TestWalkGroupExcludesConfiguredAndExcludedBools(t *testing.T) {
	args := []*ScriptArg{
		strArg("json", func(a *ScriptArg) { a.ExcludesConstraint = []string{"pretty"} }),
		{Name: "pretty", ExternalName: "pretty", Type: ArgBoolT, HasDefaultValue: true},
		{Name: "quiet2", ExternalName: "quiet2", Type: ArgBoolT, HasDefaultValue: true},
		{Name: "loud", ExternalName: "loud", Type: ArgBoolT, HasDefaultValue: true},
	}
	var notes []string
	notef := func(format string, a ...any) { notes = append(notes, format) }

	// json supplied on CLI excludes pretty -> pretty is dropped from the group
	// with a note; quiet2 and loud remain and still group (2 left).
	p := &fakePrompter{t: t, multiAnswers: [][]string{{"--loud"}}}
	tokens, err := walkInteractiveArgs(args,
		func(ext string) bool { return ext == "json" }, noBoolVals, p, notef)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--loud"}, tokens)
	assert.Len(t, notes, 1)
	assert.Equal(t, []string{"Flags"}, p.prompts)
}

func TestWalkOptionalInputSkippedOnEmpty(t *testing.T) {
	args := []*ScriptArg{strArg("name", withDefaultStr("anon"))}
	p := &fakePrompter{t: t, answers: []string{""}}
	tokens, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Empty(t, tokens, "empty answer on optional arg means use-default")
}

func TestWalkListCollectsRepeatedFlags(t *testing.T) {
	args := []*ScriptArg{{Name: "tags", ExternalName: "tags", Type: ArgStrListT}}
	p := &fakePrompter{t: t, answers: []string{"a", "b c", ""}}
	tokens, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--tags", "a", "--tags", "b c"}, tokens)
}

func TestWalkVariadicUsesGreedyFlagForm(t *testing.T) {
	// Variadic flags take all their values after one flag token (--files a b);
	// the repeated-flag form is reserved for plain list args.
	args := []*ScriptArg{{Name: "files", ExternalName: "files", Type: ArgStringT, IsVariadic: true}}
	p := &fakePrompter{t: t, answers: []string{"x.txt", "y 2.txt", ""}}
	tokens, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--files", "x.txt", "y 2.txt"}, tokens)
}

func TestWalkListDashValuesUseEqualsForm(t *testing.T) {
	// A bare "-5" or "--output" token in the synthesized argv could end a
	// variadic collection or read as a flag; the =-form is unambiguous.
	intList := []*ScriptArg{{Name: "nums", ExternalName: "nums", Type: ArgIntListT}}
	p := &fakePrompter{t: t, answers: []string{"-5", "3", ""}}
	tokens, err := walkInteractiveArgs(intList, notConfigured, noBoolVals, p, noNotes)
	assert.NoError(t, err)
	assert.Equal(t, []string{"--nums=-5", "--nums", "3"}, tokens)

	variadic := []*ScriptArg{{Name: "files", ExternalName: "files", Type: ArgStringT, IsVariadic: true}}
	p = &fakePrompter{t: t, answers: []string{"a.txt", "--output", "b.txt", ""}}
	tokens, err = walkInteractiveArgs(variadic, notConfigured, noBoolVals, p, noNotes)
	assert.NoError(t, err)
	assert.Equal(t, []string{"--files", "a.txt", "--files=--output", "--files", "b.txt"}, tokens,
		"greedy runs resume after a dash value rides the =-form")
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
		func(ext string) bool { return ext == "json" }, noBoolVals, p, notef)

	assert.NoError(t, err)
	assert.Empty(t, tokens)
	assert.Empty(t, p.prompts, "excluded arg must not be prompted")
	assert.Len(t, notes, 1)
}

func TestWalkAnsweredArgExcludesLaterArg(t *testing.T) {
	a := strArg("json", func(arg *ScriptArg) { arg.ExcludesConstraint = []string{"csv"} })
	b := strArg("csv", withDefaultStr("out.csv"))

	p := &fakePrompter{t: t, answers: []string{"out.json"}}
	tokens, err := walkInteractiveArgs([]*ScriptArg{a, b}, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--json", "out.json"}, tokens)
	assert.Equal(t, []string{"--json"}, p.prompts)
}

func TestWalkRequiresForcesOptionalArg(t *testing.T) {
	a := strArg("user", func(arg *ScriptArg) { arg.RequiresConstraint = []string{"token"} })
	b := strArg("token", func(arg *ScriptArg) { arg.IsNullable = true })

	p := &fakePrompter{t: t, answers: []string{"alice", "s3cret"}}
	tokens, err := walkInteractiveArgs([]*ScriptArg{a, b}, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--user", "alice", "--token", "s3cret"}, tokens)

	// The forced arg's validator must reject empty (no skipping), and say WHY
	// the normally-optional arg suddenly demands a value.
	forced := validatorFor(b, true, "required by --user")
	err = forced("")
	assert.Error(t, err)
	assert.Equal(t, "required by --user", err.Error())
}

func TestWalkRequiresNotForcedWhenDefaultSatisfies(t *testing.T) {
	// ra's requires has has-value semantics: a default satisfies the
	// constraint, so the walk must not force a value the user can skip.
	a := strArg("user", func(arg *ScriptArg) { arg.RequiresConstraint = []string{"token"} })
	b := strArg("token", withDefaultStr("anon-token"))

	p := &fakePrompter{t: t, answers: []string{"alice", ""}}
	tokens, err := walkInteractiveArgs([]*ScriptArg{a, b}, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--user", "alice"}, tokens, "skipped token falls back to its default")
}

func TestWalkDefaultedArgTriggersRequires(t *testing.T) {
	// Mirrors ra: an untouched defaulted arg still has a value, so its
	// requires constraint will be enforced by the final parse - the walk must
	// force the target arg even though nothing was explicitly set.
	a := strArg("mode", withDefaultStr("fast"), func(arg *ScriptArg) { arg.RequiresConstraint = []string{"token"} })
	b := strArg("token", func(arg *ScriptArg) { arg.IsNullable = true })

	p := &fakePrompter{t: t, answers: []string{"", "s3cret"}}
	tokens, err := walkInteractiveArgs([]*ScriptArg{a, b}, notConfigured, noBoolVals, p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--token", "s3cret"}, tokens)
}

func TestWalkGroupBoolValueDrivesLaterConstraints(t *testing.T) {
	// A bool toggled ON in the multipick excludes a later arg; a bool left at
	// its true default does NOT (ra's excludes only counts explicitly-set
	// flags), even though both end the prompt checked.
	defTrue := true
	args := []*ScriptArg{
		{Name: "json", ExternalName: "json", Type: ArgBoolT, HasDefaultValue: true,
			ExcludesConstraint: []string{"out"}},
		{Name: "cache", ExternalName: "cache", Type: ArgBoolT, HasDefaultValue: true, DefaultBool: &defTrue,
			ExcludesConstraint: []string{"tmp"}},
		strArg("out", func(arg *ScriptArg) { arg.IsNullable = true }),
		strArg("tmp", func(arg *ScriptArg) { arg.IsNullable = true }),
	}
	var notes []string
	notef := func(format string, a ...any) { notes = append(notes, format) }

	// json toggled on (emits --json), cache kept checked (its default, no token).
	p := &fakePrompter{t: t, answers: []string{"/tmp/x"}, multiAnswers: [][]string{{"--json", "--cache"}}}
	tokens, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, notef)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--json", "--tmp", "/tmp/x"}, tokens)
	assert.Len(t, notes, 1, "only out is skipped; cache's kept default cannot exclude")
}

func TestWalkCliBoolFalseDoesNotExclude(t *testing.T) {
	// --json=false on the CLI marks json configured, but ra's excludes only
	// counts a bool that is set AND true - the walk must still prompt for out.
	args := []*ScriptArg{
		{Name: "json", ExternalName: "json", Type: ArgBoolT, HasDefaultValue: true,
			ExcludesConstraint: []string{"out"}},
		strArg("out", func(arg *ScriptArg) { arg.IsNullable = true }),
	}
	p := &fakePrompter{t: t, answers: []string{"x.txt"}}
	tokens, err := walkInteractiveArgs(args,
		func(ext string) bool { return ext == "json" },
		func(ext string) bool { return false }, // parsed value: false
		p, noNotes)

	assert.NoError(t, err)
	assert.Equal(t, []string{"--out", "x.txt"}, tokens)
}

func TestWalkPropagatesNotInteractive(t *testing.T) {
	args := []*ScriptArg{strArg("name")}
	p := &fakePrompter{t: t, err: radish.ErrNotInteractive}
	_, err := walkInteractiveArgs(args, notConfigured, noBoolVals, p, noNotes)

	assert.ErrorIs(t, err, radish.ErrNotInteractive)
}

func TestValidators(t *testing.T) {
	min, max := 1.0, 10.0
	intArg := &ScriptArg{
		Name: "n", ExternalName: "n", Type: ArgIntT,
		RangeConstraint: &ArgRangeConstraint{Min: &min, MinInclusive: true, Max: &max, MaxInclusive: false},
	}
	v := validatorFor(intArg, true, "")
	assert.Error(t, v(""), "required rejects empty")
	assert.Error(t, v("abc"), "non-integer rejected")
	assert.Error(t, v("10"), "exclusive max rejected")
	assert.Error(t, v("0"), "below min rejected")
	assert.NoError(t, v("1"))
	assert.NoError(t, v("9"))

	floatArg := &ScriptArg{Name: "f", ExternalName: "f", Type: ArgFloatT}
	fv := validatorFor(floatArg, false, "")
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

func testTokenInfo() flagTokenInfo {
	return flagTokenInfo{
		valueTaking: map[string]bool{"--name": true, "-n": true, "--color": true},
		shorts:      map[rune]bool{'i': true, 'd': true, 'q': true, 'n': true},
	}
}

func TestStripInteractiveFlags(t *testing.T) {
	info := testTokenInfo()
	assert.Equal(t, []string{"a", "b"}, stripInteractiveFlags([]string{"-i", "a", "--interactive", "b"}, info))
	assert.Equal(t, []string{"a", "--", "-i"}, stripInteractiveFlags([]string{"a", "-i", "--", "-i"}, info),
		"tokens after -- are positional values, not flags")
	assert.Equal(t, []string{"a"}, stripInteractiveFlags([]string{"-i=true", "a", "--interactive=true"}, info),
		"=value forms of the interactive flag are stripped too")
}

func TestStripInteractiveFlagsPreservesFlagValues(t *testing.T) {
	info := testTokenInfo()
	assert.Equal(t, []string{"--name", "-i", "--color", "always"},
		stripInteractiveFlags([]string{"-i", "--name", "-i", "--color", "always"}, info),
		"a -i in value position is data; stripping it would leave --name dangling")
	assert.Equal(t, []string{"-n", "-i"},
		stripInteractiveFlags([]string{"-n", "-i", "-i"}, info),
		"short value-taking flags protect their value the same way")
}

func TestStripInteractiveFlagsHandlesClusters(t *testing.T) {
	info := testTokenInfo()
	assert.Equal(t, []string{"-dq"}, stripInteractiveFlags([]string{"-dqi"}, info),
		"the i rune is removed from a recognized short cluster")
	assert.Equal(t, []string{"-d"}, stripInteractiveFlags([]string{"-di=true"}, info),
		"an =value belonging to the trailing i is dropped with it")
	assert.Equal(t, []string{"-d=x"}, stripInteractiveFlags([]string{"-id=x"}, info),
		"an =value belonging to another short survives")
	assert.Equal(t, []string{"-xi"}, stripInteractiveFlags([]string{"-xi"}, info),
		"a token with unregistered shorts is not a cluster - left alone")
	assert.Equal(t, []string{"-5"}, stripInteractiveFlags([]string{"-5"}, info),
		"negative numbers pass through")
}
