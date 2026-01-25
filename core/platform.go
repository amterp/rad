package core

import (
	"path/filepath"
	"runtime"
	"strings"
)

// Platform abstraction layer for cross-platform compatibility.
// All platform-specific behavior should be centralized here.

// NormalizePath converts OS-specific path separators to forward slashes.
// This ensures Rad scripts are portable across platforms - forward slashes
// work on all operating systems, including Windows.
func NormalizePath(path string) string {
	return filepath.ToSlash(path)
}

// NormalizeLineEndings converts Windows-style line endings (\r\n) to Unix-style (\n).
// This ensures consistent text handling across platforms.
func NormalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}
