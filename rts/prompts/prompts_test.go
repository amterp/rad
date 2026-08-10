package prompts_test

import (
	"testing"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/prompts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func find(t *testing.T, src string, cmdPath ...string) []prompts.Site {
	t.Helper()
	parser, err := rts.NewRadParser()
	require.NoError(t, err)
	defer parser.Close()

	tree := parser.Parse(src)
	file := rts.ConvertCST(tree.Root(), src, "prompts_test.rad")
	require.NotNil(t, file, "source failed to convert; is it valid Rad?")

	resolved := check.Resolve(file)
	require.NotNil(t, resolved)

	return prompts.Find(file, resolved, cmdPath)
}

func TestFindsEachInteractiveBuiltin(t *testing.T) {
	sites := find(t, `x = input("Name?")
y = confirm("Sure?")
z = pick(["a", "b"])
w = multipick(["c", "d"])
`)

	require.Len(t, sites, 4)
	assert.Equal(t, prompts.Input, sites[0].Kind)
	assert.Equal(t, "Name?", sites[0].Prompt)
	assert.Equal(t, prompts.Confirm, sites[1].Kind)
	assert.Equal(t, "Sure?", sites[1].Prompt)
	assert.Equal(t, prompts.Pick, sites[2].Kind)
	assert.Equal(t, []string{"a", "b"}, sites[2].Options)
	assert.Equal(t, prompts.Multipick, sites[3].Kind)
	assert.Equal(t, []string{"c", "d"}, sites[3].Options)
}

// The pre-flight guard hard-blocks, so mistaking a local for the builtin would
// block a script that never prompts at all.
func TestShadowedBuiltinIsNotASite(t *testing.T) {
	sites := find(t, `fn pick(opts):
    return opts[0]

x = pick(["a", "b"])
`)

	assert.Empty(t, sites)
}

func TestKeysUseLineAndWidenOnlyWhenAmbiguous(t *testing.T) {
	sites := find(t, `a = input("one")
b = confirm("two") and confirm("three")
`)

	require.Len(t, sites, 3)
	assert.Equal(t, "1", sites[0].Key, "a lone site on a line keys by line alone")
	assert.Equal(t, 2, sites[1].Line)
	assert.Equal(t, 2, sites[2].Line)
	assert.NotEqual(t, sites[1].Key, sites[2].Key, "two sites on one line must differ")
	assert.Contains(t, sites[1].Key, "2.")
	assert.Contains(t, sites[2].Key, "2.")
}

func TestRuntimeValuesAreReportedAsUnknown(t *testing.T) {
	sites := find(t, `servers = ["x", "y"]
name = "bob"
a = pick(servers)
b = input("Hi {name}")
c = pick(["p", name])
`)

	require.Len(t, sites, 3)
	assert.Nil(t, sites[0].Options, "a computed option list is not statically known")
	assert.Empty(t, sites[1].Prompt, "an interpolated prompt is not statically known")
	assert.Nil(t, sites[2].Options, "one computed entry makes the whole list unknown")
}

func TestSecretInputIsFlagged(t *testing.T) {
	sites := find(t, `a = input("Password", secret=true)
b = input("Name")
`)

	require.Len(t, sites, 2)
	assert.True(t, sites[0].Secret)
	assert.False(t, sites[1].Secret)
}

func TestFilterArgumentIsRecordedPerBuiltin(t *testing.T) {
	sites := find(t, `a = pick(["x", "y"], "x")
b = pick(["x", "y"])
c = pick_kv(["x"], [1], "x")
d = pick_from_resource("s.json", "x")
`)

	require.Len(t, sites, 4)
	assert.True(t, sites[0].Filtered, "pick takes its filter second")
	assert.False(t, sites[1].Filtered)
	assert.True(t, sites[2].Filtered, "pick_kv takes its filter third")
	assert.True(t, sites[3].Filtered, "pick_from_resource takes its filter second")
	assert.Equal(t, []string{"x"}, sites[2].Options, "pick_kv shows its keys")
	assert.Nil(t, sites[3].Options, "resource contents are never statically known")
}

func TestConfirmGatedShellCommandIsASite(t *testing.T) {
	sites := find(t, "confirm $`rm -rf /tmp/x`\n")

	require.Len(t, sites, 1)
	assert.Equal(t, prompts.ShellConfirm, sites[0].Kind)
	assert.Equal(t, 1, sites[0].Line)
}

func TestPlainShellCommandIsNotASite(t *testing.T) {
	sites := find(t, "$`echo hi`\n")

	assert.Empty(t, sites)
}

func TestOnlyReachableFunctionsAreWalked(t *testing.T) {
	sites := find(t, `fn used():
    return confirm("reachable")

fn unused():
    return confirm("unreachable")

x = used()
`)

	require.Len(t, sites, 1)
	assert.Equal(t, "reachable", sites[0].Prompt)
}

func TestCallsAreFollowedTransitively(t *testing.T) {
	sites := find(t, `fn outer():
    return middle()

fn middle():
    return input("deep")

x = outer()
`)

	require.Len(t, sites, 1)
	assert.Equal(t, "deep", sites[0].Prompt)
}

func TestRecursionTerminates(t *testing.T) {
	sites := find(t, `fn ping(n):
    yes = confirm("again?")
    return ping(n) if yes else n

x = ping(1)
`)

	require.Len(t, sites, 1)
	assert.Equal(t, "again?", sites[0].Prompt)
}

// Without command scoping, running one command would demand answers for every
// prompt in every other command in the file.
func TestOnlyTheInvokedCommandsPromptsAreFound(t *testing.T) {
	src := `args:
    x int = 1

command deploy:
    calls do_deploy

command status:
    calls do_status

fn do_deploy():
    print(confirm("Deploy?"))

fn do_status():
    print(input("Which env?"))
`

	deploy := find(t, src, "deploy")
	require.Len(t, deploy, 1)
	assert.Equal(t, "Deploy?", deploy[0].Prompt)

	status := find(t, src, "status")
	require.Len(t, status, 1)
	assert.Equal(t, "Which env?", status[0].Prompt)

	none := find(t, src)
	assert.Empty(t, none, "no command selected means only top-level code runs")
}

func TestTopLevelCodeIsAlwaysInScope(t *testing.T) {
	sites := find(t, `command deploy:
    calls do_deploy

shared = confirm("Shared?")

fn do_deploy():
    print(input("Env?"))
`, "deploy")

	require.Len(t, sites, 2)
	assert.Equal(t, "Shared?", sites[0].Prompt)
	assert.Equal(t, "Env?", sites[1].Prompt)
}

func TestNilInputsAreSafe(t *testing.T) {
	assert.Nil(t, prompts.Find(nil, nil, nil))
}

// Each of the tests below pins a case that once slipped through: a prompt the
// walk could not see, or one it saw but described wrongly. Every one of them
// ends the same way if it regresses - the script starts running, does work, and
// then stops at a prompt nobody can answer.

func TestFunctionUsedAsAValueIsStillWalked(t *testing.T) {
	sites := find(t, `fn helper():
    return confirm("Delete everything?")

h = helper
ok = h()
`)

	require.Len(t, sites, 1, "a function handed around as a value can still be called")
	assert.Equal(t, 2, sites[0].Line)
	assert.True(t, sites[0].Repeats, "rad can't see how often a loose reference is called")
}

func TestFunctionPassedAsAnArgumentIsStillWalked(t *testing.T) {
	sites := find(t, `fn helper(x):
    return confirm("Keep {x}?")

out = [1, 2].map(helper)
`)

	require.Len(t, sites, 1)
	assert.Equal(t, 2, sites[0].Line)
}

func TestNestedFunctionsWithTheSameNameDoNotCollide(t *testing.T) {
	sites := find(t, `fn outer():
    fn helper():
        return input("INNER? ")
    return helper()

fn helper():
    return input("OUTER? ")

print(outer())
`)

	require.Len(t, sites, 1, "only the nested helper is reachable from outer()")
	assert.Equal(t, 3, sites[0].Line)
	assert.Equal(t, "INNER? ", sites[0].Prompt)
}

func TestUfcsCallsShiftPositionalArguments(t *testing.T) {
	// The interpreter prepends the receiver, so every positional here is one to
	// the right of where it is written. Read literally, pick_kv reports its
	// values as the labels an answer has to match - which no answer ever will.
	sites := find(t, `v = ["Alpha", "Beta"].pick_kv(["a", "b"])
w = ["x", "y"].pick("x")
`)

	require.Len(t, sites, 2)
	assert.Equal(t, []string{"Alpha", "Beta"}, sites[0].Options, "the keys are the labels")
	assert.True(t, sites[1].Filtered, "the filter is the second argument after the receiver")
}

func TestPromptsInLoopsAreMarkedRepeating(t *testing.T) {
	sites := find(t, `for x in ["a", "b"]:
    ok = confirm("Delete {x}?")

y = confirm("Once?")
`)

	require.Len(t, sites, 2)
	assert.True(t, sites[0].Repeats, "a prompt in a loop body asks once per pass")
	assert.False(t, sites[1].Repeats)
}

func TestLoopIterableRunsOnceSoIsNotMarkedRepeating(t *testing.T) {
	sites := find(t, `for x in pick(["a", "b"]):
    print(x)
`)

	require.Len(t, sites, 1)
	assert.False(t, sites[0].Repeats, "the iterable is evaluated once, before the loop")
}

func TestRepeatingSpreadsThroughCalledFunctions(t *testing.T) {
	sites := find(t, `fn ask():
    return confirm("Sure?")

for x in ["a", "b"]:
    ask()
`)

	require.Len(t, sites, 1)
	assert.True(t, sites[0].Repeats, "called from a loop, so its prompt repeats too")
}

func TestFunctionCalledFromSeveralPlacesRepeats(t *testing.T) {
	sites := find(t, `fn ask():
    return confirm("Sure?")

ask()
ask()
`)

	require.Len(t, sites, 1)
	assert.True(t, sites[0].Repeats)
}

func TestDefaultPromptsAreReportedButComputedOnesAreNot(t *testing.T) {
	sites := find(t, `a = input()
b = confirm()
c = input(prompt="Named")
`)

	require.Len(t, sites, 3)
	assert.Equal(t, "> ", sites[0].Prompt, "rad knows exactly what a bare input shows")
	assert.Equal(t, "Confirm? [Y/n] > ", sites[1].Prompt)
	assert.Equal(t, "Named", sites[2].Prompt, "a positional passed by name is still literal")
}

func TestMultipickBoundsAreCapturedWhenLiteral(t *testing.T) {
	sites := find(t, `a = multipick(["x", "y", "z"], min=2, max=2)
b = multipick(["x", "y"])
`)

	require.Len(t, sites, 2)
	assert.True(t, sites[0].MinSet)
	assert.Equal(t, int64(2), sites[0].Min)
	assert.True(t, sites[0].MaxSet)
	assert.Equal(t, int64(2), sites[0].Max)
	assert.False(t, sites[1].MinSet)
	assert.False(t, sites[1].MaxSet)
}

func TestCommandCallbackRunsOnceAndIsNotMarkedRepeating(t *testing.T) {
	// `calls do_deploy` invokes it exactly once. Reading it as a function passed
	// around as a value would flag every prompt inside as repeating - wrong
	// advice on the most ordinary shape a command script has.
	sites := find(t, `command deploy:
    calls do_deploy

fn do_deploy():
    c = confirm("Deploy? ")
`, "deploy")

	require.Len(t, sites, 1)
	assert.False(t, sites[0].Repeats)
}

func TestPromptsInsideLambdasRepeat(t *testing.T) {
	// A lambda has no name to count call sites for, and map runs it per element.
	sites := find(t, `out = [1, 2].map(fn(x) input("For {x}? "))
`)

	require.Len(t, sites, 1)
	assert.True(t, sites[0].Repeats)
}

func TestLambdaCommandCallbackDoesNotRepeat(t *testing.T) {
	// The lambda form of a callback runs once, like the named form - unlike the
	// lambdas handed to map and filter that the walk flags as repeating.
	sites := find(t, `command deploy:
    calls fn():
        c = confirm("Deploy? ")
        print(c)
`, "deploy")

	require.Len(t, sites, 1)
	assert.False(t, sites[0].Repeats, "a command callback body runs once")
}

func TestInteractiveBuiltinUsedAsAValueIsASiteMarkedAsValue(t *testing.T) {
	// There is no call node to key an answer to, so it can't take a --reply -
	// but it still gets a key, so --reply-na can assert it is never invoked.
	sites := find(t, `f = pick
print(f(["alpha", "beta"]))
`)

	require.Len(t, sites, 1)
	assert.True(t, sites[0].AsValue)
	assert.Equal(t, "pick", sites[0].Fn)
	assert.Equal(t, "1", sites[0].Key)
}

func TestInteractiveBuiltinPassedToMapIsMarkedAsValue(t *testing.T) {
	sites := find(t, `out = [["a", "b"]].map(pick)
`)
	require.Len(t, sites, 1)
	assert.True(t, sites[0].AsValue)
}

func TestOrdinaryBuiltinUsedAsAValueIsNotASite(t *testing.T) {
	// Only the ones that stop to ask the user something matter here.
	assert.Empty(t, find(t, `f = upper
print(f("abc"))
`))
}

func TestCallingAnInteractiveBuiltinNormallyIsNotAsValue(t *testing.T) {
	sites := find(t, `x = pick(["a", "b"])
`)
	require.Len(t, sites, 1)
	assert.False(t, sites[0].AsValue)
}

func TestLiteralFilterIsCapturedForSuggestions(t *testing.T) {
	sites := find(t, `a = pick(["zulu", "beta"], "a")
b = pick(["x", "y"], ["p", "q"])
c = pick(["x", "y"])
`)

	require.Len(t, sites, 3)
	assert.Equal(t, []string{"a"}, sites[0].Filter)
	assert.Equal(t, []string{"p", "q"}, sites[1].Filter)
	assert.Nil(t, sites[2].Filter)
}

func TestUfcsReceiverThatIsNotLiteralLeavesThePromptUnknown(t *testing.T) {
	// The receiver occupies the prompt slot, so falling back to the signature
	// default would report a prompt the user is never going to see.
	sites := find(t, `ps = ["Real prompt"]
a = ps[0].input()
`)

	require.Len(t, sites, 1)
	assert.Empty(t, sites[0].Prompt)
}
