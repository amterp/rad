package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/amterp/color"
	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

const (
	defaultScriptsMapCap = 256
	scannerBufferSize    = 64 * 1024
	// Maximum size for a single JSON log line when parsing invocation logs.
	// This protects against malformed/corrupt entries consuming excessive memory.
	maxScanLineSize = 1 * 1024 * 1024 // 1MB
)

// scriptInfo holds information about a script from invocation logs
type scriptInfo struct {
	Path           string
	LastOccurrence int64 // epoch millis
}

// diagCounts holds per-severity diagnostic counts for a single script.
type diagCounts struct {
	Errors   int
	Warnings int
	Infos    int
	Hints    int
}

// scriptResult collects the check outcome for a single script.
type scriptResult struct {
	Path    string
	Counts  diagCounts
	Ok      bool // false = couldn't load/parse the file
	Skipped bool // file not found
}

// FuncInternalCheckFromLogs implements _rad_check_from_logs, which checks rad scripts
// from invocation logs for syntax/semantic errors. This is an internal function used
// by the rad CLI for bulk checking recently-used scripts.
var FuncInternalCheckFromLogs = BuiltInFunc{
	Name: INTERNAL_FUNC_CHECK_FROM_LOGS,
	Execute: func(f FuncInvocation) RadValue {
		// Parse duration (ParseDuration handles "all" → 0)
		raw := strings.TrimSpace(f.GetStr("_duration").Plain())
		d, err := ParseDuration(raw)
		if err != nil {
			return f.ReturnErrf(rl.ErrInvalidCheckDuration, "Invalid duration: %s", err.Error())
		}
		if d < 0 {
			return f.ReturnErrf(rl.ErrInvalidCheckDuration, "Duration cannot be negative: %s", raw)
		}
		durationMillis := d.Milliseconds()

		// Get verbose flag
		verbose := f.GetBool("_verbose")

		// read and parse invocation logs
		scripts := parseInvocationLogs(f.i, f.callNode, durationMillis)

		// sort by most recent usage
		sortScriptsByLastOccurrence(scripts)

		// header
		durationDesc := raw
		if durationMillis == 0 {
			durationDesc = "all time"
		}
		RP.Printf("Found %d scripts in invocation logs (checking within %s)...\n\n", len(scripts), durationDesc)

		// prepare checker once (faster than per-file)
		chk, err := check.NewChecker()
		if err != nil {
			// fall back to per-file creation if global init fails
			chk = nil
			RP.RadStderrf("Warning! Failed to init checker once, will init per file: %v\n", err)
		}

		// First pass: collect results and determine max path width for alignment
		results := make([]scriptResult, 0, len(scripts))
		maxPathLen := 0
		for _, script := range scripts {
			if !com.FileExists(script.Path) {
				results = append(results, scriptResult{Path: script.Path, Skipped: true})
				continue
			}
			counts, ok := checkScriptWith(script.Path, chk)
			r := scriptResult{Path: script.Path, Counts: counts, Ok: ok}
			results = append(results, r)
			if len(script.Path) > maxPathLen {
				maxPathLen = len(script.Path)
			}
		}

		// Column headers and their widths
		type col struct {
			Header string
			Width  int
		}
		cols := []col{
			{"Errors", 6},
			{"Warns", 5},
			{"Info", 5},
			{"Hints", 5},
		}

		faint := color.New(color.Faint)
		green := color.New(color.FgGreen)
		red := color.New(color.FgRed)
		yellow := color.New(color.FgYellow)
		cyan := color.New(color.FgCyan)
		blue := color.New(color.FgBlue)

		// Table width: "  X " (4 chars) + path + per-column (2-space gap + width)
		colsWidth := 0
		for _, c := range cols {
			colsWidth += 2 + c.Width
		}
		tableWidth := 4 + maxPathLen + colsWidth

		// Only show table header when there are non-skipped results.
		// Skipped rows don't have count columns, so the header would be orphaned.
		hasDataRows := false
		for _, r := range results {
			if !r.Skipped {
				hasDataRows = true
				break
			}
		}
		if hasDataRows {
			headerLine := fmt.Sprintf("    %-*s", maxPathLen, "")
			for _, c := range cols {
				headerLine += fmt.Sprintf("  %*s", c.Width, c.Header)
			}
			RP.Printf("%s\n", faint.Sprint(headerLine))
		}

		totalChecked, passed, failed, skipped, loadErrors := 0, 0, 0, 0, 0
		var totalCounts diagCounts

		for _, r := range results {
			if r.Skipped {
				skipped++
				if verbose {
					RP.Printf("  %s\n", faint.Sprintf("- %s (file not found)", r.Path))
				}
				continue
			}

			totalChecked++
			totalCounts.Errors += r.Counts.Errors
			totalCounts.Warnings += r.Counts.Warnings
			totalCounts.Infos += r.Counts.Infos
			totalCounts.Hints += r.Counts.Hints

			isFail := !r.Ok || r.Counts.Errors > 0
			if isFail {
				failed++
			} else {
				passed++
			}
			if !r.Ok {
				loadErrors++
			}

			// Status icon
			var icon string
			if isFail {
				icon = red.Sprint("✗")
			} else {
				icon = green.Sprint("✓")
			}

			// Path, left-aligned
			line := fmt.Sprintf("  %s %-*s", icon, maxPathLen, r.Path)

			if r.Ok {
				line += formatCount(r.Counts.Errors, cols[0].Width, red, faint)
				line += formatCount(r.Counts.Warnings, cols[1].Width, yellow, faint)
				line += formatCount(r.Counts.Infos, cols[2].Width, cyan, faint)
				line += formatCount(r.Counts.Hints, cols[3].Width, blue, faint)
			} else {
				line += "  " + faint.Sprint("(load error)")
			}

			RP.Printf("%s\n", line)
		}

		// Dynamic-width separator
		sepWidth := tableWidth
		if sepWidth < 40 {
			sepWidth = 40
		}
		RP.Printf("\n%s\n", strings.Repeat("─", sepWidth))

		// Build summary with colored numbers
		bold := color.New(color.Bold)

		summary := fmt.Sprintf("Checked %d scripts: %s passed, %s failed",
			totalChecked,
			green.Sprintf("%d", passed),
			red.Sprintf("%d", failed))

		// Append non-zero diagnostic/status details in a single parenthetical
		var detailParts []string
		if totalCounts.Errors > 0 {
			noun := "errors"
			if totalCounts.Errors == 1 {
				noun = "error"
			}
			detailParts = append(detailParts, red.Sprintf("%d %s", totalCounts.Errors, noun))
		}
		if loadErrors > 0 {
			noun := "load errors"
			if loadErrors == 1 {
				noun = "load error"
			}
			detailParts = append(detailParts, faint.Sprintf("%d %s", loadErrors, noun))
		}
		if totalCounts.Warnings > 0 {
			noun := "warnings"
			if totalCounts.Warnings == 1 {
				noun = "warning"
			}
			detailParts = append(detailParts, yellow.Sprintf("%d %s", totalCounts.Warnings, noun))
		}
		if totalCounts.Infos > 0 {
			noun := "info diagnostics"
			if totalCounts.Infos == 1 {
				noun = "info diagnostic"
			}
			detailParts = append(detailParts, cyan.Sprintf("%d %s", totalCounts.Infos, noun))
		}
		if totalCounts.Hints > 0 {
			noun := "hints"
			if totalCounts.Hints == 1 {
				noun = "hint"
			}
			detailParts = append(detailParts, blue.Sprintf("%d %s", totalCounts.Hints, noun))
		}
		if skipped > 0 {
			detailParts = append(detailParts, faint.Sprintf("%d skipped", skipped))
		}
		if len(detailParts) > 0 {
			summary += " (" + strings.Join(detailParts, ", ") + ")"
		}

		RP.Printf("%s.\n", bold.Sprint(summary))

		if failed > 0 {
			RExit.Exit(1)
		} else {
			RExit.Exit(0)
		}
		return VOID_SENTINEL
	},
}

// parseInvocationLogs reads all invocation log files and returns unique scripts,
// filtered by durationMillis (0 => no filter). Ensures files are closed promptly.
func parseInvocationLogs(i *Interpreter, node rl.Node, durationMillis int64) []scriptInfo {
	logsDir := GetInvocationLogsDir()

	matches, err := filepath.Glob(filepath.Join(logsDir, "invocations*.jsonl"))
	if err != nil {
		i.emitErrorf(rl.ErrFileRead, node, "Failed to find invocation log files: %s", err.Error())
	}
	if len(matches) == 0 {
		return []scriptInfo{}
	}

	var cutoffTime int64
	if durationMillis > 0 {
		cutoffTime = RClock.Now().UnixMilli() - durationMillis
	}

	scriptsMap := make(map[string]*scriptInfo, defaultScriptsMapCap)

	for _, logPath := range matches {
		if err := processLogFile(logPath, cutoffTime, scriptsMap); err != nil {
			// keep going; this is best-effort
			RP.RadStderrf("Warning! %v\n", err)
		}
	}

	// map -> slice
	out := make([]scriptInfo, 0, len(scriptsMap))
	for _, info := range scriptsMap {
		out = append(out, *info)
	}
	return out
}

// processLogFile encapsulates open/scan/close so we can safely defer Close
// inside this helper without leaking FDs across the caller's loop.
func processLogFile(logPath string, cutoffTime int64, scriptsMap map[string]*scriptInfo) error {
	f, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("Failed to open log file %s: %w", logPath, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			RP.RadStderrf("Warning! Closing %s: %v\n", logPath, cerr)
		}
	}()

	sc := bufio.NewScanner(f)
	// Increase max token size to handle very long JSON lines
	sc.Buffer(make([]byte, scannerBufferSize), maxScanLineSize)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		var entry InvocationLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			RP.RadStderrf("Warning! Skipping corrupt log entry in %s: %v\n", logPath, err)
			continue
		}

		// Filter by time if requested
		if cutoffTime > 0 && entry.EpochMillis < cutoffTime {
			continue
		}

		if entry.ScriptPath == "" {
			continue
		}

		// Update or insert script info
		if existing, ok := scriptsMap[entry.ScriptPath]; ok {
			if entry.EpochMillis > existing.LastOccurrence {
				existing.LastOccurrence = entry.EpochMillis
			}
		} else {
			scriptsMap[entry.ScriptPath] = &scriptInfo{
				Path:           entry.ScriptPath,
				LastOccurrence: entry.EpochMillis,
			}
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("Error reading log file %s: %w", logPath, err)
	}
	return nil
}

// sortScriptsByLastOccurrence sorts scripts by last occurrence (most recent first)
func sortScriptsByLastOccurrence(scripts []scriptInfo) {
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].LastOccurrence > scripts[j].LastOccurrence
	})
}

// checkScriptWith runs rad checks and returns per-severity counts.
// The bool is false when the file can't be loaded or parsed at all.
func checkScriptWith(scriptPath string, reusable check.RadChecker) (diagCounts, bool) {
	result := com.LoadFile(scriptPath)
	if result.Error != nil {
		return diagCounts{}, false
	}

	chk := reusable
	if chk == nil {
		var err error
		chk, err = check.NewChecker()
		if err != nil {
			return diagCounts{}, false
		}
	}

	chk.UpdateSrc(NormalizeLineEndings(result.Content))
	checkResult, err := chk.CheckDefault()
	if err != nil {
		return diagCounts{}, false
	}

	var counts diagCounts
	for _, diag := range checkResult.Diagnostics {
		switch diag.Severity {
		case check.Error:
			counts.Errors++
		case check.Warning:
			counts.Warnings++
		case check.Info:
			counts.Infos++
		case check.Hint:
			counts.Hints++
		}
	}
	return counts, true
}

// formatCount returns a right-aligned, colored count cell for the table.
// Non-zero values use the given highlight color; zeros are dimmed.
func formatCount(n, width int, highlight, dim *color.Color) string {
	s := fmt.Sprintf("%*d", width, n)
	if n > 0 {
		return "  " + highlight.Sprint(s)
	}
	return "  " + dim.Sprint(s)
}

// ParseDuration parses a human-readable duration string into time.Duration
// Supported formats:
//   - Simple: "30d", "2w", "24h", "1y"
//   - Combinations: "2w3d", "1y2w", "3d12h30m"
//   - Special: "all" returns 0 (no time filtering)
//
// Units:
//   - y: years (365 days, ignores leap years)
//   - w: weeks (7 days)
//   - d: days (24 hours)
//   - h: hours
//   - m: minutes
//   - s: seconds
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)

	// Special case: "all" means no time filtering
	if strings.ToLower(s) == "all" {
		return 0, nil
	}

	// Regex to match all number-unit pairs
	re := regexp.MustCompile(`(\d+)([ydhms]|w)`)
	matches := re.FindAllStringSubmatch(s, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format: %q (expected formats like '30d', '2w3d', or 'all')", s)
	}

	var total time.Duration

	for _, match := range matches {
		valueStr := match[1]
		unit := match[2]

		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number in duration: %q", valueStr)
		}

		var duration time.Duration
		switch unit {
		case "y":
			// Year = 365 days (ignore leap years per spec)
			duration = time.Duration(value) * 365 * 24 * time.Hour
		case "w":
			// Week = 7 days
			duration = time.Duration(value) * 7 * 24 * time.Hour
		case "d":
			// Day = 24 hours
			duration = time.Duration(value) * 24 * time.Hour
		case "h":
			duration = time.Duration(value) * time.Hour
		case "m":
			duration = time.Duration(value) * time.Minute
		case "s":
			duration = time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("unknown unit: %q", unit)
		}

		total += duration
	}

	// Validate that we matched the entire string (no invalid characters)
	reconstructed := ""
	for _, match := range matches {
		reconstructed += match[0]
	}
	if reconstructed != s {
		return 0, fmt.Errorf("invalid duration format: %q contains invalid characters", s)
	}

	return total, nil
}
