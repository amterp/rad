package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	com "github.com/amterp/rad/core/common"
)

func RegisterInvocationLogging() {
	if !RConfig.InvocationLogging.Enabled {
		return
	}
	handler := func() {
		endEpochMillis := RClock.Now().UnixMilli()
		durationMillis := endEpochMillis - StartEpochMillis

		// Capture args if configured
		var args []string
		if RConfig.InvocationLogging.IncludeArgs && len(os.Args) > 2 {
			args = os.Args[2:] // Skip binary name and script path
		}

		entry := InvocationLogEntry{
			EpochMillis:    StartEpochMillis,
			ScriptPath:     com.ToAbsolutePath(ScriptPath),
			Args:           args,
			RadVersion:     Version,
			DurationMillis: durationMillis,
		}

		LogInvocation(entry)
		MaybeRotate()
	}
	RExit.AddPreExitCallback(handler)
}

// InvocationLogEntry represents a single script execution log entry
type InvocationLogEntry struct {
	EpochMillis    int64    `json:"epoch_millis"`
	ScriptPath     string   `json:"script_path"`
	Args           []string `json:"args,omitempty"` // omitempty when include_args=false
	RadVersion     string   `json:"rad_version"`
	DurationMillis int64    `json:"duration_millis"`
}

// GetInvocationLogPath returns the path to the current invocation log file
func GetInvocationLogPath() string {
	return filepath.Join(GetInvocationLogsDir(), "invocations.jsonl")
}

// GetInvocationLogsDir returns the directory containing invocation logs
func GetInvocationLogsDir() string {
	return filepath.Join(RadHomeInst.HomeDir, "logs")
}

// LogInvocation appends an invocation log entry to the JSONL log file
// Creates directory and file if they don't exist
func LogInvocation(entry InvocationLogEntry) {
	logPath := GetInvocationLogPath()
	logsDir := GetInvocationLogsDir()

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		RP.RadStderrf("Warning! Failed to create invocation logs directory: %v\n", err)
		return
	}

	// Serialize entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		RP.RadStderrf("Warning! Failed to serialize invocation log entry: %v\n", err)
		return
	}

	// Append newline for JSONL format
	data = append(data, '\n')

	// Open file with O_APPEND for atomic appends
	// O_CREATE: create if doesn't exist
	// O_WRONLY: write-only
	// O_APPEND: atomic appends on POSIX systems
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		RP.RadStderrf("Warning! Failed to open invocation log file: %v\n", err)
		return
	}
	defer file.Close()

	// Write entry (single write call for line integrity)
	if _, err := file.Write(data); err != nil {
		RP.RadStderrf("Warning! Failed to write to invocation log: %v\n", err)
		return
	}
}
