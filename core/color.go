package core

import (
	tblwriter "github.com/amterp/go-tbl"
	"sort"
)

type RslColor int

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

func ColorFromString(s string) (RslColor, bool) {
	color, ok := stringsToColorEnum[s]
	return color, ok
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
		RP.RadErrorExit("Bug! To Tbl mapping for " + c.String())
		panic(UNREACHABLE)
	}
}
