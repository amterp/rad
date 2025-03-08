package com

import (
	"github.com/mattn/go-isatty"
	"os"
	"strings"
)

var (
	IsTty          = checkTty()
	TerminalIsUtf8 = checkTerminalUtf8()
)

func checkTty() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

func checkTerminalUtf8() bool {
	lang := os.Getenv("LANG")
	ctype := os.Getenv("LC_CTYPE")
	// Check for UTF-8 in LANG or LC_CTYPE environment variables
	return strings.Contains(lang, "UTF-8") || strings.Contains(ctype, "UTF-8")
}
