//go:build !windows

package com

import "os"

// ttyDeviceName is the path to the process's controlling terminal. Opening it
// succeeds even when stdin/stdout are redirected, and fails with ENXIO when the
// process has no controlling terminal at all (CI, cron, an agent's tool call).
const ttyDeviceName = "/dev/tty"

// One handle drives both directions here, since /dev/tty is a single read/write
// device.
func openControllingTerminal() (*ControllingTerminal, error) {
	f, err := os.OpenFile(ttyDeviceName, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	return &ControllingTerminal{In: f, Out: f, close: func() { f.Close() }}, nil
}
