package core

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
	"io"
	"os"
	"strings"
)

type TblWriter struct {
	printer    Printer
	writer     io.Writer
	tbl        *tablewriter.Table
	headers    []string
	rows       [][]string
	numColumns int
}

func NewTblWriter(printer Printer) *TblWriter {
	stdWriter := printer.GetStdWriter()
	return &TblWriter{
		printer:    printer,
		writer:     stdWriter,
		tbl:        tablewriter.NewWriter(stdWriter),
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

func (w *TblWriter) Render() {
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd())) // todo how does this work when embedded in bash?
	if err != nil {
		w.printer.RadErrorExit(fmt.Sprintf("Error getting terminal width: %v\n", err))
	}

	colWidths := make([]int, w.numColumns)
	for i, _ := range w.headers {
		colWidths[i] = len(w.headers[i])
	}
	for _, row := range w.rows {
		for i, cell := range row {
			lines := strings.Split(cell, "\n")
			for _, line := range lines {
				if len(line) > colWidths[i] {
					colWidths[i] = len(line)
				}
			}
		}
	}
	widthNeeded := 0
	for _, l := range colWidths {
		widthNeeded += l
	}
	w.printer.RadDebug(fmt.Sprintf("TermWidth: %d, WidthNeeded: %d, ColWidthsBefore: %v\n", termWidth, widthNeeded, colWidths))
	if widthNeeded > termWidth {
		percentToKeep := float64(termWidth) / float64(widthNeeded)
		w.printer.RadDebug(fmt.Sprintf("PercentToKeep: %f\n", percentToKeep))
		for i, _ := range w.headers {
			colWidths[i] = int(float64(colWidths[i]) * percentToKeep)
			// todo improve this algo, it's not great to cut 1-2 chars off small columns to make room for large columns
		}
	}
	w.printer.RadDebug(fmt.Sprintf("ColWidthsAfter: %v\n", colWidths))

	rows := w.rows
	for i, _ := range w.headers {
		colWidth := colWidths[i]
		w.tbl.SetColMinWidth(i, colWidth)
		for _, row := range rows {
			lines := strings.Split(row[i], "\n")
			for j, line := range lines {
				if len(line) > colWidth {
					lines[j] = line[:colWidth]
				}
			}
			row[i] = strings.Join(lines, "\n")
		}
	}

	w.tbl.SetAutoWrapText(false)
	w.tbl.SetAutoFormatHeaders(true)
	w.tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	w.tbl.SetAlignment(tablewriter.ALIGN_LEFT)
	w.tbl.SetCenterSeparator("")
	w.tbl.SetColumnSeparator("")
	w.tbl.SetRowSeparator("")
	w.tbl.SetHeaderLine(false)
	w.tbl.SetBorder(false)
	w.tbl.SetTablePadding("\t")
	w.tbl.SetNoWhiteSpace(true)

	w.tbl.SetHeader(w.headers)
	for _, row := range rows {
		w.tbl.Append(row)
	}

	w.tbl.Render()
}
