package core

import (
	"fmt"
	"regexp"
	"runtime/debug"

	"github.com/amterp/rts/rsl"

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
	// if no specific column specified for sorting
	generalSort *GeneralSort
	// if specific columns listed for sorting, mutually exclusive with generalSort
	// in-order of sorting priority
	colWiseSorting []ColumnSort
	colToMods      map[string]*radFieldMods
}

type radFieldMods struct {
	identifierNode *ts.Node
	colors         []radColorMod
	lambda         *Lambda
}

func newRadFieldMods(identifierNode *ts.Node) *radFieldMods {
	return &radFieldMods{
		identifierNode: identifierNode,
		colors:         make([]radColorMod, 0),
	}
}

func (i *Interpreter) runRadBlock(radBlockNode *ts.Node) {
	srcNode := i.getChild(radBlockNode, rsl.F_SOURCE)
	radTypeNode := i.getChild(radBlockNode, rsl.F_RAD_TYPE)
	typeStr := i.sd.Src[radTypeNode.StartByte():radTypeNode.EndByte()]

	var blockType RadBlockType
	switch typeStr {
	case rsl.KEYWORD_RAD:
		blockType = Rad
	case rsl.KEYWORD_REQUEST:
		blockType = Request
	case rsl.KEYWORD_DISPLAY:
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
		colWiseSorting:   make([]ColumnSort, 0),
		colToMods:        make(map[string]*radFieldMods),
	}

	radStmtNodes := i.getChildren(radBlockNode, rsl.F_STMT)
	for _, radStmtNode := range radStmtNodes {
		ri.evalRad(&radStmtNode)
	}

	ri.execute()
}

func (r *radInvocation) evalRad(node *ts.Node) {
	if !IsTest {
		defer func() {
			if re := recover(); re != nil {
				r.i.errorDetailsf(node, fmt.Sprintf("%s\n%s", re, debug.Stack()), "Bug! Panic'd here")
			}
		}()
	}
	r.unsafeEvalRad(node)
}

func (r *radInvocation) unsafeEvalRad(node *ts.Node) {
	switch node.Kind() {
	case rsl.K_RAD_FIELD_STMT:
		// todo validate no field names conflict with keywords e.g. 'asc'. Would be nice to do in static analysis tho.
		identifierNodes := r.i.getChildren(node, rsl.F_IDENTIFIER)
		for _, identifierNode := range identifierNodes {
			r.fields = append(r.fields, &identifierNode)
		}
	case rsl.K_RAD_SORT_STMT:
		if r.generalSort != nil || len(r.colWiseSorting) > 0 {
			r.i.errorf(node, "Only one sort statement allowed per rad block")
		}

		specifierNodes := r.i.getChildren(node, rsl.F_SPECIFIER)
		if len(specifierNodes) == 0 {
			r.generalSort = &GeneralSort{
				Node: node,
				Dir:  Asc,
			}
		} else {
			for _, specifierNode := range specifierNodes {
				r.evalRad(&specifierNode)
			}
		}
	case rsl.K_RAD_SORT_SPECIFIER:
		firstNode := r.i.getChild(node, rsl.F_FIRST) // we can assume this non-nil, otherwise this node wouldn't exist
		secondNode := r.i.getChild(node, rsl.F_SECOND)

		if secondNode == nil {
			firstNodeSrc := r.i.sd.Src[firstNode.StartByte():firstNode.EndByte()]
			if firstNodeSrc == rsl.KEYWORD_ASC || firstNodeSrc == rsl.KEYWORD_DESC {
				dir := lo.Ternary(firstNodeSrc == rsl.KEYWORD_ASC, Asc, Desc)
				r.generalSort = &GeneralSort{
					Node: node,
					Dir:  dir,
				}
				return
			}
		}

		dir := Asc
		if secondNode != nil {
			switch secondNode.Kind() {
			case rsl.K_ASC:
				dir = Asc
			case rsl.K_DESC:
				dir = Desc
			default:
				r.i.errorf(secondNode, "Bug! Unknown direction %q", secondNode.Kind())
			}
		}

		r.colWiseSorting = append(r.colWiseSorting, ColumnSort{
			ColIdentifier: firstNode,
			Dir:           dir,
		})
	case rsl.K_RAD_FIELD_MODIFIER_STMT:
		identifierNodes := r.i.getChildren(node, rsl.F_IDENTIFIER)
		stmtNodes := r.i.getChildren(node, rsl.F_MOD_STMT)
		var fields []radField
		for _, identifierNode := range identifierNodes {
			identifierStr := r.i.sd.Src[identifierNode.StartByte():identifierNode.EndByte()]
			fields = append(fields, radField{
				node: &identifierNode,
				name: identifierStr,
			})
		}
		for _, stmtNode := range stmtNodes {
			switch stmtNode.Kind() {
			case rsl.K_RAD_FIELD_MOD_COLOR:
				// todo could I replace this syntax with a 'map' lambda operation?
				clrExprNode := r.i.getChild(&stmtNode, rsl.F_COLOR)
				clrStr := r.i.evaluate(clrExprNode, 1)[0].RequireStr(r.i, clrExprNode)
				clr := AttrFromString(r.i, clrExprNode, clrStr.Plain())
				regexExprNode := r.i.getChild(&stmtNode, rsl.F_REGEX)
				regexStr := r.i.evaluate(regexExprNode, 1)[0].RequireStr(r.i, regexExprNode)
				regex, err := regexp.Compile(regexStr.Plain())
				if err != nil {
					r.i.errorf(regexExprNode, fmt.Sprintf("Invalid regex pattern: %s", err))
				}
				for _, field := range fields {
					mods := r.loadFieldMods(field)
					mods.colors = append(mods.colors, radColorMod{color: clr.ToTblColor(), regex: regex})
				}
			case rsl.K_RAD_FIELD_MOD_MAP:
				lambdaNode := r.i.getChild(&stmtNode, rsl.F_LAMBDA)
				lambdaIdentifierNodes := r.i.getChildren(lambdaNode, rsl.F_IDENTIFIER)
				var lambdaIdentifiers []string
				for _, lambdaIdentifierNode := range lambdaIdentifierNodes {
					lambdaIdentifier := r.i.sd.Src[lambdaIdentifierNode.StartByte():lambdaIdentifierNode.EndByte()]
					lambdaIdentifiers = append(lambdaIdentifiers, lambdaIdentifier)
				}
				exprNode := r.i.getChild(lambdaNode, rsl.F_EXPR)
				lambda := Lambda{
					Node:     lambdaNode,
					Args:     lambdaIdentifiers,
					ExprNode: exprNode,
				}
				for _, field := range fields {
					mods := r.loadFieldMods(field)
					mods.lambda = &lambda
				}
			}
		}
	case rsl.K_RAD_IF_STMT:
		altNodes := r.i.getChildren(node, rsl.F_ALT)
		for _, altNode := range altNodes {
			condNode := r.i.getChild(&altNode, rsl.F_CONDITION)

			shouldExecute := true
			if condNode != nil {
				condResult := r.i.evaluate(condNode, 1)[0].TruthyFalsy()
				shouldExecute = condResult
			}

			if shouldExecute {
				stmtNodes := r.i.getChildren(&altNode, rsl.F_STMT)
				for _, stmtNode := range stmtNodes {
					r.evalRad(&stmtNode)
				}
				break
			}
		}
	}
}

func (r *radInvocation) loadFieldMods(field radField) *radFieldMods {
	mods, ok := r.colToMods[field.name]
	if !ok {
		mods = newRadFieldMods(field.node)
		r.colToMods[field.name] = mods
	}
	return mods
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

	// check all field mods are for fields that actually exist
	fieldNames := lo.Map(radFields, func(f radField, _ int) string { return f.name })
	for field, mods := range r.colToMods {
		if !lo.Contains(fieldNames, field) {
			r.i.errorf(mods.identifierNode, "Cannot modify undefined field %q", field)
		}
	}

	data, err := r.resolveData()
	if err != nil {
		r.i.errorf(r.srcExprNode, fmt.Sprintf("Error resolving data: %v", err))
	}

	if data != nil {
		jsonFields := lo.Map(radFields, func(field radField, _ int) JsonFieldVar {
			fieldVar, ok := r.i.env.GetJsonFieldVar(field.name)
			if !ok {
				r.i.errorf(field.node, "Undefined JSON field %q", field.name)
			}
			return *fieldVar
		})

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

	applySorting(r.i, radFields, r.generalSort, r.colWiseSorting)

	if r.blockType == Request {
		return
	}

	columns := lo.FilterMap(radFields, func(field radField, _ int) ([]string, bool) {
		if r.fieldsToNotPrint.Has(field.name) {
			return nil, false
		}
		fieldVals, ok := r.i.env.GetVar(field.name)
		if !ok {
			r.i.errorf(field.node, "Values for field %q not found in environment", field.name)
		}
		list := fieldVals.RequireList(r.i, field.node)
		return toTblStr(r.i, r.colToMods, field.name, list), true
	})

	tbl := NewTblWriter()

	tbl.SetHeader(headers)
	for i := range columns[0] {
		row := lo.Map(columns, func(column []string, _ int) string {
			return column[i]
		})
		tbl.Append(row)
	}

	tbl.SetColumnColoring(r.colToMods)

	// todo ensure failed requests get nicely printed
	tbl.Render()
}

func (r *radInvocation) resolveData() (data interface{}, err error) {
	if r.srcExprNode == nil {
		return nil, nil
	}

	src := r.i.evaluate(r.srcExprNode, 1)[0]

	if r.blockType == Rad || r.blockType == Request {
		str := src.RequireStr(r.i, r.srcExprNode)
		return RReq.RequestJson(str.Plain())
	}

	if r.blockType == Display {
		visitor := NewTypeVisitor(r.i, r.srcExprNode)
		visitor.VisitList = func(val RslValue, _ *RslList) {
			data = RslToJsonType(val)
		}
		visitor.VisitMap = func(val RslValue, _ *RslMap) {
			data = RslToJsonType(val)
		}
		visitor.Default = func(val RslValue) {
			r.i.errorf(r.srcExprNode, "Display block source can only be a list or a map. Got %q", TypeAsString(val))
		}
		src.Accept(visitor, true)
		return
	} else {
		r.i.errorf(r.srcExprNode, "Bug! Unknown rad block type %q", r.blockType)
		panic(UNREACHABLE)
	}
}

func applySorting(i *Interpreter, fields []radField, generalSort *GeneralSort, colWiseSort []ColumnSort) {
	if generalSort != nil {
		if len(colWiseSort) > 0 {
			i.errorf(generalSort.Node, "Bug! General and column-wise sort expected to be mutually exclusive")
		}
		for _, field := range fields {
			colWiseSort = append(colWiseSort, ColumnSort{
				ColIdentifier: field.node,
				Dir:           generalSort.Dir,
			})
		}
	}

	sortColumns(i, fields, colWiseSort)
}

func toTblStr(i *Interpreter, colToMods map[string]*radFieldMods, fieldName string, column *RslList) []string {
	mods, ok := colToMods[fieldName]
	if !ok || mods.lambda == nil {
		return ToStringArrayQuoteStr(column.Values, false)
	}
	var newVals []string
	for _, val := range column.Values {
		identifier := mods.lambda.Args[0]
		i.runWithChildEnv(func() {
			i.env.SetVarIgnoringEnclosing(identifier, val)
			newVal := i.evaluate(mods.lambda.ExprNode, 1)[0]
			newVals = append(newVals, ToPrintableQuoteStr(newVal, false))
		})
	}
	return newVals
}
