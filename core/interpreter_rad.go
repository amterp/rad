package core

import (
	"fmt"
	"regexp"

	tblwriter "github.com/amterp/go-tbl"
	"github.com/samber/lo"
	"github.com/scylladb/go-set/strset"
)

type RadBlockInterpreter struct {
	i          *MainInterpreter
	invocation *radInvocation
}

func NewRadBlockInterpreter(i *MainInterpreter) *RadBlockInterpreter {
	return &RadBlockInterpreter{i: i}
}

func (r RadBlockInterpreter) Run(block RadBlock) {
	var url *string
	if block.Source != nil {
		src := (*block.Source).Accept(r.i) // todo might be a json blob, in the future
		switch coerced := src.(type) {
		case RslString:
			s := coerced.Plain()
			url = &s
		default:
			r.i.error(block.RadKeyword, "URL must be a string")
		}
	}
	r.invocation = &radInvocation{
		ri:               &r,
		block:            block,
		url:              url,
		fields:           nil,
		fieldsToNotPrint: strset.New(),
		colToColor:       make(map[string][]radColorMod),
		colToMapOp:       make(map[string]Lambda),
	}

	for _, stmt := range block.Stmts {
		stmt.Accept(r)
	}

	if block.RadType == Request {
		for _, field := range r.invocation.fields.Identifiers {
			r.invocation.fieldsToNotPrint.Add(field.GetLexeme())
		}
	}

	r.invocation.execute()
	r.invocation = nil
}

func (r RadBlockInterpreter) VisitFieldsRadStmt(fields Fields) {
	r.invocation.fields = &fields
}

func (r RadBlockInterpreter) VisitFieldModsRadStmt(mods FieldMods) {
	r.invocation.assertHasFields("field modifier")

	declaredFields := lo.Map(r.invocation.fields.Identifiers, func(field Token, _ int) string {
		return field.GetLexeme()
	})

	for _, identifier := range mods.Identifiers {
		if !lo.Contains(declaredFields, identifier.GetLexeme()) {
			r.i.error(identifier, fmt.Sprintf("Field %q not found in fields", identifier.GetLexeme()))
		}
	}

	modVisitor := fieldModVisitor{
		identifiers: mods.Identifiers,
		invocation:  r.invocation,
	}
	for _, mod := range mods.Mods {
		mod.Accept(modVisitor)
	}
}

func (r RadBlockInterpreter) VisitSortRadStmt(sort Sort) {
	r.invocation.assertHasFields("sort")
	if sort.GeneralSort != nil {
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

func (r RadBlockInterpreter) VisitRadIfStmtRadStmt(ifStmt RadIfStmt) {
	for _, caseStmt := range ifStmt.Cases {
		val := caseStmt.Condition.Accept(r.i)
		if bval, ok := val.(bool); ok {
			if bval {
				for _, stmt := range caseStmt.Body {
					stmt.Accept(r)
				}
				return
			}
		} else {
			r.i.error(caseStmt.IfToken, fmt.Sprintf("If condition must be a boolean, got %s", TypeAsString(val)))
		}
	}

	if ifStmt.ElseBlock != nil {
		for _, stmt := range *ifStmt.ElseBlock {
			stmt.Accept(r)
		}
	}
}

// == radInvocation ==

type radInvocation struct {
	ri               *RadBlockInterpreter
	block            RadBlock
	url              *string
	fields           *Fields
	fieldsToNotPrint *strset.Set
	sorting          []ColumnSort
	colToColor       map[string][]radColorMod
	colToMapOp       map[string]Lambda
}

type radColorMod struct {
	color tblwriter.Color
	regex *regexp.Regexp
}

func (r *radInvocation) execute() {
	if r.fields == nil {
		// todo instead of just printing, return as string and let user decide what to do with it?
		executeRequestPassthrough(r)
		return
	}

	fields := r.fields.Identifiers
	if r.url != nil {
		jsonFields := lo.Map(fields, func(field Token, _ int) JsonFieldVar {
			return r.ri.i.env.GetJsonField(field)
		})

		data, err := RReq.RequestJson(*r.url)
		if err != nil {
			r.error(fmt.Sprintf("Error requesting JSON: %v", err))
		}

		trie := CreateTrie(r.ri.i, r.block.RadKeyword, jsonFields)
		trie.TraverseTrie(data)
	}

	headers := lo.FilterMap(fields, func(field Token, _ int) (string, bool) {
		if r.fieldsToNotPrint.Has(field.GetLexeme()) {
			return "", false
		}
		return field.GetLexeme(), true
	})

	if len(headers) == 0 {
		return
	}

	applySorting(r.ri.i, r.block.RadKeyword, fields, r.sorting)

	columns := lo.FilterMap(fields, func(field Token, _ int) ([]string, bool) {
		fieldName := field.GetLexeme()
		if r.fieldsToNotPrint.Has(fieldName) {
			return nil, false
		}
		fieldVals := r.ri.i.env.GetByToken(field)
		switch coerced := fieldVals.(type) {
		case []interface{}:
			return toTblStr(r.ri.i, r.colToMapOp, fieldName, coerced), true
		default:
			// could maybe print single value for all rows? so populate an array with appropriate # of values
			r.error(fmt.Sprintf("Field %q must be an array, got %s", fieldName, TypeAsString(fieldVals)))
			panic(UNREACHABLE)
		}
	})

	tbl := NewTblWriter()

	tbl.SetHeader(headers)
	for i := range columns[0] {
		row := lo.Map(columns, func(column []string, _ int) string {
			return column[i]
		})
		tbl.Append(row)
	}

	tbl.SetColumnColoring(headers, r.colToColor)

	// todo ensure failed requests get nicely printed
	tbl.Render()
}

func applySorting(i *MainInterpreter, token Token, fields []Token, sorting []ColumnSort) {
	if len(sorting) == 0 {
		return
	}

	sortColumns(i, token, fields, sorting)
}

func toTblStr(i *MainInterpreter, mapOps map[string]Lambda, fieldName string, column []interface{}) []string {
	lambda, ok := mapOps[fieldName]
	if !ok {
		return ToStringArray(column)
	}
	var newVals []string
	for _, val := range column {
		identifier := lambda.Args[0]
		i.runWithChildEnv(func() {
			i.env.SetAndImplyTypeWithTokenIgnoringEnclosing(identifier, identifier.GetLexeme(), val)
			newVal := lambda.Op.Accept(i)
			newVals = append(newVals, ToPrintable(newVal))
		})
	}
	return newVals
}

// When no fields are specified, we'll simply perform the request and print the output.
func executeRequestPassthrough(r *radInvocation) {
	url := r.url
	if url == nil {
		r.error("Bug! URL should've been validated earlier to be present for passthrough rad block")
		panic(UNREACHABLE)
	}

	// execute request, don't expect responses, just print out the response body
	resp, err := RReq.Get(*url, nil)
	if err != nil {
		r.error(fmt.Sprintf("Error requesting: %v", err))
	}

	data := resp.Body
	// todo weird to even allow this. if we allow returning the data in the future, maybe it'll make sense. and we
	//  would allow just the request block version?
	if r.block.RadType != Request {
		if len(data) == 0 || data[len(data)-1] != '\n' {
			RP.Print(data + "\n")
		} else {
			RP.Print(data)
		}
	}
}

func (r *radInvocation) assertHasFields(stmtType string) {
	if r.fields == nil {
		r.error(fmt.Sprintf("%s statement must be preceded by a 'fields' statement", stmtType))
		panic(UNREACHABLE)
	}
}

func (r *radInvocation) error(msg string) {
	r.ri.i.error(r.block.RadKeyword, msg)
}

// == fieldModVisitor ==

type fieldModVisitor struct {
	identifiers []Token
	invocation  *radInvocation
}

func (f fieldModVisitor) VisitColorRadFieldModStmt(color Color) {
	colorValue := color.ColorValue.Accept(f.invocation.ri.i)
	switch coerced := colorValue.(type) {
	case RslString:
		coercedColor, ok := ColorFromString(coerced.Plain())
		if !ok {
			f.invocation.ri.i.error(color.ColorToken, fmt.Sprintf("Invalid color value %q. Allowed: %s",
				coerced.Plain(), COLOR_STRINGS))
		}
		regex := color.Regex.Accept(f.invocation.ri.i)
		switch coercedRegex := regex.(type) {
		case RslString:
			regex, err := regexp.Compile(coercedRegex.Plain())
			if err != nil {
				f.invocation.ri.i.error(color.ColorToken, fmt.Sprintf("Error compiling regex pattern: %s", err))
			}
			for _, identifier := range f.identifiers {
				identifierLexeme := identifier.GetLexeme()
				mods := f.invocation.colToColor[identifierLexeme]
				mods = append(mods, radColorMod{color: coercedColor.ToTblColor(), regex: regex})
				f.invocation.colToColor[identifierLexeme] = mods
			}
		}
	default:
		f.invocation.ri.i.error(color.ColorToken, "Color value must be a string")
	}
}

func (f fieldModVisitor) VisitMapModRadFieldModStmt(mapMod MapMod) {
	for _, identifier := range f.identifiers {
		identifierLexeme := identifier.GetLexeme()
		f.invocation.colToMapOp[identifierLexeme] = mapMod.Op
	}
}
