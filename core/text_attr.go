package core

import (
	"sort"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/amterp/color"
	tblwriter "github.com/amterp/go-tbl"
)

type RadTextAttr int

// when adding attrs, add 1) here, 2) attrEnumToStrings, 3) ToTblColor, and 4) ToFatihAttr
const (
	PLAIN RadTextAttr = iota
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

	BOLD
	ITALIC
	UNDERLINE
)

var ATTR_STRINGS = make([]string, 0)

var attrEnumToStrings = map[RadTextAttr]string{
	PLAIN:     "plain",
	BLACK:     "black",
	RED:       "red",
	GREEN:     "green",
	YELLOW:    "yellow",
	BLUE:      "blue",
	MAGENTA:   "magenta",
	CYAN:      "cyan",
	WHITE:     "white",
	ORANGE:    "orange",
	PINK:      "pink",
	BOLD:      "bold",
	ITALIC:    "italic",
	UNDERLINE: "underline",
}

var stringsToAttrEnum = make(map[string]RadTextAttr)

func init() {
	for attr, str := range attrEnumToStrings {
		stringsToAttrEnum[str] = attr
		ATTR_STRINGS = append(ATTR_STRINGS, str)
		sort.Strings(ATTR_STRINGS)
	}
}

func (a RadTextAttr) String() string {
	if s, ok := attrEnumToStrings[a]; ok {
		return s
	}
	return "unknown"
}

func TryColorFromString(str string) (RadTextAttr, bool) {
	clr, ok := stringsToAttrEnum[str]
	if !ok {
		return PLAIN, false
	}
	return clr, true
}

func AttrFromString(i *Interpreter, node *ts.Node, str string) RadTextAttr {
	clr, ok := TryColorFromString(str)
	if !ok {
		i.errorf(node, "Invalid color value %q. Allowed: %s", str, ATTR_STRINGS)
	}
	return clr
}

func (a RadTextAttr) ToTblColor() tblwriter.Color {
	switch a {
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
		RP.RadErrorExit("Bug! No Tbl mapping for " + a.String())
		panic(UNREACHABLE)
	}
}

func (a RadTextAttr) AddAttrTo(clr *color.Color) {
	switch a {
	case PLAIN:
		clr.Add(color.Reset)
	case BLACK:
		clr.Add(color.FgBlack)
	case RED:
		clr.Add(color.FgRed)
	case GREEN:
		clr.Add(color.FgGreen)
	case YELLOW:
		clr.Add(color.FgYellow)
	case BLUE:
		clr.Add(color.FgBlue)
	case MAGENTA:
		clr.Add(color.FgMagenta)
	case CYAN:
		clr.Add(color.FgCyan)
	case WHITE:
		clr.Add(color.FgWhite)
	case ORANGE:
		clr.AddRGB(255, 128, 0)
	case PINK:
		clr.AddRGB(255, 172, 187)
	case BOLD:
		clr.Add(color.Bold)
	case ITALIC:
		clr.Add(color.Italic)
	case UNDERLINE:
		clr.Add(color.Underline)
	default:
		RP.RadErrorExit("Bug! No fatih mapping for " + a.String())
		panic(UNREACHABLE)
	}
}

func (a RadTextAttr) Colorize(str string) string {
	clr := color.New()
	a.AddAttrTo(clr)
	return clr.Sprint(str)
}
