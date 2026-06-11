package com

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not resolve home dir: %v", err)
	}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bare tilde", "~", home},
		{"tilde with subpath", "~/foo/bar.txt", filepath.Join(home, "foo/bar.txt")},
		{"no tilde absolute", "/etc/hosts", "/etc/hosts"},
		{"no tilde relative", "foo/bar.txt", "foo/bar.txt"},
		{"tilde mid-string untouched", "/foo/~/bar", "/foo/~/bar"},
		{"tilde literal name untouched", "~backup", "~backup"},
		{"tilde user untouched", "~bob/config.txt", "~bob/config.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandTilde(tt.in); got != tt.want {
				t.Errorf("ExpandTilde(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestToAbsolutePathExpandsTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not resolve home dir: %v", err)
	}

	want := filepath.Join(home, "foo/bar.txt")
	if got := ToAbsolutePath("~/foo/bar.txt"); got != want {
		t.Errorf("ToAbsolutePath(\"~/foo/bar.txt\") = %q, want %q", got, want)
	}
}

func TestWriteFileWithDirCreatesMissingDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "out.json")

	if err := CreateFilePathAndWriteString(path, "hello"); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}
}

func TestWriteFileWithDirReplacesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := CreateFilePathAndWriteString(path, "first version, longer content"); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if err := CreateFilePathAndWriteString(path, "second"); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if string(got) != "second" {
		t.Errorf("content = %q, want %q (old content must be fully replaced)", got, "second")
	}
}

func TestWriteFileWithDirLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := CreateFilePathAndWriteString(path, "data"); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "out.json" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("dir should contain only out.json, got %v", names)
	}
}

func TestWriteFileWithDirPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits don't map to Windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := CreateFilePathAndWriteString(path, "data"); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0644 {
		t.Errorf("perm = %o, want 0644", perm)
	}
}
