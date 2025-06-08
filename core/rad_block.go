package core

import (
	"fmt"
	com "rad/core/common"
	"regexp"
	"runtime/debug"

	"github.com/amterp/rad/rts/rl"

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
	lambda         *RadFn
}

func newRadFieldMods(identifierNode *ts.Node) *radFieldMods {
	return &radFieldMods{
		identifierNode: identifierNode,
		colors:         make([]radColorMod, 0),
	}
}

func (i *Interpreter) runRadBlock(radBlockNode *ts.Node) {
	srcNode := i.getChild(radBlockNode, rl.F_SOURCE)
	radTypeNode := i.getChild(radBlockNode, rl.F_RAD_TYPE)
	typeStr := i.sd.Src[radTypeNode.StartByte():radTypeNode.EndByte()]

	var blockType RadBlockType
	switch typeStr {
	case rl.KEYWORD_RAD:
		blockType = RadBlock
	case rl.KEYWORD_REQUEST:
		blockType = RequestBlock
	case rl.KEYWORD_DISPLAY:
		blockType = DisplayBlock
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

	radStmtNodes := i.getChildren(radBlockNode, rl.F_STMT)
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
	case rl.K_RAD_FIELD_STMT:
		// todo validate no field names conflict with keywords e.g. 'asc'. Would be nice to do in static analysis tho.
		identifierNodes := r.i.getChildren(node, rl.F_IDENTIFIER)
		for _, identifierNode := range identifierNodes {
			r.fields = append(r.fields, &identifierNode)
		}
	case rl.K_RAD_SORT_STMT:
		if r.generalSort != nil || len(r.colWiseSorting) > 0 {
			r.i.errorf(node, "Only one sort statement allowed per rad block")
		}

		specifierNodes := r.i.getChildren(node, rl.F_SPECIFIER)
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
	case rl.K_RAD_SORT_SPECIFIER:
		firstNode := r.i.getChild(node, rl.F_FIRST) // we can assume this non-nil, otherwise this node wouldn't exist
		secondNode := r.i.getChild(node, rl.F_SECOND)

		if secondNode == nil {
			firstNodeSrc := r.i.sd.Src[firstNode.StartByte():firstNode.EndByte()]
			if firstNodeSrc == rl.KEYWORD_ASC || firstNodeSrc == rl.KEYWORD_DESC {
				dir := lo.Ternary(firstNodeSrc == rl.KEYWORD_ASC, Asc, Desc)
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
			case rl.K_ASC:
				dir = Asc
			case rl.K_DESC:
				dir = Desc
			default:
				r.i.errorf(secondNode, "Bug! Unknown direction %q", secondNode.Kind())
			}
		}

		r.colWiseSorting = append(r.colWiseSorting, ColumnSort{
			ColIdentifier: firstNode,
			Dir:           dir,
		})
	case rl.K_RAD_FIELD_MODIFIER_STMT:
		identifierNodes := r.i.getChildren(node, rl.F_IDENTIFIER)
		stmtNodes := r.i.getChildren(node, rl.F_MOD_STMT)
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
			case rl.K_RAD_FIELD_MOD_COLOR:
				// todo could I replace this syntax with a 'map' lambda operation?
				clrExprNode := r.i.getChild(&stmtNode, rl.F_COLOR)
				clrStr := r.i.evaluate(clrExprNode, EXPECT_ONE_OUTPUT).RequireStr(r.i, clrExprNode)
				clr := AttrFromString(r.i, clrExprNode, clrStr.Plain())
				regexExprNode := r.i.getChild(&stmtNode, rl.F_REGEX)
				regexStr := r.i.evaluate(regexExprNode, EXPECT_ONE_OUTPUT).RequireStr(r.i, regexExprNode)
				regex, err := regexp.Compile(regexStr.Plain())
				if err != nil {
					r.i.errorf(regexExprNode, fmt.Sprintf("Invalid regex pattern: %s", err))
				}
				for _, field := range fields {
					mods := r.loadFieldMods(field)
					mods.colors = append(mods.colors, radColorMod{color: clr.ToTblColor(), regex: regex})
				}
			case rl.K_RAD_FIELD_MOD_MAP:
				lambdaNode := r.i.getChild(&stmtNode, rl.F_LAMBDA)

				var lambda RadFn
				if lambdaNode.Kind() == rl.K_FN_LAMBDA {
					lambda = NewLambda(r.i, lambdaNode)
				} else if lambdaNode.Kind() == rl.K_IDENTIFIER {
					identifier := GetSrc(r.i.sd.Src, lambdaNode)
					val, ok := r.i.env.GetVar(identifier)
					if !ok {
						r.i.errorf(lambdaNode, "Undefined lambda %q", identifier)
					}
					lambda, ok = val.TryGetFn()
					if !ok {
						r.i.errorf(lambdaNode, "Expected function, got '%s'", TypeAsString(val))
					}
				} else {
					r.i.errorf(lambdaNode, "Bug! Unknown lambda type %q", lambdaNode.Kind())
				}

				for _, field := range fields {
					mods := r.loadFieldMods(field)
					mods.lambda = &lambda
				}
			}
		}
	case rl.K_RAD_IF_STMT:
		altNodes := r.i.getChildren(node, rl.F_ALT)
		for _, altNode := range altNodes {
			condNode := r.i.getChild(&altNode, rl.F_CONDITION)

			shouldExecute := true
			if condNode != nil {
				condResult := r.i.evaluate(condNode, EXPECT_ONE_OUTPUT).TruthyFalsy()
				shouldExecute = condResult
			}

			if shouldExecute {
				stmtNodes := r.i.getChildren(&altNode, rl.F_STMT)
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

	if r.blockType == RequestBlock {
		return
	}

	longestColumnLen := 0
	cellsRowThenColumn := lo.FilterMap(radFields, func(field radField, _ int) ([]RadString, bool) {
		if r.fieldsToNotPrint.Has(field.name) {
			return nil, false
		}
		fieldVals, ok := r.i.env.GetVar(field.name)
		if !ok {
			r.i.errorf(field.node, "Values for field %q not found in environment", field.name)
		}
		columnValues := fieldVals.RequireList(r.i, field.node)
		longestColumnLen = com.IntMax(longestColumnLen, columnValues.LenInt())
		return columnStrings(r.i, r.colToMods, field.name, columnValues), true
	})

	tbl := NewTblWriter()

	tbl.SetHeader(headers)

	// transform columnar data to rows and append to table
	for i := range longestColumnLen {
		row := lo.Map(cellsRowThenColumn, func(column []RadString, _ int) RadString {
			if i >= len(column) {
				return EMPTY_STR
			}
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

	src := r.i.evaluate(r.srcExprNode, EXPECT_ONE_OUTPUT)

	if r.blockType == RadBlock || r.blockType == RequestBlock {
		str := src.RequireStr(r.i, r.srcExprNode)
		return RReq.RequestJson(str.Plain())
	}

	if r.blockType == DisplayBlock {
		NewTypeVisitor(r.i, r.srcExprNode).ForList(func(val RadValue, _ *RadList) {
			data = RadToJsonType(val)
		}).ForMap(func(val RadValue, _ *RadMap) {
			data = RadToJsonType(val)
		}).ForDefault(func(val RadValue) {
			r.i.errorf(r.srcExprNode, "Display block source can only be a list or a map. Got %q", TypeAsString(val))
		}).Visit(src)
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

func columnStrings(i *Interpreter, colToMods map[string]*radFieldMods, fieldName string, column *RadList) []RadString {
	mods, ok := colToMods[fieldName]
	if !ok || mods.lambda == nil {
		return toStringArrayQuoteStr(column.Values, false)
	}

	reprNode := mods.lambda.ReprNode
	var newVals []RadString
	for _, val := range column.Values {
		mapped := mods.lambda.Execute(NewFuncInvocationArgs(i, reprNode, FUNC_MAP, NewPosArgs(NewPosArg(reprNode, val)), NO_NAMED_ARGS_INPUT, EXPECT_ONE_OUTPUT))
		newVals = append(newVals, toStringQuoteStr(mapped, false))
	}

	return newVals
}

func toStringArrayQuoteStr(v []RadValue, quoteStrings bool) []RadString {
	output := make([]RadString, len(v))
	for i, val := range v {
		output[i] = toStringQuoteStr(val, quoteStrings)
	}
	return output
}

func toStringQuoteStr(v RadValue, quoteStrings bool) RadString {
	switch coerced := v.Val.(type) {
	case RadString:
		return coerced
	default:
		str := ToPrintableQuoteStr(coerced, quoteStrings)
		return NewRadString(str)
	}
}
