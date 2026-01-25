package com

import (
	"path/filepath"
	"runtime"
	"strings"
)

// Platform abstraction layer for cross-platform compatibility.
// All platform-specific behavior should be centralized here.
//
// PATH NORMALIZATION STRATEGY:
// Rad normalizes all paths to use forward slashes, regardless of the host OS.
// This ensures scripts are portable across platforms - forward slashes work
// on all operating systems, including Windows. Any function that returns a
// path to user code should call NormalizePath() before returning.
//
// LINE ENDING STRATEGY:
// Rad normalizes line endings to Unix-style (\n) when reading text files.
// This ensures consistent string processing across platforms. Binary file
// reads are NOT normalized. Any function that reads text content should
// call NormalizeLineEndings() on the result.

// NormalizePath converts OS-specific path separators to forward slashes.
// This ensures Rad scripts are portable across platforms - forward slashes
// work on all operating systems, including Windows.
//
// Call this on any path before returning it to user code.
func NormalizePath(path string) string {
	return filepath.ToSlash(path)
}

// NormalizeLineEndings converts Windows-style line endings (\r\n) to Unix-style (\n).
// This ensures consistent text handling across platforms.
//
// Call this on text content read from files before returning to user code.
func NormalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// ToAbsoluteNormalizedPath expands ~ to home directory, resolves to absolute path,
// and normalizes to forward slashes. This is the standard way to process user-provided
// paths before returning them.
func ToAbsoluteNormalizedPath(path string) string {
	return NormalizePath(ToAbsolutePath(path))
}
