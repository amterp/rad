//go:build unix

package core

import (
	"os"
	"syscall"
)

// tryRotate attempts to rotate the log file using POSIX file locking
// Uses flock to ensure only one process performs rotation at a time
func tryRotate(config InvocationLoggingConfig, logPath string, maxBytes int64) {
	lockPath := getLockFilePath()

	// Open/create lock file
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		RP.RadStderrf("Warning! Failed to open rotation lock file: %v\n", err)
		return
	}
	defer lockFile.Close()

	// Try to acquire exclusive lock (non-blocking)
	// LOCK_EX: exclusive lock
	// LOCK_NB: non-blocking (fail immediately if already locked)
	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		// Another process is rotating, skip
		return
	}
	// Lock will be automatically released when lockFile is closed (defer above)

	// Re-check size under lock (another process may have already rotated)
	info, err := os.Stat(logPath)
	if err != nil {
		// File doesn't exist or other error - no rotation needed
		return
	}

	if info.Size() < maxBytes {
		// Already rotated by another process
		return
	}

	// Perform rotation
	if err := performRotation(logPath); err != nil {
		RP.RadStderrf("Warning! Failed to rotate log file: %v\n", err)
		return
	}

	// Cleanup old logs
	cleanupOldLogs(config)
}
