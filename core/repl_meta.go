package core

import (
	"os"
	"sort"
	"strings"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/rl"
)

// Meta commands are prefixed with ':' - vim and ghci both spell it that way,
// and no Rad statement can begin with one, so the prefix is unambiguous. They
// are only recognized at the start of a fresh turn, never inside a block you
// are part-way through typing.

// isReplMeta reports whether a turn's source is a meta command rather than Rad.
func isReplMeta(src string) bool {
	return strings.HasPrefix(strings.TrimSpace(src), ":")
}

// runReplMeta executes a meta command. It returns false when the command asked
// the session to end.
func (s *ReplSession) runReplMeta(src string) (keepGoing bool) {
	fields := strings.Fields(strings.TrimSpace(src))
	cmd := fields[0]
	args := fields[1:]

	switch cmd {
	case ":exit", ":quit":
		return false
	case ":help":
		s.metaHelp()
	case ":vars":
		s.metaVars()
	case ":docs":
		s.metaDocs(args)
	case ":load":
		s.metaLoad(args)
	case ":clear":
		RP.Print("\033[H\033[2J")
	case ":reset":
		s.resetEnv()
		RP.Printf("Session reset.\n")
	default:
		RP.RadStderrf("Unknown command %q. Try :help.\n", cmd)
	}
	return true
}

func (s *ReplSession) metaHelp() {
	RP.Print(strings.Join(com.Wrap(
		"Type Rad and press Enter to run it. A block construct keeps taking "+
			"lines until you enter a blank one, which is also how you get out "+
			"of a buffer you would rather abandon.",
		DiagnosticProseWidth(),
	), "\n") + "\n\n")
	RP.Print(strings.Join([]string{
		"  :help            this",
		"  :vars            variables defined this session",
		"  :docs <topic>    docs for a function, error code, or page",
		"  :load <file>     run a .rad file into this session",
		"  :clear           clear the screen",
		"  :reset           forget every variable and start over",
		"  :exit, :quit     leave (Ctrl+D does too)",
		"",
		"  Ctrl+C           abandon the line, or interrupt what is running",
		"  Up / Down        recall earlier input",
		"",
	}, "\n"))
	RP.Print(strings.Join(com.Wrap(
		"Note that 'defer' runs at the end of the turn that registered it, "+
			"not when the session ends - a turn is this REPL's whole notion of "+
			"a script.",
		DiagnosticProseWidth(),
	), "\n") + "\n")
}

// metaVars lists what the session holds, skipping the builtins seeded into the
// root environment - you asked what *you* defined.
//
// The test is whether the name still *holds* a builtin, not whether a builtin
// claims it. Rad lets a variable shadow one, and `count = 3` is exactly the
// kind of thing you want to see listed.
func (s *ReplSession) metaVars() {
	names := s.interpreter.env.AllVarNames()
	sort.Strings(names)

	shown := 0
	for _, name := range names {
		val, ok := s.interpreter.env.GetVar(name)
		if !ok {
			continue
		}
		if _, isBuiltIn := FunctionsByName[name]; isBuiltIn && val.Type() == rl.RadFnT {
			continue
		}
		RP.Printf("  %-16s %-8s %s\n", name, TypeAsString(val), ToPrintable(val))
		shown++
	}
	if shown == 0 {
		RP.Printf("No variables defined yet.\n")
	}
}

func (s *ReplSession) metaDocs(args []string) {
	if len(args) == 0 {
		RP.RadStderrf("Usage: :docs <function, error code, or page>\n")
		return
	}
	topic := args[0]
	content, ok := GetDocTopic(topic)
	if !ok {
		RP.RadStderrf("No docs for %q. Try a function name, an error code like RAD20028, or a page slug.\n", topic)
		return
	}
	RP.Printf("%s\n", RenderMarkdownForTerminal(content))
}

// metaLoad runs a file's contents as one turn, which is how a REPL session
// picks up helper functions without pasting them.
func (s *ReplSession) metaLoad(args []string) {
	if len(args) == 0 {
		RP.RadStderrf("Usage: :load <file.rad>\n")
		return
	}
	path := args[0]
	src, err := os.ReadFile(path)
	if err != nil {
		RP.RadStderrf("Cannot read %s: %v\n", path, err)
		return
	}
	s.runSource(NormalizeLineEndings(string(src)))
}

func (s *ReplSession) resetEnv() {
	s.interpreter.env = NewEnv(s.interpreter)
	s.interpreter.InitBuiltIns()
}
