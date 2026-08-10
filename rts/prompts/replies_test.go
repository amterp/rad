package prompts_test

import (
	"testing"

	"github.com/amterp/rad/rts/prompts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func site(kind prompts.Kind, key string, line int, opts ...string) prompts.Site {
	return prompts.Site{Kind: kind, Key: key, Line: line, Col: 1, Options: opts}
}

func TestReplyValuesParseByKind(t *testing.T) {
	sites := []prompts.Site{
		site(prompts.Confirm, "1", 1),
		site(prompts.Input, "2", 2),
		site(prompts.Pick, "3", 3, "a", "b"),
		site(prompts.Multipick, "4", 4, "x", "y"),
	}

	r, err := prompts.ParseReplies(
		[]string{"1:yes", "2:https://example.com:8080", "3:b", "4:x,y"}, nil, sites)
	require.NoError(t, err)

	a, out := r.Take(1, 1)
	assert.Equal(t, prompts.Answered, out)
	assert.True(t, a.Bool)

	a, _ = r.Take(2, 1)
	assert.Equal(t, "https://example.com:8080", a.Text,
		"only the first colon separates key from value")

	a, _ = r.Take(3, 1)
	assert.Equal(t, "b", a.Text)

	a, _ = r.Take(4, 1)
	assert.Equal(t, []string{"x", "y"}, a.List)
}

func TestMultipickEscaping(t *testing.T) {
	sites := []prompts.Site{site(prompts.Multipick, "1", 1)}

	r, err := prompts.ParseReplies([]string{`1:a\,b,c`}, nil, sites)
	require.NoError(t, err)
	a, _ := r.Take(1, 1)
	assert.Equal(t, []string{"a,b", "c"}, a.List, `\, is a literal comma`)

	r, err = prompts.ParseReplies([]string{`1:a\\b`}, nil, sites)
	require.NoError(t, err)
	a, _ = r.Take(1, 1)
	assert.Equal(t, []string{`a\b`}, a.List, `\\ is a literal backslash`)

	r, err = prompts.ParseReplies([]string{`1:C:\Users\x`}, nil, sites)
	require.NoError(t, err)
	a, _ = r.Take(1, 1)
	assert.Equal(t, []string{`C:\Users\x`}, a.List,
		"a backslash before an ordinary char is kept, so paths survive")
}

func TestExactMatchingIsEnforcedAgainstKnownOptions(t *testing.T) {
	sites := []prompts.Site{site(prompts.Pick, "1", 1, "alpha", "beta")}

	_, err := prompts.ParseReplies([]string{"1:alph"}, nil, sites)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not one of the options")
	assert.Contains(t, err.Error(), "alpha, beta")
}

func TestComputedOptionsAreNotCheckedUpFront(t *testing.T) {
	sites := []prompts.Site{{Kind: prompts.Pick, Key: "1", Line: 1, Col: 1}}

	r, err := prompts.ParseReplies([]string{"1:anything"}, nil, sites)
	require.NoError(t, err, "runtime options can only be checked at the prompt")
	a, _ := r.Take(1, 1)
	assert.Equal(t, "anything", a.Text)
}

func TestRepeatedKeyFormsAnOrderedQueuePerSite(t *testing.T) {
	sites := []prompts.Site{site(prompts.Confirm, "1", 1)}

	r, err := prompts.ParseReplies([]string{"1:y", "1:n"}, nil, sites)
	require.NoError(t, err)

	a, out := r.Take(1, 1)
	assert.Equal(t, prompts.Answered, out)
	assert.True(t, a.Bool)

	a, out = r.Take(1, 1)
	assert.Equal(t, prompts.Answered, out)
	assert.False(t, a.Bool)

	_, out = r.Take(1, 1)
	assert.Equal(t, prompts.Exhausted, out,
		"running out must stop, not silently reuse the last answer")
}

func TestReplyNaMarksSiteUnreachable(t *testing.T) {
	sites := []prompts.Site{site(prompts.Confirm, "1", 1)}

	r, err := prompts.ParseReplies(nil, []string{"1"}, sites)
	require.NoError(t, err)

	_, out := r.Take(1, 1)
	assert.Equal(t, prompts.Unreachable, out)
}

func TestReplyAndReplyNaTogetherIsRejected(t *testing.T) {
	sites := []prompts.Site{site(prompts.Confirm, "1", 1)}

	_, err := prompts.ParseReplies([]string{"1:y"}, []string{"1"}, sites)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicting flags")
}

func TestStaleKeyIsRejectedWithTheRealKeys(t *testing.T) {
	sites := []prompts.Site{site(prompts.Confirm, "7", 7)}

	_, err := prompts.ParseReplies([]string{"3:y"}, nil, sites)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no prompt at 3")
	assert.Contains(t, err.Error(), "7")
}

func TestBareLineKeyNamesThePreciseKeys(t *testing.T) {
	sites := []prompts.Site{
		{Kind: prompts.Confirm, Key: "5.1", Line: 5, Col: 1},
		{Kind: prompts.Confirm, Key: "5.20", Line: 5, Col: 20},
	}

	_, err := prompts.ParseReplies([]string{"5:y"}, nil, sites)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "5.1")
	assert.Contains(t, err.Error(), "5.20")
}

func TestSecretInputCannotBeAnswered(t *testing.T) {
	sites := []prompts.Site{{Kind: prompts.Input, Key: "1", Line: 1, Col: 1, Secret: true}}

	_, err := prompts.ParseReplies([]string{"1:hunter2"}, nil, sites)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "visible to other processes")

	_, err = prompts.ParseReplies(nil, []string{"1"}, sites)
	assert.NoError(t, err, "--reply-na still passes a secret site on a dead branch")
}

func TestMalformedReplyIsRejected(t *testing.T) {
	sites := []prompts.Site{site(prompts.Confirm, "1", 1)}

	_, err := prompts.ParseReplies([]string{"1"}, nil, sites)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "<line>:<value>")
}

func TestInvalidBoolIsRejected(t *testing.T) {
	sites := []prompts.Site{site(prompts.Confirm, "1", 1)}

	_, err := prompts.ParseReplies([]string{"1:maybe"}, nil, sites)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected yes or no")
}

func TestUnansweredExcludesAnsweredAndUnreachable(t *testing.T) {
	sites := []prompts.Site{
		site(prompts.Confirm, "1", 1),
		site(prompts.Confirm, "2", 2),
		site(prompts.Confirm, "3", 3),
	}

	r, err := prompts.ParseReplies([]string{"1:y"}, []string{"2"}, sites)
	require.NoError(t, err)

	unanswered := r.Unanswered()
	require.Len(t, unanswered, 1)
	assert.Equal(t, "3", unanswered[0].Key)
}

func TestNilRepliesIsSafe(t *testing.T) {
	var r *prompts.Replies
	_, out := r.Take(1, 1)
	assert.Equal(t, prompts.NoReply, out)
	assert.Nil(t, r.Unanswered())
}

func TestEmptyMultipickAnswerSelectsNothing(t *testing.T) {
	// The only way to write "select nothing", which a min=0 prompt accepts and
	// a user at a terminal can do by just submitting.
	site := prompts.Site{Kind: prompts.Multipick, Key: "1", Line: 1, Col: 1,
		Options: []string{"a", "b"}}
	r, err := prompts.ParseReplies([]string{"1:"}, nil, []prompts.Site{site})
	require.NoError(t, err)

	answer, outcome := r.Take(1, 1)
	assert.Equal(t, prompts.Answered, outcome)
	assert.Empty(t, answer.List)
}

func TestMultipickRejectsDuplicateSelections(t *testing.T) {
	// The picker toggles, so it can never return a duplicate - and a duplicate
	// would otherwise satisfy min=2 with a single distinct choice.
	site := prompts.Site{Kind: prompts.Multipick, Key: "1", Line: 1, Col: 1,
		Options: []string{"a", "b"}, Min: 2, MinSet: true}
	_, err := prompts.ParseReplies([]string{"1:a,a"}, nil, []prompts.Site{site})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "selected twice")
}

func TestMultipickBoundsAreCheckedUpFront(t *testing.T) {
	site := prompts.Site{Kind: prompts.Multipick, Key: "1", Line: 1, Col: 1,
		Options: []string{"a", "b", "c"}, Min: 2, MinSet: true, Max: 2, MaxSet: true}

	_, err := prompts.ParseReplies([]string{"1:a"}, nil, []prompts.Site{site})
	require.Error(t, err, "below min should fail before the script runs, not during")
	assert.Contains(t, err.Error(), "at least 2")

	_, err = prompts.ParseReplies([]string{"1:a,b,c"}, nil, []prompts.Site{site})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at most 2")
}

func TestPendingReportsAnAnswerWithoutConsumingIt(t *testing.T) {
	// A filtered pick uses this to keep its queue in step with executions even
	// on the passes where it resolves itself and asks nothing.
	site := prompts.Site{Kind: prompts.Pick, Key: "1", Line: 1, Col: 1}
	r, err := prompts.ParseReplies([]string{"1:a", "1:b"}, nil, []prompts.Site{site})
	require.NoError(t, err)

	assert.True(t, r.Pending(1, 1))
	assert.True(t, r.Pending(1, 1), "peeking must not consume")

	first, _ := r.Take(1, 1)
	assert.Equal(t, "a", first.Text)
	assert.True(t, r.Pending(1, 1))

	r.Take(1, 1)
	assert.False(t, r.Pending(1, 1), "queue is empty now")
}

func TestPendingIsFalseForAnUnreachableSite(t *testing.T) {
	site := prompts.Site{Kind: prompts.Pick, Key: "1", Line: 1, Col: 1}
	r, err := prompts.ParseReplies(nil, []string{"1"}, []prompts.Site{site})
	require.NoError(t, err)

	assert.False(t, r.Pending(1, 1), "--reply-na supplies no answer to peek at")
	_, outcome := r.Take(1, 1)
	assert.Equal(t, prompts.Unreachable, outcome)
}

func TestAmbiguousKeyNamesTheAlternativesInColumnOrder(t *testing.T) {
	sites := []prompts.Site{
		{Kind: prompts.Confirm, Key: "1.5", Line: 1, Col: 5},
		{Kind: prompts.Confirm, Key: "1.20", Line: 1, Col: 20},
	}
	_, err := prompts.ParseReplies([]string{"1:yes"}, nil, sites)
	require.Error(t, err)
	// Source order, not lexicographic - "1.20 or 1.5" reads as a mistake.
	assert.Contains(t, err.Error(), "1.5 or 1.20")
}
