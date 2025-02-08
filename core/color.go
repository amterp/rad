package core

import (
	"sort"

	ts "github.com/tree-sitter/go-tree-sitter"

	tblwriter "github.com/amterp/go-tbl"
	"github.com/fatih/color"
)

type RslColor int

// when adding colors, add 1) here, 2) colorEnumToStrings, 3) ToTblColor, and 4) ToFatihColor
const (
	PLAIN RslColor = iota
	BLACK
	RED
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	WHITE
	ORANGE
	PINK
)

var COLOR_STRINGS = make([]string, 0)

var colorEnumToStrings = map[RslColor]string{
	PLAIN:   "plain",
	BLACK:   "black",
	RED:     "red",
	GREEN:   "green",
	YELLOW:  "yellow",
	BLUE:    "blue",
	MAGENTA: "magenta",
	CYAN:    "cyan",
	WHITE:   "white",
	ORANGE:  "orange",
	PINK:    "pink",
}

var stringsToColorEnum = make(map[string]RslColor)

func init() {
	for color, str := range colorEnumToStrings {
		stringsToColorEnum[str] = color
		COLOR_STRINGS = append(COLOR_STRINGS, str)
		sort.Strings(COLOR_STRINGS)
	}
}

func (c RslColor) String() string {
	if s, ok := colorEnumToStrings[c]; ok {
		return s
	}
	return "unknown"
}

func ColorFromString(i *Interpreter, node *ts.Node, str string) RslColor {
	clr, ok := stringsToColorEnum[str]
	if !ok {
		i.errorf(node, "Invalid color value %q. Allowed: %s", str, COLOR_STRINGS)
	}
	return clr
}

func (c RslColor) ToTblColor() tblwriter.Color {
	switch c {
	case PLAIN:
		return tblwriter.Plain
	case BLACK:
		return tblwriter.Black
	case RED:
		return tblwriter.Red
	case GREEN:
		return tblwriter.Green
	case YELLOW:
		return tblwriter.Yellow
	case BLUE:
		return tblwriter.Blue
	case MAGENTA:
		return tblwriter.Magenta
	case CYAN:
		return tblwriter.Cyan
	case WHITE:
		return tblwriter.White
	case ORANGE:
		return tblwriter.Orange
	case PINK:
		return tblwriter.Pink
	default:
		RP.RadErrorExit("Bug! No Tbl mapping for " + c.String())
		panic(UNREACHABLE)
	}
}

func (c RslColor) ToFatihColor() *color.Color {
	switch c {
	case PLAIN:
		return color.New(color.Reset)
	case BLACK:
		return color.New(color.FgBlack)
	case RED:
		return color.New(color.FgRed)
	case GREEN:
		return color.New(color.FgGreen)
	case YELLOW:
		return color.New(color.FgYellow)
	case BLUE:
		return color.New(color.FgBlue)
	case MAGENTA:
		return color.New(color.FgMagenta)
	case CYAN:
		return color.New(color.FgCyan)
	case WHITE:
		return color.New(color.FgWhite)
	case ORANGE:
		return color.RGB(255, 128, 0)
	case PINK:
		return color.RGB(255, 172, 187)
	default:
		RP.RadErrorExit("Bug! No fatih mapping for " + c.String())
		panic(UNREACHABLE)
	}
}

func (c RslColor) Colorize(str string) string {
	return c.ToFatihColor().Sprint(str)
}
