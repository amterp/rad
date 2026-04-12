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

// TblWriter wraps go-tbl for Rad's table rendering, adding terminal-width-aware
// truncation and column coloring. Single-use: Render() mutates internal state
// (truncating cells/headers), so a fresh instance must be created for each table.
type TblWriter struct {
	writer      io.Writer
	tbl         *tblwriter.Table
	headers     []string
	rows        [][]RadString
	colToColors map[string][]radColorMod
	numColumns  int
	transpose   bool
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

func (w *TblWriter) SetTranspose(transpose bool) {
	w.transpose = transpose
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

	if w.transpose {
		w.measureAndTruncateTransposed(termWidth)
	} else {
		w.measureAndTruncateNormal(termWidth)
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

	// Column color modifiers are incompatible with transpose: go-tbl's
	// applyTranspose() clears columnModsByIdx, so colors would be silently
	// dropped. Skip setting them and let the caller handle the warning.
	if len(w.colToColors) > 0 && !w.transpose {
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

	for _, row := range w.rows {
		rowStr := lo.Map(row, func(cell RadString, _ int) string {
			return cell.String()
		})
		w.tbl.Append(rowStr)
	}

	w.tbl.SetTranspose(w.transpose)
	w.tbl.Render()
}

// measureAndTruncateNormal measures column widths in the original (non-transposed) layout,
// reduces them if they exceed terminal width, and truncates cell strings to fit.
func (w *TblWriter) measureAndTruncateNormal(termWidth int) {
	colWidths := make([]int, w.numColumns)
	for i := range w.headers {
		colWidths[i] = utf8.RuneCountInString(w.headers[i])
	}
	for _, row := range w.rows {
		for i, cell := range row {
			for _, line := range strings.Split(cell.Plain(), "\n") {
				if runeLen := utf8.RuneCountInString(line); runeLen > colWidths[i] {
					colWidths[i] = runeLen
				}
			}
		}
	}

	colWidths = reduceColWidths(colWidths, termWidth)

	for i := range w.headers {
		colWidth := colWidths[i]
		w.tbl.SetColMinWidth(i, colWidth)
		for _, row := range w.rows {
			row[i] = truncateCell(row[i], colWidth)
		}
	}
}

// measureAndTruncateTransposed measures column widths in the transposed layout
// (col 0 = header names, cols 1..N = one per original data row), reduces them
// if they exceed terminal width, and truncates cell strings to fit.
func (w *TblWriter) measureAndTruncateTransposed(termWidth int) {
	numTransposedCols := len(w.rows) + 1
	colWidths := make([]int, numTransposedCols)

	// transposed col 0: header names stacked vertically
	for _, h := range w.headers {
		if runeLen := utf8.RuneCountInString(h); runeLen > colWidths[0] {
			colWidths[0] = runeLen
		}
	}

	// transposed col j+1: all field values from original row j
	for rowIdx, row := range w.rows {
		for _, cell := range row {
			for _, line := range strings.Split(cell.Plain(), "\n") {
				if runeLen := utf8.RuneCountInString(line); runeLen > colWidths[rowIdx+1] {
					colWidths[rowIdx+1] = runeLen
				}
			}
		}
	}

	colWidths = reduceColWidths(colWidths, termWidth)

	// truncate headers (they become values in transposed col 0)
	headerWidth := colWidths[0]
	for i, h := range w.headers {
		w.headers[i] = truncateString(h, headerWidth)
	}

	// truncate cells based on their transposed column position
	for rowIdx, row := range w.rows {
		maxWidth := colWidths[rowIdx+1]
		for colIdx := range row {
			row[colIdx] = truncateCell(row[colIdx], maxWidth)
		}
	}
	// skip SetColMinWidth - go-tbl's applyTranspose rebuilds column sizing
}

// reduceColWidths returns a new slice with column widths proportionally reduced
// to fit within the terminal width. If all columns already fit, returns a copy unchanged.
func reduceColWidths(colWidths []int, termWidth int) []int {
	numColumns := len(colWidths)
	if numColumns == 0 {
		return nil
	}

	result := make([]int, numColumns)
	copy(result, colWidths)

	widthNeeded := 0
	for _, l := range result {
		widthNeeded += l
	}
	widthNeeded += len(tblPadding) * (numColumns - 1)
	// +3 to allow room for e.g. scrollbars and other paraphernalia which may be present in people's terminals and
	// doesn't get counted by term.GetSize
	widthNeeded += 3

	RP.RadDebugf(
		fmt.Sprintf("TermWidth: %d, WidthNeeded: %d, ColWidthsBefore: %v\n", termWidth, widthNeeded, result),
	)
	if widthNeeded <= termWidth {
		return result
	}

	// We're over our size limit, as determined by the width of the terminal.
	// 1. Determine the total amount of chars we need to cut down
	// 2. Determine the # of chars each column is entitled to (proportionally)
	// 3. Determine the # of chars each column is *over* its entitled amount
	// 4. For each column, calculate the % proportion of total 'overspill' that column is responsible for
	// 5. Cut down every column breaching its entitlement, in proportion to how responsible they are for the overspill
	charsToReduce := widthNeeded - termWidth
	eachColumnEntitledChars := termWidth / numColumns
	charsOverEntitlement := lo.Map(result, func(width int, _ int) int {
		return max(0, width-eachColumnEntitledChars)
	})
	totalCharsOverEntitlement := lo.Sum(charsOverEntitlement)

	if totalCharsOverEntitlement > 0 {
		// At least one column exceeds its entitlement - reduce proportionally
		proportionOfOver := lo.Map(charsOverEntitlement, func(charsOver int, _ int) float64 {
			return float64(charsOver) / float64(totalCharsOverEntitlement)
		})
		charsToRemove := lo.Map(proportionOfOver, func(proportion float64, _ int) int {
			return int(float64(charsToReduce) * proportion)
		})
		for i, chars := range charsToRemove {
			result[i] -= chars
		}
	} else {
		// All columns are within entitlement but collectively overflow
		// (e.g. padding overhead pushes us over). Distribute reduction evenly.
		perCol := charsToReduce / numColumns
		remainder := charsToReduce % numColumns
		for i := range result {
			reduce := perCol
			if i < remainder {
				reduce++
			}
			result[i] -= reduce
		}
	}

	RP.RadDebugf(fmt.Sprintf(
		"CharsToReduce: %d, EachColEntitldChars: %d, CharsOverEntitl: %v, TotCharsOverEntitl: %d, ColWidthsAfter: %v",
		charsToReduce,
		eachColumnEntitledChars,
		charsOverEntitlement,
		totalCharsOverEntitlement,
		result,
	))
	return result
}

// truncateCell truncates a RadString cell's content to fit within maxWidth.
func truncateCell(cell RadString, maxWidth int) RadString {
	lines := strings.Split(cell.Plain(), "\n")
	changed := false
	for j, line := range lines {
		truncated := truncateString(line, maxWidth)
		if truncated != line {
			lines[j] = truncated
			changed = true
		}
	}
	if !changed {
		return cell
	}
	return cell.CopyAttrTo(strings.Join(lines, "\n"))
}

// truncateString truncates a single string to fit within maxWidth runes,
// appending an ellipsis if truncation occurs.
func truncateString(s string, maxWidth int) string {
	if utf8.RuneCountInString(s) <= maxWidth {
		return s
	}
	// todo in theory we should be wrapping, rather than just cutting off.
	// todo if contents contain escape codes, we may cut them off. Should perhaps be truncating before escape codes.
	if com.TerminalIsUtf8 {
		if maxWidth <= 1 {
			return s
		}
		return string([]rune(s)[:maxWidth-1]) + "…"
	}
	if maxWidth <= 3 {
		return s
	}
	return string([]rune(s)[:maxWidth-3]) + "..."
}
