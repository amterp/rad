//go:build !darwin && !linux && !windows

package core

import "os"

// getAccessTimeMillis is a fallback for unsupported Unix variants.
func getAccessTimeMillis(fi os.FileInfo) (int64, bool) {
	return 0, false
}
