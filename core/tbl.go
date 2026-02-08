package core

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/rl"

	tblwriter "github.com/amterp/go-tbl"
	"github.com/samber/lo"
	"golang.org/x/term"
)

const (
	tblPadding = "  "
)

type GeneralSort struct {
	Span rl.Span
	Dir  SortDir
}
type ColumnSort struct {
	ColName string
	Span    rl.Span
	Dir     SortDir
}

type TblWriter struct {
	writer      io.Writer
	tbl         *tblwriter.Table
	headers     []string
	rows        [][]RadString
	colToColors map[string][]radColorMod
	numColumns  int
}

func NewTblWriter() *TblWriter {
	stdWriter := RP.GetStdWriter()
	return &TblWriter{
		writer:     stdWriter,
		tbl:        tblwriter.NewWriter(stdWriter),
		numColumns: 0,
	}
}

func (w *TblWriter) SetHeader(headers []string) {
	w.headers = headers
	w.numColumns = len(headers)
}

func (w *TblWriter) Append(row []RadString) {
	w.rows = append(w.rows, row)
	if w.numColumns < len(row) {
		w.numColumns = len(row)
	}
}

func (w *TblWriter) SetColumnColoring(colToMods map[string]*radFieldMods) {
	var colorMods = make(map[string][]radColorMod)
	for colName, mods := range colToMods {
		if len(mods.colors) > 0 {
			colorMods[colName] = mods.colors
		}
	}
	w.colToColors = colorMods
}

func (w *TblWriter) Render() {
	// todo this should almost definitely be mocked out for tests
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		RP.RadDebugf(fmt.Sprintf("Error getting terminal width, setting to 9999: %v\n", err))
		termWidth = 9999
	}

	// resolve how many chars each column needs to fully display its contents
	colWidths := make([]int, w.numColumns)
	for i := range w.headers {
		colWidths[i] = len(w.headers[i])
	}
	for _, row := range w.rows {
		for i, cell := range row {
			lines := strings.Split(cell.Plain(), "\n")
			for _, line := range lines {
				if len(line) > colWidths[i] {
					colWidths[i] = utf8.RuneCountInString(line)
				}
			}
		}
	}

	// count the width needed for all the columns
	widthNeeded := 0
	for _, l := range colWidths {
		widthNeeded += l
	}
	widthNeeded += len(tblPadding) * (w.numColumns - 1)
	// +3 to allow room for e.g. scrollbars and other paraphernalia which may be present in people's terminals and
	// doesn't get counted by term.GetSize
	widthNeeded += 3

	RP.RadDebugf(
		fmt.Sprintf("TermWidth: %d, WidthNeeded: %d, ColWidthsBefore: %v\n", termWidth, widthNeeded, colWidths),
	)
	if widthNeeded > termWidth {
		// we're over our size limit, as determined by the width of the terminal.
		// 1. determined the total amount of chars we need to cut down
		// 2. determine the # of chars each column is entitled to (proportionally i.e. if 100 width, 4 columns, each is entitled to 25).
		// 3. determine the # of chars each column is *over* its entitled amount
		// 4. for each column, calculate the % proportion of total 'overspill' that column is responsible for
		// 5. cut down every column breaching its entitlement, in proportion to how responsible they are for the overspill
		charsToReduce := widthNeeded - termWidth
		eachColumnEntitledChars := termWidth / w.numColumns
		charsOverEntitlement := lo.Map(colWidths, func(width int, _ int) int {
			return max(0, width-eachColumnEntitledChars)
		})
		totalCharsOverEntitlement := lo.Sum(charsOverEntitlement)
		proportionOfOver := lo.Map(charsOverEntitlement, func(charsOver int, _ int) float64 {
			return float64(charsOver) / float64(totalCharsOverEntitlement)
		})
		charsToRemove := lo.Map(proportionOfOver, func(proportion float64, _ int) int {
			return int(float64(charsToReduce) * proportion)
		})
		for i, chars := range charsToRemove {
			colWidths[i] -= chars
		}
		RP.RadDebugf(fmt.Sprintf(
			"CharsToReduce: %d, EachColEntitldChars: %d, CharsOverEntitl: %v, TotCharsOverEntitl: %d, PropOfOver: %v, CharsToRm: %v, ColWidthsAfter: %v",
			charsToReduce,
			eachColumnEntitledChars,
			charsOverEntitlement,
			totalCharsOverEntitlement,
			proportionOfOver,
			charsToRemove,
			colWidths,
		))
	}

	// truncate cells to fit within column widths
	rows := w.rows
	for i := range w.headers {
		colWidth := colWidths[i]
		w.tbl.SetColMinWidth(i, colWidth)
		for _, row := range rows {
			cell := row[i]
			lines := strings.Split(cell.Plain(), "\n")
			for j, line := range lines {
				if utf8.RuneCountInString(line) > colWidth &&
					colWidth > 3 { // >3 to prevent slice indexing problem for ellipses below
					// todo in theory we should be wrapping, rather than just cutting off.
					// todo if contents contain escape codes, we may cut them off. Should perhaps be truncating before escape codes.
					if com.TerminalIsUtf8 {
						lines[j] = line[:colWidth-1]
						lines[j] += "…"
					} else {
						lines[j] = line[:colWidth-3]
						lines[j] += "..."
					}
				}
			}
			rejoined := strings.Join(lines, "\n")
			row[i] = cell.CopyAttrTo(rejoined)
		}
	}

	w.tbl.SetAutoFormatHeaders(false)
	w.tbl.SetHeaderAlignment(tblwriter.ALIGN_LEFT)
	w.tbl.SetAlignment(tblwriter.ALIGN_LEFT)
	w.tbl.SetCenterSeparator("")
	w.tbl.SetColumnSeparator("")
	w.tbl.SetRowSeparator("")
	w.tbl.SetAutoWrapText(false)
	w.tbl.SetHeaderLine(false)
	w.tbl.EnableBorder(false)
	w.tbl.SetTablePadding(tblPadding)
	w.tbl.SetNoWhiteSpace(true)

	w.tbl.SetHeader(w.headers)
	var colors []tblwriter.Color
	for range w.headers {
		colors = append(colors, tblwriter.Yellow)
	}
	w.tbl.SetHeaderColors(colors...)

	switch FlagColor.Value {
	case COLOR_NEVER:
		w.tbl.ToggleColor(false)
	case COLOR_ALWAYS:
		w.tbl.ToggleColor(true)
	}

	if len(w.colToColors) > 0 {
		columnModByIdx := make(map[int]tblwriter.ColumnMod)
		for i, header := range w.headers {
			if colColors, ok := w.colToColors[header]; ok {
				var colColorMods []tblwriter.ColumnColorMod
				for _, colorMod := range colColors {
					colColorMods = append(colColorMods, tblwriter.NewColumnColorMod(colorMod.regex, colorMod.color))
				}
				columnModByIdx[i] = tblwriter.NewColumnMod(colColorMods)
			}
		}
		w.tbl.SetColumnMods(columnModByIdx)
	}

	for _, row := range rows {
		rowStr := lo.Map(row, func(cell RadString, _ int) string {
			return cell.String()
		})
		w.tbl.Append(rowStr)
	}

	w.tbl.Render()
}
