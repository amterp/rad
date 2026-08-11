//go:build windows

package com

import "os"

// Windows has no single read/write equivalent of /dev/tty. It exposes the
// console as two reserved names: CONIN$ is the input buffer, CONOUT$ the active
// screen buffer. Writing a prompt to CONIN$ would queue it as input rather than
// draw it, leaving the user typing blind, so the two are opened separately.
// Opening either fails when the process has no attached console - the case rad
// wants to detect.
const (
	ttyInDeviceName  = "CONIN$"
	ttyOutDeviceName = "CONOUT$"
)

// Both are opened read/write, which is what the Win32 docs call for on these
// devices even when only one direction is used.
func openControllingTerminal() (*ControllingTerminal, error) {
	in, err := os.OpenFile(ttyInDeviceName, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	out, err := os.OpenFile(ttyOutDeviceName, os.O_RDWR, 0)
	if err != nil {
		in.Close()
		return nil, err
	}
	return &ControllingTerminal{
		In:  in,
		Out: out,
		close: func() {
			in.Close()
			out.Close()
		},
	}, nil
}
