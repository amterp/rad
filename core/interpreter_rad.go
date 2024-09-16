package core

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"io"
	"net/http"
	"net/url"
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
	r.invocation = &radInvocation{ri: &r, block: block, url: url.(string)}
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
	block  RadBlock
	url    string
	fields Fields
}

func (r *radInvocation) execute() {
	urlToQuery := r.encodeUrl(r.url)
	RP.Print(fmt.Sprintf("Querying url: %s\n", urlToQuery))
	resp, err := http.Get(urlToQuery)
	if err != nil {
		r.error(fmt.Sprintf("Error on HTTP request: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.error(fmt.Sprintf("Error reading HTTP body: %v. Body: %v", err, body))
	}

	isValidJson := json.Valid(body)
	if !isValidJson {
		r.error(fmt.Sprintf("Received invalid JSON in response (truncated max 50 chars): [%s]", body[:50]))
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		r.error(fmt.Sprintf("Error unmarshalling JSON: %v", err))
	}

	jsonFields := lo.Map(r.fields.Identifiers, func(field Token, _ int) JsonFieldVar {
		return r.ri.i.env.GetJsonField(field)
	})
	trie := CreateTrie(r.block.RadKeyword, jsonFields)
	trie.TraverseTrie(data)

	columns := lo.Map(jsonFields, func(field JsonFieldVar, _ int) []string {
		return r.ri.i.env.GetByToken(field.Name).GetStringArray()
	})

	tbl := NewTblWriter()

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

	// todo ensure failed requests get nicely printed
	tbl.Render()
}

// todo test this more, might need additional query param encoding
func (r *radInvocation) encodeUrl(rawUrl string) string {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		r.error(fmt.Sprintf("Error parsing URL: %v", err))
	}
	return parsedUrl.String()
}

func (r *radInvocation) error(msg string) {
	r.ri.i.error(r.block.RadKeyword, msg)
}
