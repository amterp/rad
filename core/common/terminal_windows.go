//go:build windows

package com

// ttyDeviceName is the Windows equivalent of /dev/tty: a reserved name that
// resolves to the console's input buffer regardless of stdin redirection.
// Opening it fails when the process has no attached console.
const ttyDeviceName = "CONIN$"
