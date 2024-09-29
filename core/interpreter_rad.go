package core

import (
	"fmt"
	"github.com/samber/lo"
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

func (r RadBlockInterpreter) VisitSortRadStmt(sort Sort) {
	if sort.GeneralSort != nil {
		// depend on the fact that field stmt must be the first thing in the block, and so already visited
		for i, _ := range r.invocation.fields.Identifiers {
			r.invocation.sorting = append(r.invocation.sorting, ColumnSort{ColIdx: i, Dir: *sort.GeneralSort})
		}
		return
	}

	fieldToIdx := make(map[string]int)
	for i, identifier := range r.invocation.fields.Identifiers {
		fieldToIdx[identifier.GetLexeme()] = i
	}

	for i, identifier := range sort.Identifiers {
		dir := sort.Directions[i]
		if idx, ok := fieldToIdx[identifier.GetLexeme()]; !ok {
			r.i.error(r.invocation.block.RadKeyword, fmt.Sprintf("Sort field '%s' not found in fields", identifier.GetLexeme()))
		} else {
			r.invocation.sorting = append(r.invocation.sorting, ColumnSort{ColIdx: idx, Dir: dir})
		}
	}
}

// == radInvocation ==

type radInvocation struct {
	ri      *RadBlockInterpreter
	block   RadBlock
	url     string
	fields  Fields
	sorting []ColumnSort
}

func (r *radInvocation) execute() {
	data, err := RReq.RequestJson(r.url)
	if err != nil {
		r.error(fmt.Sprintf("Error requesting JSON: %v", err))
	}

	jsonFields := lo.Map(r.fields.Identifiers, func(field Token, _ int) JsonFieldVar {
		return r.ri.i.env.GetJsonField(field)
	})
	trie := CreateTrie(r.block.RadKeyword, jsonFields)
	trie.TraverseTrie(data)

	columns := lo.Map(jsonFields, func(field JsonFieldVar, _ int) []string {
		return ToStringArray(r.ri.i.env.GetByToken(field.Name).GetMixedArray())
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

	tbl.SetSorting(r.sorting)

	// todo ensure failed requests get nicely printed
	tbl.Render()
}

func (r *radInvocation) error(msg string) {
	r.ri.i.error(r.block.RadKeyword, msg)
}
