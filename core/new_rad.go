package core

import (
	"fmt"
	"regexp"
	"runtime/debug"

	tblwriter "github.com/amterp/go-tbl"
	"github.com/samber/lo"
	"github.com/scylladb/go-set/strset"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type radInvocation struct {
	i                *Interpreter
	radKeywordNode   *ts.Node
	srcExprNode      *ts.Node
	blockType        RadBlockType
	fields           []*ts.Node
	fieldsToNotPrint *strset.Set
	colToMapOp       map[string]Lambda
}

func (i *Interpreter) runRadBlock(radBlockNode *ts.Node) {
	srcNode := i.getChild(radBlockNode, F_SOURCE)
	radTypeNode := i.getChild(radBlockNode, F_RAD_TYPE)
	typeStr := i.sd.Src[radTypeNode.StartByte():radTypeNode.EndByte()]

	var blockType RadBlockType
	switch typeStr {
	case "rad":
		blockType = Rad
	case "request":
		blockType = Request
	case "display":
		blockType = Display
	default:
		i.errorf(radTypeNode, "Bug! Unknown rad block type %q", typeStr)
	}

	ri := radInvocation{
		i:                i,
		radKeywordNode:   radTypeNode,
		srcExprNode:      srcNode,
		blockType:        blockType,
		fields:           make([]*ts.Node, 0),
		fieldsToNotPrint: strset.New(),
		colToMapOp:       make(map[string]Lambda),
	}

	radStmtNodes := i.getChildren(radBlockNode, F_STMT)
	for _, radStmtNode := range radStmtNodes {
		ri.evalRad(&radStmtNode)
	}

	ri.execute()
}

func (r *radInvocation) evalRad(node *ts.Node) {
	defer func() {
		if re := recover(); re != nil {
			r.i.errorDetailsf(node, fmt.Sprintf("%s\n%s", re, debug.Stack()), "Bug! Panic'd here")
		}
	}()
	r.unsafeEvalRad(node)
}

func (r *radInvocation) unsafeEvalRad(node *ts.Node) {
	switch node.Kind() {
	case K_RAD_FIELD_STMT:
		identifierNodes := r.i.getChildren(node, F_IDENTIFIER)
		for _, identifierNode := range identifierNodes {
			r.fields = append(r.fields, &identifierNode)
		}
	}
}

type radInvocationOld struct {
	i                *Interpreter
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

type radField struct {
	node *ts.Node
	name string
}

func (r *radInvocation) execute() {
	if len(r.fields) == 0 {
		r.i.errorf(r.radKeywordNode, "No fields specified in rad block")
	}

	radFields := lo.Map(r.fields, func(fieldIdentifierNode *ts.Node, _ int) radField {
		name := r.i.sd.Src[fieldIdentifierNode.StartByte():fieldIdentifierNode.EndByte()]
		return radField{node: fieldIdentifierNode, name: name}
	})

	srcStr := r.sourceString()
	if srcStr != nil {
		jsonFields := lo.Map(radFields, func(field radField, _ int) JsonFieldVar {
			fieldVar, ok := r.i.env.GetJsonFieldVar(field.name)
			if !ok {
				r.i.errorf(field.node, "Undefined JSON field %q", field.name)
			}
			return *fieldVar
		})

		data, err := RReq.RequestJson(srcStr.Plain())
		if err != nil {
			r.i.errorf(r.srcExprNode, fmt.Sprintf("Error requesting JSON: %v", err))
		}

		trie := CreateTrie(r.i, r.radKeywordNode, jsonFields)
		trie.TraverseTrie(data)
	}

	headers := lo.FilterMap(radFields, func(field radField, _ int) (string, bool) {
		if r.fieldsToNotPrint.Has(field.name) {
			return "", false
		}
		return field.name, true
	})

	if len(headers) == 0 {
		return
	}

	//applySorting(r.i, fields, r.sorting)

	columns := lo.FilterMap(radFields, func(field radField, _ int) ([]string, bool) {
		if r.fieldsToNotPrint.Has(field.name) {
			return nil, false
		}
		fieldVals, ok := r.i.env.GetVar(field.name)
		if !ok {
			r.i.errorf(field.node, "Values for field %q not found in environment", field.name)
		}
		list := fieldVals.RequireList(r.i, field.node)
		return toTblStr(r.i, r.colToMapOp, field.name, list), true
	})

	tbl := NewTblWriter()

	tbl.SetHeader(headers)
	for i := range columns[0] {
		row := lo.Map(columns, func(column []string, _ int) string {
			return column[i]
		})
		tbl.Append(row)
	}

	//tbl.SetColumnColoring(headers, r.colToColor)

	// todo ensure failed requests get nicely printed
	tbl.Render()
}

func (r *radInvocation) sourceString() *RslString {
	if r.srcExprNode == nil {
		return nil
	}
	str := r.i.evaluate(r.srcExprNode, 1)[0].RequireStr(r.i, r.srcExprNode)
	return &str
}

func applySorting(i *Interpreter, fields []radField, sorting []ColumnSort) {
	if len(sorting) == 0 {
		return
	}

	sortColumns(i, fields, sorting)
}

func toTblStr(i *Interpreter, mapOps map[string]Lambda, fieldName string, column *RslList) []string {
	lambda, ok := mapOps[fieldName]
	if !ok {
		return ToStringArrayQuoteStr(column.Values, false)
	}
	var newVals []string
	for _, val := range column.Values {
		identifier := lambda.Args[0]
		i.runWithChildEnv(func() {
			i.env.SetVarIgnoringEnclosing(identifier, val)
			newVals = append(newVals, ToPrintableQuoteStr(i.evaluate(lambda.ExprNode, 1)[0], false))
		})
	}
	return newVals
}

//func (r *radInvocation) assertHasFields(stmtType string) {
//	if r.fields == nil {
//		r.error(fmt.Sprintf("%s statement must be preceded by a 'fields' statement", stmtType))
//		panic(UNREACHABLE)
//	}
//}

// == fieldModVisitor ==

type fieldModVisitor struct {
	identifiers []Token
	invocation  *radInvocation
}

//func (f fieldModVisitor) VisitColorRadFieldModStmt(color Color) {
//	colorValue := color.ColorValue.Accept(f.invocation.ri.i)
//	switch coerced := colorValue.(type) {
//	case RslString:
//		coercedColor, ok := ColorFromString(coerced.Plain())
//		if !ok {
//			f.invocation.ri.i.error(color.ColorToken, fmt.Sprintf("Invalid color value %q. Allowed: %s",
//				coerced.Plain(), COLOR_STRINGS))
//		}
//		regex := color.Regex.Accept(f.invocation.ri.i)
//		switch coercedRegex := regex.(type) {
//		case RslString:
//			regex, err := regexp.Compile(coercedRegex.Plain())
//			if err != nil {
//				f.invocation.ri.i.error(color.ColorToken, fmt.Sprintf("Error compiling regex pattern: %s", err))
//			}
//			for _, identifier := range f.identifiers {
//				identifierLexeme := identifier.GetLexeme()
//				mods := f.invocation.colToColor[identifierLexeme]
//				mods = append(mods, radColorMod{color: coercedColor.ToTblColor(), regex: regex})
//				f.invocation.colToColor[identifierLexeme] = mods
//			}
//		}
//	default:
//		f.invocation.ri.i.error(color.ColorToken, "Color value must be a string")
//	}
//}
//
//func (f fieldModVisitor) VisitMapModRadFieldModStmt(mapMod MapMod) {
//	for _, identifier := range f.identifiers {
//		identifierLexeme := identifier.GetLexeme()
//		f.invocation.colToMapOp[identifierLexeme] = mapMod.Op
//	}
//}
