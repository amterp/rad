package com

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandTilde resolves a leading "~" (i.e. exactly "~" or a "~/" prefix) to the
// user's home directory. Anything else is returned unchanged: "~user" (another
// user's home) is not supported and is left as a literal path rather than
// silently misexpanded, and a path like "~backup" is treated as a literal name.
// If the home dir can't be resolved, the path is returned untouched so the
// failure surfaces honestly at the os call site.
func ExpandTilde(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[1:]) // drop the "~"
		}
	}
	return path
}

func ToAbsolutePath(path string) string {
	abs, _ := filepath.Abs(ExpandTilde(path)) // todo handle err?
	return abs
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsRegularFile returns true if the path exists and is a regular file
// (not a directory, device, pipe, socket, etc.).
func IsRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func LoadJson(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

func CreateFilePathAndWriteJson(path string, jsonData interface{}) error {
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return err
	}
	return writeFileWithDir(path, jsonBytes)
}

func CreateFilePathAndWriteString(path string, str string) error {
	return writeFileWithDir(path, []byte(str))
}

func DeleteFileIfExists(relativePath string) error {
	if _, err := os.Stat(relativePath); err == nil {
		err := os.Remove(relativePath)
		if err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check file: %w", err)
	}
	return nil
}

func LoadFile(path string) LoadFileResult {
	data, err := os.ReadFile(path)
	if err != nil {
		return LoadFileResult{Error: fmt.Errorf("failed to read file: %w", err)}
	}
	sizeBytes := int64(len(data))
	return LoadFileResult{Content: string(data), SizeBytes: sizeBytes}
}

type LoadFileResult struct {
	Content   string
	SizeBytes int64
	Error     error
}

// writeFileWithDir writes data to path atomically: the bytes land in a temp
// file in the same directory (same filesystem, so the rename can't cross
// devices) which is then renamed over the target. A crash or concurrent
// writer can no longer leave a partially-written file behind. We skip
// fsync - these are state/stash files, not a database; the rename gives us
// all-or-nothing content, which is the guarantee callers need.
func writeFileWithDir(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Leading dot keeps in-flight temp files out of glob/listing-based
	// consumers of the directory (e.g. stash file readers).
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func(err error) error {
		_ = os.Remove(tmpPath)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return cleanup(err)
	}
	if err := tmp.Close(); err != nil {
		return cleanup(err)
	}
	// CreateTemp creates 0600; restore the 0644 these files have always had.
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return cleanup(err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return cleanup(err)
	}
	return nil
}
