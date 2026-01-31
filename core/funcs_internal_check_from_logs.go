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
	ts "github.com/tree-sitter/go-tree-sitter"
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

		totalChecked, passed, failed, skipped := 0, 0, 0, 0

		for _, script := range scripts {
			if !com.FileExists(script.Path) {
				skipped++
				if verbose {
					dimmed := color.New(color.Faint)
					RP.Printf("  %s\n", dimmed.Sprintf("- %s (file not found)", script.Path))
				}
				continue
			}

			ok := checkScriptWith(script.Path, chk)
			totalChecked++
			if ok {
				green := color.New(color.FgGreen)
				RP.Printf("  %s %s\n", green.Sprint("✓"), script.Path)
				passed++
			} else {
				red := color.New(color.FgRed)
				RP.Printf("  %s %s\n", red.Sprint("✗"), script.Path)
				failed++
			}
		}

		// Print separator
		RP.Printf("\n────────────────────────────────────────\n")

		// Build summary with colored numbers
		bold := color.New(color.Bold)
		green := color.New(color.FgGreen)
		red := color.New(color.FgRed)
		faint := color.New(color.Faint)

		summary := fmt.Sprintf("Checked %d scripts: %s passed, %s failed",
			totalChecked,
			green.Sprintf("%d", passed),
			red.Sprintf("%d", failed))

		if skipped > 0 {
			summary += fmt.Sprintf(" (%s skipped)", faint.Sprintf("%d", skipped))
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
func parseInvocationLogs(i *Interpreter, node *ts.Node, durationMillis int64) []scriptInfo {
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

// checkScriptWith runs rad checks; reuses a checker if provided.
func checkScriptWith(scriptPath string, reusable check.RadChecker) bool {
	result := com.LoadFile(scriptPath)
	if result.Error != nil {
		return false
	}

	chk := reusable
	if chk == nil {
		var err error
		chk, err = check.NewChecker()
		if err != nil {
			return false
		}
	}

	chk.UpdateSrc(NormalizeLineEndings(result.Content))
	checkResult, err := chk.CheckDefault()
	if err != nil {
		return false
	}

	for _, diag := range checkResult.Diagnostics {
		if diag.Severity == check.Error {
			return false
		}
	}
	return true
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
