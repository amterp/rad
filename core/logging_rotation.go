package core

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

// MaybeRotate checks if the current log file exceeds the size threshold
// and rotates it if necessary. Non-fatal: warns on errors but doesn't crash.
func MaybeRotate() {
	cfg := *RConfig.InvocationLogging
	logPath := GetInvocationLogPath()

	// Pre-lock size check (fast path - avoid lock if not needed)
	info, err := os.Stat(logPath)
	if err != nil {
		// File doesn't exist or other error - no rotation needed
		return
	}

	maxBytes := int64(cfg.MaxSizeMB) * 1024 * 1024
	if info.Size() < maxBytes {
		// Below threshold, no rotation needed
		return
	}

	// Size exceeds threshold, attempt platform-specific rotation
	tryRotate(cfg, logPath, maxBytes)
}

// cleanupOldLogs removes old rolled logs, keeping only the newest N files
func cleanupOldLogs(config InvocationLoggingConfig) {
	logsDir := GetInvocationLogsDir()

	// Find all rolled log files (invocations-*.jsonl)
	pattern := filepath.Join(logsDir, "invocations-*.jsonl")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		RP.RadStderrf("Warning! Failed to find rolled log files for cleanup: %v\n", err)
		return
	}

	if len(matches) <= config.KeepRolledLogs {
		// Nothing to clean up
		return
	}

	// Sort by filename (lexicographic = chronological due to timestamp format)
	sort.Strings(matches)

	// Delete oldest files (keep newest KeepRolledLogs files)
	toDelete := matches[:len(matches)-config.KeepRolledLogs]
	for _, path := range toDelete {
		if err := os.Remove(path); err != nil {
			RP.RadStderrf("Warning! Failed to delete old log file %s: %v\n", path, err)
			// Continue cleanup even if one file fails
		}
	}
}

// generateRolledLogName generates a rolled log filename with UTC timestamp
// Format: invocations-20060102T150405.000.jsonl
func generateRolledLogName() string {
	timestamp := time.Now().UTC().Format("20060102T150405.000")
	return "invocations-" + timestamp + ".jsonl"
}

func getLockFilePath() string {
	return filepath.Join(GetInvocationLogsDir(), ".lock")
}

// performRotation executes the actual rotation: rename current log to rolled name
// Should only be called when holding the rotation lock
func performRotation(logPath string) error {
	logsDir := GetInvocationLogsDir()
	rolledName := generateRolledLogName()
	rolledPath := filepath.Join(logsDir, rolledName)

	// Atomic rename (on POSIX systems)
	if err := os.Rename(logPath, rolledPath); err != nil {
		return err
	}

	return nil
}
