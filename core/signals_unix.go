//go:build unix

package core

import (
	"os"
	"syscall"
)

// unixSignals is the full POSIX-ish set we expose on unix-likes.
// The integer is the conventional POSIX signal number (used for 128+sig
// exit codes); we hardcode rather than reading from syscall constants
// because they are stable and this lets the table stay readable.
var unixSignals = map[string]signalEntry{
	sigNameSigint:   {osSignal: syscall.SIGINT, sigNum: 2},
	sigNameSigterm:  {osSignal: syscall.SIGTERM, sigNum: 15},
	sigNameSighup:   {osSignal: syscall.SIGHUP, sigNum: 1},
	sigNameSigusr1:  {osSignal: syscall.SIGUSR1, sigNum: 10},
	sigNameSigusr2:  {osSignal: syscall.SIGUSR2, sigNum: 12},
	sigNameSigpipe:  {osSignal: syscall.SIGPIPE, sigNum: 13},
	sigNameSigwinch: {osSignal: syscall.SIGWINCH, sigNum: 28},
}

func supportedSignals() map[string]signalEntry {
	return unixSignals
}

// resolveSignalName returns the os.Signal and POSIX signal number for the
// given Rad-level signal name. Returns an error if the name is unknown or
// (on this platform) unsupported.
func resolveSignalName(name string) (os.Signal, int, error) {
	entry, ok := unixSignals[name]
	if !ok {
		// If the name is one we recognize globally but happens not to be
		// in the unix table (shouldn't currently happen but be defensive),
		// return the unsupported-here error. Otherwise return the
		// unknown-altogether error.
		if isKnownSignalName(name) {
			return nil, 0, errUnsupportedSignal(name)
		}
		return nil, 0, errUnknownSignal(name)
	}
	return entry.osSignal, entry.sigNum, nil
}
