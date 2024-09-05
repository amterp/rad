package core

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/samber/lo"
	"io"
	"net/http"
	"os"
)

type RadBlockInterpreter struct {
	i          *MainInterpreter
	invocation *radInvocation
}

func NewRadBlockInterpreter(i *MainInterpreter) *RadBlockInterpreter {
	return &RadBlockInterpreter{i: i}
}

func (r RadBlockInterpreter) Run(block RadBlock) {
	url := block.Url.Accept(r.i)
	switch url.(type) {
	case string:
		break
	default:
		r.i.error(block.RadKeyword, "URL must be a string")
	}
	r.invocation = &radInvocation{ri: &r, url: url.(string)}
	for _, stmt := range block.Stmts {
		stmt.Accept(r)
	}
	r.invocation.execute()
	r.invocation = nil
}

func (r RadBlockInterpreter) VisitFieldsRadStmt(fields Fields) {
	r.invocation.fields = fields
}

// == radInvocation ==

type radInvocation struct {
	ri     *RadBlockInterpreter
	url    string
	fields Fields
}

func (r *radInvocation) execute() {
	fmt.Printf("Querying URL: %s\n", r.url)
	resp, err := http.Get(r.url)
	if err != nil {
		panic(fmt.Sprintf("Error on HTTP request: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error reading HTTP body: %v. Body: %v", err, body))
	}

	isValidJson := json.Valid(body)
	if !isValidJson {
		panic(fmt.Sprintf("Received invalid JSON in response (truncated max 50 chars): [%s]", body[:50]))
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		panic(fmt.Sprintf("Error unmarshalling JSON: %v", err))
	}

	jsonFields := lo.Map(r.fields.Identifiers, func(field Token, _ int) JsonFieldVar {
		return r.ri.i.env.GetJsonField(field)
	})
	trie := CreateTrie(jsonFields)
	TraverseTrie(data, trie)

	columns := lo.Map(jsonFields, func(field JsonFieldVar, _ int) []string {
		return r.ri.i.env.GetByToken(field.Name).GetStringArray()
	})

	tbl := tablewriter.NewWriter(os.Stdout)

	headers := lo.Map(jsonFields, func(field JsonFieldVar, _ int) string {
		return field.Name.GetLexeme()
	})

	tbl.SetHeader(headers)
	for i := range columns[0] {
		row := lo.Map(columns, func(column []string, _ int) string {
			return column[i]
		})
		tbl.Append(row)
	}

	// default formatting
	tbl.SetAutoWrapText(false)
	tbl.SetAutoFormatHeaders(true)
	tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetCenterSeparator("")
	tbl.SetColumnSeparator("")
	tbl.SetRowSeparator("")
	tbl.SetHeaderLine(false)
	tbl.SetBorder(false)
	tbl.SetTablePadding("\t") // pad with tabs
	tbl.SetNoWhiteSpace(true)

	tbl.Render()
}
