package core

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fakeLookPath returns a lookPath function that succeeds for any name in
// the available set and fails otherwise.
func fakeLookPath(available ...string) func(string) (string, error) {
	set := make(map[string]bool, len(available))
	for _, a := range available {
		set[a] = true
	}
	return func(name string) (string, error) {
		if set[name] {
			return name, nil
		}
		return "", exec.ErrNotFound
	}
}

// fakeGetEnv returns a getEnv function backed by the given map.
func fakeGetEnv(env map[string]string) func(string) string {
	return func(key string) string { return env[key] }
}

func TestResolveShell_SHELL_set_wins_on_unix(t *testing.T) {
	// On Unix, SHELL is trusted as-is - no LookPath validation - because the
	// user's explicit choice takes precedence and a missing path failing
	// loudly is the right behavior for genuine misconfiguration.
	path, flag, err := resolveShell(
		fakeGetEnv(map[string]string{"SHELL": "/usr/local/bin/zsh"}),
		fakeLookPath(), // empty - should not be consulted
		false,
	)
	assert.NoError(t, err)
	assert.Equal(t, "/usr/local/bin/zsh", path)
	assert.Equal(t, "-c", flag)
}

func TestResolveShell_SHELL_set_wins_on_windows_when_resolvable(t *testing.T) {
	path, flag, err := resolveShell(
		fakeGetEnv(map[string]string{"SHELL": "pwsh.exe"}),
		fakeLookPath("pwsh.exe"),
		true,
	)
	assert.NoError(t, err)
	assert.Equal(t, "pwsh.exe", path)
	assert.Equal(t, "-c", flag)
}

func TestResolveShell_Windows_GitBash_SHELL_falls_through(t *testing.T) {
	// Git Bash / MSYS2 / Cygwin auto-set SHELL to a Unix-style path that
	// native Win32 exec can't see. We should fall through to the Windows
	// candidate chain rather than handing back an unusable path.
	path, flag, err := resolveShell(
		fakeGetEnv(map[string]string{"SHELL": "/usr/bin/bash"}),
		fakeLookPath("powershell.exe", "cmd.exe"), // /usr/bin/bash NOT in PATH
		true,
	)
	assert.NoError(t, err)
	assert.Equal(t, "powershell.exe", path)
	assert.Equal(t, "-c", flag)
}

func TestResolveShell_whitespace_only_SHELL_treated_as_unset(t *testing.T) {
	for _, isWindows := range []bool{false, true} {
		fallback := "/bin/sh"
		if isWindows {
			fallback = "cmd.exe"
		}
		path, _, err := resolveShell(
			fakeGetEnv(map[string]string{"SHELL": "   "}),
			fakeLookPath(fallback),
			isWindows,
		)
		assert.NoError(t, err)
		assert.Equal(t, fallback, path)
	}
}

func TestResolveShell_SHELL_set_to_cmd_uses_slash_c(t *testing.T) {
	cmdPath := `C:\Windows\System32\cmd.exe`
	path, flag, err := resolveShell(
		fakeGetEnv(map[string]string{"SHELL": cmdPath}),
		fakeLookPath(cmdPath),
		true,
	)
	assert.NoError(t, err)
	assert.Equal(t, cmdPath, path)
	assert.Equal(t, "/c", flag)
}

func TestResolveShell_Windows_prefers_pwsh(t *testing.T) {
	path, flag, err := resolveShell(
		fakeGetEnv(nil),
		fakeLookPath("pwsh.exe", "powershell.exe", "cmd.exe"),
		true,
	)
	assert.NoError(t, err)
	assert.Equal(t, "pwsh.exe", path)
	assert.Equal(t, "-c", flag)
}

func TestResolveShell_Windows_falls_back_to_powershell(t *testing.T) {
	path, flag, err := resolveShell(
		fakeGetEnv(nil),
		fakeLookPath("powershell.exe", "cmd.exe"),
		true,
	)
	assert.NoError(t, err)
	assert.Equal(t, "powershell.exe", path)
	assert.Equal(t, "-c", flag)
}

func TestResolveShell_Windows_falls_back_to_cmd(t *testing.T) {
	path, flag, err := resolveShell(
		fakeGetEnv(nil),
		fakeLookPath("cmd.exe"),
		true,
	)
	assert.NoError(t, err)
	assert.Equal(t, "cmd.exe", path)
	assert.Equal(t, "/c", flag)
}

func TestResolveShell_Unix_uses_bin_sh(t *testing.T) {
	path, flag, err := resolveShell(
		fakeGetEnv(nil),
		fakeLookPath("/bin/sh"),
		false,
	)
	assert.NoError(t, err)
	assert.Equal(t, "/bin/sh", path)
	assert.Equal(t, "-c", flag)
}

func TestResolveShell_no_shell_available_returns_error(t *testing.T) {
	for _, isWindows := range []bool{false, true} {
		_, _, err := resolveShell(
			fakeGetEnv(nil),
			fakeLookPath(), // nothing available
			isWindows,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SHELL environment variable")
	}
}

func TestResolveShell_lookPath_other_error_treated_as_not_found(t *testing.T) {
	// Some lookPath implementations may return errors other than ErrNotFound
	// (e.g., permission errors). We should still skip the candidate and
	// continue down the chain.
	calls := 0
	lookPath := func(name string) (string, error) {
		calls++
		if name == "pwsh.exe" {
			return "", errors.New("permission denied")
		}
		if name == "powershell.exe" {
			return name, nil
		}
		return "", exec.ErrNotFound
	}
	path, flag, err := resolveShell(fakeGetEnv(nil), lookPath, true)
	assert.NoError(t, err)
	assert.Equal(t, "powershell.exe", path)
	assert.Equal(t, "-c", flag)
	assert.Equal(t, 2, calls)
}

func TestShellExecFlag(t *testing.T) {
	tests := []struct {
		shellPath string
		want      string
	}{
		// POSIX shells -> -c
		{"/bin/sh", "-c"},
		{"/bin/bash", "-c"},
		{"/usr/bin/zsh", "-c"},
		{"/usr/local/bin/fish", "-c"},
		// PowerShell -> -c (accepted as short form of -Command)
		{"pwsh", "-c"},
		{"pwsh.exe", "-c"},
		{"powershell.exe", "-c"},
		{`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, "-c"},
		// cmd.exe -> /c, case- and extension-insensitive
		{"cmd", "/c"},
		{"cmd.exe", "/c"},
		{"CMD.EXE", "/c"},
		{`C:\Windows\System32\cmd.exe`, "/c"},
	}
	for _, tt := range tests {
		t.Run(tt.shellPath, func(t *testing.T) {
			assert.Equal(t, tt.want, shellExecFlag(tt.shellPath))
		})
	}
}
