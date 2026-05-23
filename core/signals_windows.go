//go:build windows

package core

import (
	"os"
	"syscall"
)

// windowsSignals is the subset of signals that map cleanly to Windows.
// Windows does not have POSIX signals; only SIGINT (via Ctrl+C / os.Interrupt)
// and SIGTERM (process termination) are meaningful. Everything else returns
// an unsupported-on-this-platform error at registration time so scripts fail
// loudly rather than silently no-op.
var windowsSignals = map[string]signalEntry{
	sigNameSigint:  {osSignal: os.Interrupt, sigNum: 2},
	sigNameSigterm: {osSignal: syscall.SIGTERM, sigNum: 15},
}

func supportedSignals() map[string]signalEntry {
	return windowsSignals
}

func resolveSignalName(name string) (os.Signal, int, error) {
	entry, ok := windowsSignals[name]
	if !ok {
		if isKnownSignalName(name) {
			return nil, 0, errUnsupportedSignal(name)
		}
		return nil, 0, errUnknownSignal(name)
	}
	return entry.osSignal, entry.sigNum, nil
}
