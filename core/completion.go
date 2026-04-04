package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amterp/ra"
)

const completionDescription = "Generate shell tab-completion scripts."

const completionLongDescription = `Generate shell tab-completion scripts.
Enables tab-completion for the rad CLI and your Rad scripts.

Add to your shell startup file (e.g. ~/.bashrc or ~/.zshrc):
  eval "$(rad completion bash)"                          # rad CLI
  eval "$(rad completion bash ~/bin/myscript)"           # one script
  eval "$(rad completion bash ~/.rad/bin/* ~/scripts/*)" # all scripts in dirs
  eval "$(rad completion zsh)"                           # zsh variant

When scripts are specified, only script completions are generated.
Use a separate line without scripts for rad CLI completions.

Non-Rad files in glob expansions are silently skipped.
Re-source your shell config after adding new scripts.`

// handleCompletionCommand handles `rad completion <shell> [scripts...]`.
// Generates shell completion scripts for the rad CLI and optionally for custom scripts.
//
// Note: This runs before the printer (RP) is initialized, so it uses direct
// stderr/stdout writes and os.Exit rather than the RP/RExit abstractions.
// Arg parsing and help output are delegated to Ra for consistent formatting.
func (r *RadRunner) handleCompletionCommand(args []string) error {
	// Use os.Exit directly since the Rad exit handler (RExit) depends on RP,
	// which isn't initialized this early. Restore after parsing.
	ra.SetExitFunc(os.Exit)
	defer ra.SetExitFunc(RExit.Exit)

	cmd := ra.NewCmd("rad completion")
	cmd.SetDescription(completionLongDescription)
	cmd.SetHelpEnabled(true)

	shellPtr, _ := ra.NewString("shell").
		SetEnumConstraint([]string{"bash", "zsh"}).
		SetUsage("Shell to generate completions for.").
		SetPositionalOnly(true).
		Register(cmd)

	scriptsPtr, _ := ra.NewStringSlice("scripts").
		SetOptional(true).
		SetVariadic(true).
		SetPositionalOnly(true).
		SetUsage("Paths to Rad scripts to generate completions for.").
		Register(cmd)

	cmd.ParseOrExit(args) // handles -h/--help, validation errors, etc.

	shell := *shellPtr
	var scriptPaths []string
	if scriptsPtr != nil {
		scriptPaths = *scriptsPtr
	}

	var genFunc func(w *os.File, cmdName, funcPrefix, completionCmd string) error
	switch shell {
	case "bash":
		genFunc = func(w *os.File, cmdName, funcPrefix, completionCmd string) error {
			return ra.GenBashCompletionFull(w, cmdName, funcPrefix, completionCmd)
		}
	case "zsh":
		genFunc = func(w *os.File, cmdName, funcPrefix, completionCmd string) error {
			return ra.GenZshCompletionFull(w, cmdName, funcPrefix, completionCmd)
		}
	}

	if len(scriptPaths) == 0 {
		// No scripts specified: generate completions for the rad CLI itself
		if err := genFunc(os.Stdout, "rad", "", "rad"); err != nil {
			fmt.Fprintf(os.Stderr, "rad completion: failed to generate rad completion: %s\n", err)
			os.Exit(1)
		}
	} else {
		// Scripts specified: generate completions for each script only.
		// Rad CLI completions are expected to come from a separate
		// `eval "$(rad completion bash)"` line.
		for _, scriptPath := range scriptPaths {
			if err := generateScriptCompletion(scriptPath, genFunc); err != nil {
				fmt.Fprintf(os.Stderr, "rad completion: skipping %s: %s\n", scriptPath, err)
			}
		}
	}

	return nil
}

// generateScriptCompletion generates a completion function for a single script.
// Function names are prefixed with "rad_" to namespace them (e.g., _rad_harvest_completions).
func generateScriptCompletion(
	scriptPath string,
	genFunc func(w *os.File, cmdName, funcPrefix, completionCmd string) error,
) error {
	// Resolve to absolute path so the completion function works from any directory
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Verify the file exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("cannot stat: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("is a directory")
	}

	// Silently skip non-Rad files. This is expected when users pass glob
	// patterns (e.g., ~/bin/*) that include a mix of Rad and non-Rad files.
	if !isRadScript(absPath) {
		return nil
	}

	// Derive command name from filename (what the user types)
	cmdName := scriptNameFromPath(absPath)
	if cmdName == "" {
		return fmt.Errorf("cannot derive command name from path")
	}

	// Shell-quote the path to prevent injection and handle spaces/special chars.
	// The generated script will contain: out=$(rad '/path/to/script' __complete ...)
	quotedPath := shellQuote(absPath)

	return genFunc(os.Stdout, cmdName, "rad", "rad "+quotedPath)
}

// shellQuote wraps a string in single quotes for safe embedding in shell scripts.
// Single quotes within the string are escaped using the ' + \' + ' idiom.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// isRadScript checks if a file is a Rad script by reading its shebang line.
// Checks that the interpreter binary name is exactly "rad" (not a substring
// like "gradle" or "radare2").
func isRadScript(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return false
	}
	line := scanner.Text()

	if !strings.HasPrefix(line, "#!") {
		return false
	}

	// Split shebang into tokens and check if "rad" appears as an exact token
	// or as the basename of the interpreter path.
	// Matches: #!/usr/bin/env rad, #!/usr/bin/rad, #!/path/to/rad
	// Does not match: #!/usr/bin/env gradle, #!/usr/bin/radare2
	tokens := strings.Fields(line[2:])
	for _, token := range tokens {
		base := filepath.Base(token)
		if base == "rad" {
			return true
		}
	}
	return false
}

// scriptNameFromPath extracts a command name from a script file path.
// Strips directory and common extensions (.rad).
func scriptNameFromPath(path string) string {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, ".rad")
	return name
}

// registerEmbeddedCommandsForCompletion registers embedded commands as Ra subcommands
// with their full args/commands metadata, so shell completion works at all depths
// (e.g., "rad docs <TAB>" shows docs' subcommands).
// Also registers the "completion" command itself so it appears in completions.
func (r *RadRunner) registerEmbeddedCommandsForCompletion() {
	for _, embCmd := range Cmds {
		scriptData := ExtractMetadata(embCmd.Src)
		raSubCmd := ra.NewCmd(embCmd.Name)
		if scriptData.Description != nil {
			raSubCmd.SetDescription(*scriptData.Description)
		}
		raSubCmd.SetHelpEnabled(true)
		registerArgsForCompletion(raSubCmd, scriptData)
		if _, err := RRootCmd.RegisterCmd(raSubCmd); err != nil {
			fmt.Fprintf(os.Stderr, "rad: warning: failed to register completion for command %q: %s\n", embCmd.Name, err)
		}
	}

	// Register the 'completion' command (handled in Go, not an embedded script)
	completionCmd := ra.NewCmd("completion")
	completionCmd.SetDescription(completionDescription)
	if _, err := RRootCmd.RegisterCmd(completionCmd); err != nil {
		fmt.Fprintf(os.Stderr, "rad: warning: failed to register completion for command %q: %s\n", "completion", err)
	}
}

// registerArgsForCompletion registers a script's args and commands on a Ra Cmd
// for completion purposes. Unlike the normal registration path, this doesn't
// track command invocations or wire up interpreter state.
func registerArgsForCompletion(cmd *ra.Cmd, data *ScriptData) {
	hasCommands := len(data.Commands) > 0

	if !hasCommands {
		// No commands: register script args on root as positional+flag
		for _, arg := range data.Args {
			flag := CreateFlag(arg)
			flag.Register(cmd, AsScriptArg)
		}
		return
	}

	// Commands exist: register args on each subcommand
	for _, scriptCmd := range data.Commands {
		raSubCmd := ra.NewCmd(scriptCmd.ExternalName)
		if scriptCmd.Description != nil {
			raSubCmd.SetDescription(*scriptCmd.Description)
		}
		raSubCmd.SetHelpEnabled(true)

		// Script-level args as flag-only on each subcommand
		for _, arg := range data.Args {
			flag := CreateFlag(arg)
			flag.Register(raSubCmd, AsScriptFlagOnly)
		}

		// Command-specific args as positional+flag
		for _, arg := range scriptCmd.Args {
			flag := CreateFlag(arg)
			flag.Register(raSubCmd, AsCommandArg)
		}

		if _, err := cmd.RegisterCmd(raSubCmd); err != nil {
			fmt.Fprintf(os.Stderr, "rad: warning: failed to register completion for subcommand %q: %s\n", scriptCmd.ExternalName, err)
		}
	}
}
