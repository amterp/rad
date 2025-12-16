//go:build darwin

package core

import (
	"os"
	"syscall"
)

// getAccessTimeMillis extracts the access time from a FileInfo on macOS.
func getAccessTimeMillis(fi os.FileInfo) (int64, bool) {
	if sys, ok := fi.Sys().(*syscall.Stat_t); ok {
		// macOS uses Atimespec
		millis := sys.Atimespec.Sec*1000 + sys.Atimespec.Nsec/1_000_000
		return millis, true
	}
	return 0, false
}
