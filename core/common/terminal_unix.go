//go:build !windows

package com

// ttyDeviceName is the path to the process's controlling terminal. Opening it
// succeeds even when stdin/stdout are redirected, and fails with ENXIO when the
// process has no controlling terminal at all (CI, cron, an agent's tool call).
const ttyDeviceName = "/dev/tty"
