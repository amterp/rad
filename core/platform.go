package core

import (
	com "github.com/amterp/rad/core/common"
)

// Platform abstraction layer for cross-platform compatibility.
// The canonical implementation lives in core/common/platform.go.
// These are re-exported here for convenience within the core package.
//
// See common/platform.go for the full documentation on platform strategy.

// NormalizePath converts OS-specific path separators to forward slashes.
// This ensures Rad scripts are portable across platforms.
func NormalizePath(path string) string {
	return com.NormalizePath(path)
}

// NormalizeLineEndings converts Windows-style line endings (\r\n) to Unix-style (\n).
// This ensures consistent text handling across platforms.
func NormalizeLineEndings(s string) string {
	return com.NormalizeLineEndings(s)
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return com.IsWindows()
}
