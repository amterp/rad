package core

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	tblwriter "github.com/amterp/go-tbl"
	"github.com/samber/lo"
	"golang.org/x/term"
)

const (
	padding = "  "
)

var (
	terminalSupportsUtf8 = terminalIsUtf8()
)

type ColumnSort struct {
	ColIdx int
	Dir    SortDir
}

type TblWriter struct {
	writer      io.Writer
	tbl         *tblwriter.Table
	headers     []string
	rows        [][]string
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

func (w *TblWriter) Append(row []string) {
	w.rows = append(w.rows, row)
	if w.numColumns < len(row) {
		w.numColumns = len(row)
	}
}

func (w *TblWriter) SetColumnColoring(headers []string, colToColors map[string][]radColorMod) {
	for colName, _ := range colToColors {
		if !lo.Contains(headers, colName) {
			RP.RadErrorExit(fmt.Sprintf("Column to color '%s' is not a valid header\n", colName))
		}
	}

	w.colToColors = colToColors
}

func (w *TblWriter) Render() {
	// todo this should almost definitely be mocked out for tests
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		RP.RadDebug(fmt.Sprintf("Error getting terminal width, setting to 9999: %v\n", err))
		termWidth = 9999
	}

	// resolve how many chars each column needs to fully display its contents
	colWidths := make([]int, w.numColumns)
	for i, _ := range w.headers {
		colWidths[i] = len(w.headers[i])
	}
	for _, row := range w.rows {
		for i, cell := range row {
			lines := strings.Split(cell, "\n")
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
	widthNeeded += len(padding) * (w.numColumns - 1)
	// +3 to allow room for e.g. scrollbars and other paraphernalia which may be present in people's terminals and
	// doesn't get counted by term.GetSize
	widthNeeded += 3

	RP.RadDebug(fmt.Sprintf("TermWidth: %d, WidthNeeded: %d, ColWidthsBefore: %v\n", termWidth, widthNeeded, colWidths))
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
		RP.RadDebug(fmt.Sprintf(
			"CharsToReduce: %d, EachColEntitldChars: %d, CharsOverEntitl: %v, TotCharsOverEntitl: %d, PropOfOver: %v, CharsToRm: %v, ColWidthsAfter: %v",
			charsToReduce, eachColumnEntitledChars, charsOverEntitlement, totalCharsOverEntitlement, proportionOfOver, charsToRemove, colWidths))
	}

	// truncate cells to fit within column widths
	rows := w.rows
	for i, _ := range w.headers {
		colWidth := colWidths[i]
		w.tbl.SetColMinWidth(i, colWidth)
		for _, row := range rows {
			lines := strings.Split(row[i], "\n")
			for j, line := range lines {
				if utf8.RuneCountInString(line) > colWidth && colWidth > 3 { // >3 to prevent slice indexing problem for ellipses below
					// todo in theory we should be wrapping, rather than just cutting off.
					if terminalSupportsUtf8 {
						lines[j] = line[:colWidth-1]
						lines[j] += "…"
					} else {
						lines[j] = line[:colWidth-3]
						lines[j] += "..."
					}
				}
			}
			row[i] = strings.Join(lines, "\n")
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
	w.tbl.SetTablePadding(padding)
	w.tbl.SetNoWhiteSpace(true)

	w.tbl.SetHeader(w.headers)
	var colors []tblwriter.Color
	for _, _ = range w.headers {
		colors = append(colors, tblwriter.Yellow)
	}
	w.tbl.SetHeaderColors(colors...)
	w.tbl.ToggleColor(!FlagNoColor.Value)

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
		w.tbl.Append(row)
	}

	w.tbl.Render()
}

func terminalIsUtf8() bool {
	lang := os.Getenv("LANG")
	ctype := os.Getenv("LC_CTYPE")
	// Check for UTF-8 in LANG or LC_CTYPE environment variables
	return strings.Contains(lang, "UTF-8") || strings.Contains(ctype, "UTF-8")
}
