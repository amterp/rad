//go:build linux

package core

import (
	"os"
	"syscall"
)

// getAccessTimeMillis extracts the access time from a FileInfo on Linux.
func getAccessTimeMillis(fi os.FileInfo) (int64, bool) {
	if sys, ok := fi.Sys().(*syscall.Stat_t); ok {
		// Linux uses Atim
		millis := sys.Atim.Sec*1000 + sys.Atim.Nsec/1_000_000
		return millis, true
	}
	return 0, false
}
