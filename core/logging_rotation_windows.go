//go:build windows

package core

// tryRotate is a stub for Windows that warns about unsupported rotation
func tryRotate(config InvocationLoggingConfig, logPath string, maxBytes int64) {
	RP.RadStderrf("Warning! Log rotation not yet supported on Windows. Log file at %s has exceeded size limit. Please manually rotate.\n", logPath)
}
