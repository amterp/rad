package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RadBlockInterpreter struct {
	i          *MainInterpreter
	invocation *RadInvocation
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
	r.invocation = &RadInvocation{url: url.(string)}
	for _, stmt := range block.Stmts {
		stmt.Accept(r)
	}
	r.invocation.execute()
	r.invocation = nil
}

func (r RadBlockInterpreter) VisitFieldsRadStmt(fields Fields) {
	r.invocation.fields = fields
}

// == RadInvocation ==

type RadInvocation struct {
	url    string
	fields Fields
}

func (r *RadInvocation) execute() {
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
	fmt.Printf("Response: %v\n", data)
}
