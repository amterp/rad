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
// Rad normalizes line endings to Unix-style (\n) only for script source code
// (reading .rad files and stdin scripts). User data read via read_file(),
// read_stdin(), and load_stash_file() is NOT normalized - this preserves
// data fidelity so that a read-then-write round-trip doesn't corrupt files.
// Users who want line splitting that handles all endings can use split_lines().

// NormalizePath converts OS-specific path separators to forward slashes.
// This ensures Rad scripts are portable across platforms - forward slashes
// work on all operating systems, including Windows.
//
// Call this on any path before returning it to user code.
func NormalizePath(path string) string {
	return filepath.ToSlash(path)
}

// NormalizeLineEndings converts Windows-style line endings (\r\n) to Unix-style (\n).
//
// Used for script source code only - not for user data.
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
