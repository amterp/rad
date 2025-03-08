package com

import "github.com/fatih/color"

var (
	plain     = color.New(color.Reset)
	green     = color.New(color.FgGreen)
	greenBold = color.New(color.FgGreen, color.Bold)
	yellow    = color.New(color.FgYellow)
	cyan      = color.New(color.FgCyan)
	bold      = color.New(color.Bold)

	PlainF     = plain.FprintfFunc()
	GreenF     = green.FprintfFunc()
	GreenBoldF = greenBold.FprintfFunc()
	YellowF    = yellow.FprintfFunc()
	CyanF      = cyan.FprintfFunc()
	BoldF      = bold.FprintfFunc()

	PlainS     = plain.SprintfFunc()
	GreenS     = green.SprintfFunc()
	GreenBoldS = greenBold.SprintfFunc()
	YellowS    = yellow.SprintfFunc()
	CyanS      = cyan.SprintfFunc()
	BoldS      = bold.SprintfFunc()
)
