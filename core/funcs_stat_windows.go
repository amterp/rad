//go:build windows

package core

import "os"

// getAccessTimeMillis is a stub for Windows that returns false.
// Access time is not easily available through the standard Go interface on Windows.
func getAccessTimeMillis(fi os.FileInfo) (int64, bool) {
	return 0, false
}
