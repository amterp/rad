package prompts

import (
	"fmt"
	"sort"
	"strings"
)

// Outcome is what the runtime should do when it reaches a site.
type Outcome int

const (
	// NoReply means nothing was supplied for this site: prompt if there's a
	// terminal, stop if there isn't.
	NoReply Outcome = iota
	// Answered means an answer was supplied and consumed.
	Answered
	// Unreachable means the caller asserted with --reply-na that this site
	// wouldn't be reached. Reaching it is a clean failure, never a guess.
	Unreachable
	// Exhausted means answers were supplied for this site but all have been
	// used - a loop ran more times than the caller accounted for. Deliberately
	// not "reuse the last answer": one `y` silently approving five hundred
	// deletions is exactly the accident this design exists to prevent.
	Exhausted
)

// Answer is a reply already parsed and validated against its site's kind, so
// the runtime never has to interpret raw text.
type Answer struct {
	// Text carries an input or pick answer.
	Text string
	// Bool carries a confirm or shell-confirm answer.
	Bool bool
	// List carries a multipick selection.
	List []string
}

// Replies holds the command-line answers, bound to the sites they address.
type Replies struct {
	bySite map[string]*siteReplies
	byPos  map[[2]int]*siteReplies
}

type siteReplies struct {
	site        Site
	answers     []Answer
	consumed    int
	unreachable bool
}

// ParseReplies binds --reply and --reply-na values to sites, validating each
// answer against the kind of prompt it addresses.
//
// Everything it can catch, it catches here rather than at the prompt: a stale
// key usually means the script moved under the caller, and finding that out
// before any of the script runs is the whole point.
func ParseReplies(replyArgs, naArgs []string, sites []Site) (*Replies, error) {
	r := &Replies{
		bySite: make(map[string]*siteReplies, len(sites)),
		byPos:  make(map[[2]int]*siteReplies, len(sites)),
	}
	for _, site := range sites {
		entry := &siteReplies{site: site}
		r.bySite[site.Key] = entry
		r.byPos[[2]int{site.Line, site.Col}] = entry
	}

	for _, raw := range replyArgs {
		key, value, ok := strings.Cut(raw, ":")
		if !ok {
			return nil, fmt.Errorf(
				"invalid --reply %q: expected <line>:<value>, for example --reply 14:yes", raw)
		}
		entry, err := r.lookup(key, sites)
		if err != nil {
			return nil, err
		}
		if entry.site.AsValue {
			return nil, fmt.Errorf(
				"cannot answer %s: the script passes it around as a value rather than "+
					"calling it, so rad can't tell where it runs and has nothing to attach "+
					"an answer to. Use --reply-na %s if this run won't invoke it",
				entry.site.describe(), entry.site.Key)
		}
		if entry.site.Secret {
			return nil, fmt.Errorf(
				"cannot answer %s from the command line: arguments are visible to other "+
					"processes on this machine. Run it where there's a terminal, or use "+
					"--reply-na %s if this run won't reach it",
				entry.site.describe(), entry.site.Key)
		}
		answer, err := parseAnswer(entry.site, value)
		if err != nil {
			return nil, err
		}
		entry.answers = append(entry.answers, answer)
	}

	for _, key := range naArgs {
		entry, err := r.lookup(key, sites)
		if err != nil {
			return nil, err
		}
		entry.unreachable = true
	}

	for _, entry := range r.bySite {
		if entry.unreachable && len(entry.answers) > 0 {
			return nil, fmt.Errorf(
				"conflicting flags for %s: --reply supplies an answer while --reply-na "+
					"says it won't be reached. Drop whichever is wrong",
				entry.site.describe())
		}
	}

	return r, nil
}

func (r *Replies) lookup(key string, sites []Site) (*siteReplies, error) {
	if entry, ok := r.bySite[key]; ok {
		return entry, nil
	}

	// A bare line number when that line holds several prompts is a common near
	// miss, so name the real keys instead of just rejecting it.
	var onLine []string
	for _, site := range sites {
		if fmt.Sprintf("%d", site.Line) == key {
			onLine = append(onLine, site.Key)
		}
	}
	if len(onLine) > 0 {
		return nil, fmt.Errorf(
			"line %s holds %d prompts, so it needs a more precise key: %s",
			key, len(onLine), strings.Join(onLine, " or "))
	}

	if key == "" {
		return nil, fmt.Errorf("missing key: expected <line>:<value>, for example --reply 14:yes")
	}

	if len(sites) == 0 {
		return nil, fmt.Errorf(
			"no prompt at %s: this script has no interactive prompts on this code path, "+
				"so there is nothing for --reply to answer", key)
	}
	return nil, fmt.Errorf(
		"no prompt at %s. This script prompts at: %s. "+
			"If the script changed, re-run it without --reply to get the current list",
		key, strings.Join(allKeys(sites), ", "))
}

func allKeys(sites []Site) []string {
	keys := make([]string, 0, len(sites))
	for _, site := range sites {
		keys = append(keys, site.Key)
	}
	return keys
}

// parseAnswer interprets a raw value according to the site's kind. The kind is
// always known statically, so a value never has to be guessed at from its own
// shape - which is why an input answer can be taken completely verbatim.
func parseAnswer(site Site, value string) (Answer, error) {
	switch site.Kind {
	case Confirm, ShellConfirm:
		b, err := parseBool(value)
		if err != nil {
			return Answer{}, fmt.Errorf("invalid answer %q for %s: %v", value, site.describe(), err)
		}
		return Answer{Bool: b}, nil

	case Input:
		return Answer{Text: value}, nil

	case Pick:
		if err := checkOption(site, value); err != nil {
			return Answer{}, err
		}
		return Answer{Text: value}, nil

	case Multipick:
		selections := splitEscaped(value)
		seen := make(map[string]bool, len(selections))
		for _, sel := range selections {
			if err := checkOption(site, sel); err != nil {
				return Answer{}, err
			}
			// The interactive picker toggles, so it can never return the same
			// option twice. A repeat is a caller mistake, and left alone it can
			// satisfy a `min` with fewer distinct choices than it asks for.
			if seen[sel] {
				return Answer{}, fmt.Errorf(
					"%q is selected twice for %s; list each option at most once",
					sel, site.describe())
			}
			seen[sel] = true
		}
		if err := checkBounds(site, len(selections)); err != nil {
			return Answer{}, err
		}
		return Answer{List: selections}, nil

	default:
		return Answer{}, fmt.Errorf("unsupported prompt kind at %s", site.Key)
	}
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes", "true":
		return true, nil
	case "n", "no", "false":
		return false, nil
	default:
		return false, fmt.Errorf("expected yes or no")
	}
}

// checkOption rejects an answer that isn't one of the choices. Only literal
// option lists can be checked here; computed ones are checked at the prompt,
// where the real values are finally known.
func checkOption(site Site, value string) error {
	if site.Options == nil {
		return nil
	}
	for _, opt := range site.Options {
		if opt == value {
			return nil
		}
	}
	return fmt.Errorf(
		"%q is not one of the options for %s. Valid options: %s. "+
			"Answers must match exactly - rad won't guess which one you meant",
		value, site.describe(), strings.Join(site.Options, ", "))
}

// checkBounds rejects a selection count the prompt would not have accepted.
// Only literal bounds are known here; computed ones are checked at the prompt.
func checkBounds(site Site, n int) error {
	if site.MinSet && int64(n) < site.Min {
		return fmt.Errorf(
			"%s needs at least %d selections, but the answer gave %d",
			site.describe(), site.Min, n)
	}
	if site.MaxSet && int64(n) > site.Max {
		return fmt.Errorf(
			"%s allows at most %d selections, but the answer gave %d",
			site.describe(), site.Max, n)
	}
	return nil
}

// describe names a site the way an error message should: by its prompt text
// where the script wrote one, and by its key either way so the reader knows
// which flag to change.
func (s Site) describe() string {
	kind := s.Label()
	if s.Secret {
		kind = "secret " + kind
	}
	if s.AsValue {
		return fmt.Sprintf("the %s used as a value at %s", kind, s.Key)
	}
	if s.Prompt == "" {
		return fmt.Sprintf("the %s at %s", kind, s.Key)
	}
	return fmt.Sprintf("the %s %q at %s", kind, s.Prompt, s.Key)
}

// splitEscaped splits a multipick answer on commas, honoring `\,` for a
// literal comma and `\\` for a literal backslash. A backslash before anything
// else is kept as-is so a Windows path doesn't quietly lose its separators.
func splitEscaped(value string) []string {
	if value == "" {
		// Selecting nothing is a real answer when the prompt allows it, and the
		// only way to write it. Without this it splits to one empty option.
		return nil
	}

	var out []string
	var cur strings.Builder
	escaped := false

	for _, r := range value {
		if escaped {
			if r != ',' && r != '\\' {
				cur.WriteRune('\\')
			}
			cur.WriteRune(r)
			escaped = false
			continue
		}
		switch r {
		case '\\':
			escaped = true
		case ',':
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if escaped {
		cur.WriteRune('\\') // a trailing backslash escapes nothing
	}

	return append(out, cur.String())
}

// Pending reports whether the site at this position still has an answer
// waiting.
//
// A pick given a filter can resolve itself without prompting. When it does and
// an answer is waiting, the answer has to be consumed anyway: the caller counts
// executions, not prompts, and a queue that skips a pass hands every later
// answer to the wrong one. Callers use this to decide whether to look before
// they take that shortcut.
func (r *Replies) Pending(line, col int) bool {
	if r == nil {
		return false
	}
	entry, ok := r.byPos[[2]int{line, col}]
	return ok && !entry.unreachable && entry.consumed < len(entry.answers)
}

// Supplied reports how many answers were given for the site at this position.
//
// It exists so a site that runs out can say how far the answers got. "Ran out
// after 1 pass" is a fact the caller can check against the script; "ran out" on
// its own leaves them guessing whether they miscounted or the script re-asked.
func (r *Replies) Supplied(line, col int) int {
	if r == nil {
		return 0
	}
	entry, ok := r.byPos[[2]int{line, col}]
	if !ok {
		return 0
	}
	return len(entry.answers)
}

// Take consumes the next answer for the site at this position.
//
// Answers are per-site queues rather than one global queue, so inserting a
// prompt earlier in a script can't silently re-target every later answer.
func (r *Replies) Take(line, col int) (Answer, Outcome) {
	if r == nil {
		return Answer{}, NoReply
	}
	entry, ok := r.byPos[[2]int{line, col}]
	if !ok {
		return Answer{}, NoReply
	}
	if entry.unreachable {
		return Answer{}, Unreachable
	}
	if len(entry.answers) == 0 {
		return Answer{}, NoReply
	}
	if entry.consumed >= len(entry.answers) {
		return Answer{}, Exhausted
	}

	answer := entry.answers[entry.consumed]
	entry.consumed++
	return answer, Answered
}

// Unanswered returns the sites with no --reply and no --reply-na, in source
// order. These are what the pre-flight guard reports.
func (r *Replies) Unanswered() []Site {
	if r == nil {
		return nil
	}
	var out []Site
	for _, entry := range r.bySite {
		if entry.unreachable || len(entry.answers) > 0 {
			continue
		}
		out = append(out, entry.site)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		return out[i].Col < out[j].Col
	})
	return out
}
