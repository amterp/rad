package radfmt_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	radtesting "github.com/amterp/rad/core/testing"
	"github.com/stretchr/testify/require"
)

// ruleHeadingRe matches a RULES.md rule heading and captures its ID and status,
// e.g. "### F12 - Assignment spacing `implemented`" -> ("F12", "implemented").
// The status is the trailing backticked word; titles must not contain backticks.
var ruleHeadingRe = regexp.MustCompile("(?m)^###\\s+(F\\d+)\\b.*`(\\w+)`\\s*$")

// looseHeadingRe matches any line that looks like a rule heading (`### F<n> ...`)
// so parseRules can flag headings the strict ruleHeadingRe would silently ignore
// (e.g. trailing text after the status backtick), rather than dropping the rule.
var looseHeadingRe = regexp.MustCompile(`(?m)^###\s+F\d+\b.*$`)

// ruleTagRe matches a rule reference like "[F12]" in code comments or snapshot
// titles.
var ruleTagRe = regexp.MustCompile(`\[(F\d+)\]`)

// TestRuleCoverage is the enforcement arm of the rule system (see RULES.md and
// AGENTS.md). It keeps the spec, the code, and the snapshots in lockstep:
//   - every active rule is tagged in the code and demonstrated by a snapshot
//     (byte-level rules use [raw] snapshots - see snapshot_test.go), and
//   - nothing references a rule ID that RULES.md doesn't define.
//
// A dropped rule, an undocumented tag, or a rule with no demonstrating snapshot
// fails the build - that's what stops the docs from rotting.
func TestRuleCoverage(t *testing.T) {
	rules := parseRules(t)
	codeTags := collectCodeTags(t)
	snapTags := collectSnapshotTags(t)

	for _, id := range sortedRuleIDs(rules) {
		status := rules[id]
		switch status {
		case "implemented", "passthrough":
			require.Containsf(t, codeTags, id,
				"rule %s (%s) has no `// [%s]` code tag - tag its enforcing site", id, status, id)
			require.Containsf(t, snapTags, id,
				"rule %s (%s) has no demonstrating snapshot - add a case titled `[%s] ...`", id, status, id)
		case "limitation", "deferred", "roadmap":
			// Documented only; intentionally exempt from tag/snapshot enforcement.
		default:
			t.Errorf("rule %s has unknown status %q (see the status legend in RULES.md)", id, status)
		}
	}

	// No orphans: every tag in code or snapshots must name a rule RULES.md defines.
	for id := range codeTags {
		require.Containsf(t, rules, id, "code tag [%s] has no matching rule in RULES.md", id)
	}
	for id := range snapTags {
		require.Containsf(t, rules, id, "snapshot tag [%s] has no matching rule in RULES.md", id)
	}
}

// parseRules reads RULES.md and returns a map of rule ID to status.
func parseRules(t *testing.T) map[string]string {
	data, err := os.ReadFile("RULES.md")
	require.NoError(t, err, "read RULES.md")

	out := map[string]string{}
	for _, m := range ruleHeadingRe.FindAllStringSubmatch(string(data), -1) {
		id, status := m[1], m[2]
		if prev, dup := out[id]; dup {
			t.Errorf("rule %s defined twice in RULES.md (statuses %q and %q); IDs are unique", id, prev, status)
		}
		out[id] = status
	}
	require.NotEmpty(t, out, "no rules parsed from RULES.md - did the heading format change?")

	// A heading that looks like a rule but the strict regex rejects (e.g. trailing
	// text after the status backtick) would otherwise vanish silently. Fail loudly
	// so a malformed heading can't quietly exempt a rule from coverage.
	for _, line := range looseHeadingRe.FindAllString(string(data), -1) {
		if !ruleHeadingRe.MatchString(line) {
			t.Errorf("malformed rule heading in RULES.md (status must be the only "+
				"backticked word, with nothing after it): %q", line)
		}
	}
	return out
}

// collectCodeTags scans the package's non-test Go sources for `[Fn]` tags.
func collectCodeTags(t *testing.T) map[string]bool {
	files, err := filepath.Glob("*.go")
	require.NoError(t, err)

	out := map[string]bool{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		require.NoError(t, err)
		for _, m := range ruleTagRe.FindAllStringSubmatch(string(data), -1) {
			out[m[1]] = true
		}
	}
	return out
}

// collectSnapshotTags scans every snapshot title for `[Fn]` tags.
func collectSnapshotTags(t *testing.T) map[string]bool {
	files, err := filepath.Glob("snapshots/*.snap")
	require.NoError(t, err)

	out := map[string]bool{}
	for _, f := range files {
		cases, err := radtesting.ParseSnapshotFile(f)
		require.NoErrorf(t, err, "parse snapshot file %s", f)
		for _, c := range cases {
			for _, m := range ruleTagRe.FindAllStringSubmatch(c.Title, -1) {
				out[m[1]] = true
			}
		}
	}
	return out
}

func sortedRuleIDs(rules map[string]string) []string {
	out := make([]string, 0, len(rules))
	for id := range rules {
		out = append(out, id)
	}
	// Sort numerically (F2 before F10) so coverage failures read in rule order.
	sort.Slice(out, func(i, j int) bool { return ruleNum(out[i]) < ruleNum(out[j]) })
	return out
}

// ruleNum extracts the integer from a rule ID like "F12" -> 12 for ordering.
func ruleNum(id string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(id, "F"))
	return n
}
