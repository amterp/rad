package com

import (
	"os"
	"path/filepath"
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
