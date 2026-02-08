package core

import (
	"fmt"
	"regexp"
	"runtime/debug"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts/rl"

	tblwriter "github.com/amterp/go-tbl"
	"github.com/samber/lo"
	"github.com/scylladb/go-set/strset"
)

type radInvocation struct {
	i                *Interpreter
	radBlockNode     rl.Node
	srcExprNode      rl.Node
	blockType        RadBlockType
	fields           []radField
	fieldsToNotPrint *strset.Set
	// if no specific column specified for sorting
	generalSort *GeneralSort
	// if specific columns listed for sorting, mutually exclusive with generalSort
	// in-order of sorting priority
	colWiseSorting []ColumnSort
	colToMods      map[string]*radFieldMods
}

type radFieldMods struct {
	identifierNode rl.Node
	colors         []radColorMod
	lambda         *RadFn
	filter         *RadFn
}

func newRadFieldMods(identifierNode rl.Node) *radFieldMods {
	return &radFieldMods{
		identifierNode: identifierNode,
		colors:         make([]radColorMod, 0),
	}
}

func (i *Interpreter) runRadBlock(n *rl.RadBlock) {
	var blockType RadBlockType
	switch n.BlockType {
	case rl.KEYWORD_RAD:
		blockType = RadBlock
	case rl.KEYWORD_REQUEST:
		blockType = RequestBlock
	case rl.KEYWORD_DISPLAY:
		blockType = DisplayBlock
	default:
		i.emitErrorf(rl.ErrInternalBug, n, "Bug: Unknown rad block type %q", n.BlockType)
	}

	ri := radInvocation{
		i:                i,
		radBlockNode:     n,
		srcExprNode:      n.Source,
		blockType:        blockType,
		fields:           make([]radField, 0),
		fieldsToNotPrint: strset.New(),
		colWiseSorting:   make([]ColumnSort, 0),
		colToMods:        make(map[string]*radFieldMods),
	}

	for _, stmt := range n.Stmts {
		ri.evalRad(stmt)
	}

	ri.execute()
}

func (r *radInvocation) evalRad(node rl.Node) {
	if !IsTest {
		defer func() {
			if re := recover(); re != nil {
				r.i.emitErrorWithHint(rl.ErrInternalBug, node, "Bug: Panic'd here",
					fmt.Sprintf("%s\n%s", re, debug.Stack()))
			}
		}()
	}
	r.unsafeEvalRad(node)
}

func (r *radInvocation) unsafeEvalRad(node rl.Node) {
	switch n := node.(type) {
	case *rl.RadField:
		for _, idNode := range n.Identifiers {
			if ident, ok := idNode.(*rl.Identifier); ok {
				r.fields = append(r.fields, radField{
					node: idNode,
					name: ident.Name,
				})
			}
		}
	case *rl.RadSort:
		if r.generalSort != nil || len(r.colWiseSorting) > 0 {
			r.i.emitError(rl.ErrUnsupportedOperation, node, "Only one sort statement allowed per rad block")
		}

		if len(n.Specifiers) == 0 {
			r.generalSort = &GeneralSort{
				Span: n.Span(),
				Dir:  Asc,
			}
		} else {
			for _, spec := range n.Specifiers {
				dir := lo.Ternary(spec.Ascending, Asc, Desc)

				if spec.Field == "" {
					// General sort (just asc/desc with no field)
					r.generalSort = &GeneralSort{
						Span: n.Span(),
						Dir:  dir,
					}
				} else {
					r.colWiseSorting = append(r.colWiseSorting, ColumnSort{
						ColName: spec.Field,
						Span:    n.Span(),
						Dir:     dir,
					})
				}
			}
		}
	case *rl.RadFieldMod:
		if n.ModType == "" {
			// Container level: Fields holds the target identifiers, Args holds the child modifiers
			var fields []radField
			for _, idNode := range n.Fields {
				if ident, ok := idNode.(*rl.Identifier); ok {
					fields = append(fields, radField{
						node: idNode,
						name: ident.Name,
					})
				}
			}
			for _, modNode := range n.Args {
				r.applyModifier(fields, modNode)
			}
		}
	case *rl.RadIf:
		for _, branch := range n.Branches {
			shouldExecute := true
			if branch.Condition != nil {
				shouldExecute = r.i.eval(branch.Condition).Val.TruthyFalsy()
			}

			if shouldExecute {
				for _, stmt := range branch.Body {
					r.evalRad(stmt)
				}
				break
			}
		}
	}
}

func (r *radInvocation) applyModifier(fields []radField, modNode rl.Node) {
	mod, ok := modNode.(*rl.RadFieldMod)
	if !ok {
		return
	}

	switch mod.ModType {
	case "color":
		if len(mod.Args) >= 2 {
			clrVal := r.i.eval(mod.Args[0]).Val.RequireStr(r.i, mod.Args[0])
			clr := AttrFromString(r.i, mod.Args[0], clrVal.Plain())
			regexVal := r.i.eval(mod.Args[1]).Val.RequireStr(r.i, mod.Args[1])
			regex, err := regexp.Compile(regexVal.Plain())
			if err != nil {
				r.i.emitErrorf(rl.ErrInvalidRegex, mod.Args[1], "Invalid regex pattern: %s", err)
			}
			for _, field := range fields {
				mods := r.loadFieldMods(field)
				mods.colors = append(mods.colors, radColorMod{color: clr.ToTblColor(), regex: regex})
			}
		}
	case "map":
		if len(mod.Args) >= 1 {
			lambda := r.resolveLambdaForModifier(mod.Args[0], "map")
			for _, field := range fields {
				mods := r.loadFieldMods(field)
				mods.lambda = &lambda
			}
		}
	case "filter":
		if len(mod.Args) >= 1 {
			lambda := r.resolveLambdaForModifier(mod.Args[0], "filter")
			for _, field := range fields {
				mods := r.loadFieldMods(field)
				mods.filter = &lambda
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
	node rl.Node
	name string
}

func (r *radInvocation) execute() {
	if len(r.fields) == 0 {
		r.i.emitError(rl.ErrInvalidSyntax, r.radBlockNode, "No fields specified in rad block")
	}

	radFields := r.fields

	// check all field mods are for fields that actually exist
	fieldNames := lo.Map(radFields, func(f radField, _ int) string { return f.name })
	for field, mods := range r.colToMods {
		if !lo.Contains(fieldNames, field) {
			r.i.emitErrorf(rl.ErrUndefinedVariable, mods.identifierNode, "Cannot modify undefined field %q", field)
		}
	}

	data, err := r.resolveData()
	if err != nil {
		r.i.emitErrorf(rl.ErrGenericRuntime, r.srcExprNode, "Error resolving data: %v", err)
	}

	if data != nil {
		jsonFields := lo.Map(radFields, func(field radField, _ int) JsonFieldVar {
			fieldVar, ok := r.i.env.GetJsonFieldVar(field.name)
			if !ok {
				r.i.emitErrorf(rl.ErrUndefinedVariable, field.node, "Undefined JSON field %q", field.name)
			}
			return *fieldVar
		})

		trie := CreateTrie(r.i, r.radBlockNode, jsonFields)
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

	// Execution order: filter -> sort -> map
	if r.blockType == RadBlock || r.blockType == RequestBlock {
		indicesToKeep := r.applyFilters(radFields)
		r.filterColumns(radFields, indicesToKeep)
		applySorting(r.i, radFields, r.generalSort, r.colWiseSorting)
		r.applyMaps(radFields)
	} else {
		// Display block: save/restore pattern
		savedValues := make(map[string]*RadList)
		for _, field := range radFields {
			fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
			column := fieldVals.RequireList(r.i, field.node)
			savedValues[field.name] = &RadList{Values: append([]RadValue{}, column.Values...)}
		}

		indicesToKeep := r.applyFilters(radFields)
		r.filterColumns(radFields, indicesToKeep)
		applySorting(r.i, radFields, r.generalSort, r.colWiseSorting)
		r.applyMaps(radFields)

		defer func() {
			for _, field := range radFields {
				if saved, ok := savedValues[field.name]; ok {
					r.i.env.SetVar(field.name, newRadValue(r.i, field.node, saved))
				}
			}
		}()
	}

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
			r.i.emitErrorf(rl.ErrUndefinedVariable, field.node, "Values for field %q not found in environment", field.name)
		}
		columnValues := fieldVals.RequireList(r.i, field.node)
		longestColumnLen = com.IntMax(longestColumnLen, columnValues.LenInt())
		return toStringArrayQuoteStr(columnValues.Values, false), true
	})

	tbl := NewTblWriter()
	tbl.SetHeader(headers)

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
	tbl.Render()
}

func (r *radInvocation) applyFilters(radFields []radField) []int64 {
	hasFilters := false
	for _, mods := range r.colToMods {
		if mods.filter != nil {
			hasFilters = true
			break
		}
	}
	if !hasFilters {
		return nil
	}

	if len(radFields) == 0 {
		return nil
	}
	firstFieldVals := r.i.env.GetVarElseBug(r.i, radFields[0].node, radFields[0].name)
	firstColumn := firstFieldVals.RequireList(r.i, radFields[0].node)
	rowCount := firstColumn.LenInt()

	if rowCount == 0 {
		return nil
	}

	keepRow := make([]bool, rowCount)
	for i := range keepRow {
		keepRow[i] = true
	}

	for _, field := range radFields {
		mods, hasMods := r.colToMods[field.name]
		if !hasMods || mods.filter == nil {
			continue
		}

		fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
		column := fieldVals.RequireList(r.i, field.node)

		wantsContext := lambdaWantsContext(mods.filter)

		var srcList *RadList
		if wantsContext {
			originalValues := make([]RadValue, len(column.Values))
			copy(originalValues, column.Values)
			srcList = &RadList{Values: originalValues}
		}

		for rowIdx := 0; rowIdx < column.LenInt(); rowIdx++ {
			if !keepRow[rowIdx] {
				continue
			}

			val := column.Values[rowIdx]
			reprSpan := mods.filter.ReprSpan
			var reprNode rl.Node
			if reprSpan != nil {
				reprNode = rl.NewLitNull(*reprSpan) // dummy node for span
			}

			var args []PosArg
			args = append(args, NewPosArg(reprNode, val))

			if wantsContext {
				ctx := newRadBlockContext(r.i, reprNode, int64(rowIdx), srcList, field.name)
				args = append(args, NewPosArg(reprNode, ctx))
			}

			filterResult := mods.filter.Execute(
				NewFnInvocation(
					r.i,
					reprNode,
					FUNC_FILTER,
					args,
					NO_NAMED_ARGS_INPUT,
					mods.filter.IsBuiltIn(),
				),
			)

			if !filterResult.TruthyFalsy() {
				keepRow[rowIdx] = false
			}
		}
	}

	keptIndices := make([]int64, 0)
	for i, keep := range keepRow {
		if keep {
			keptIndices = append(keptIndices, int64(i))
		}
	}

	return keptIndices
}

func (r *radInvocation) filterColumns(radFields []radField, indicesToKeep []int64) {
	if indicesToKeep == nil {
		return
	}

	if len(radFields) > 0 {
		expectedLen := -1
		for _, field := range radFields {
			fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
			column := fieldVals.RequireList(r.i, field.node)
			actualLen := column.LenInt()

			if expectedLen == -1 {
				expectedLen = actualLen
			} else if actualLen != expectedLen {
				r.i.emitErrorf(rl.ErrInternalBug, field.node,
					"Bug: Field %q has %d rows but expected %d. All fields must have identical row counts.",
					field.name, actualLen, expectedLen)
			}
		}

		for _, idx := range indicesToKeep {
			if idx < 0 || idx >= int64(expectedLen) {
				r.i.emitErrorf(rl.ErrInternalBug, radFields[0].node,
					"Bug: Filter index %d is out of bounds for %d rows", idx, expectedLen)
			}
		}
	}

	for _, field := range radFields {
		fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
		column := fieldVals.RequireList(r.i, field.node)

		newValues := make([]RadValue, len(indicesToKeep))
		for newIdx, oldIdx := range indicesToKeep {
			newValues[newIdx] = column.Values[oldIdx]
		}

		column.Values = newValues
		r.i.env.SetVar(field.name, newRadValue(r.i, field.node, column))
	}
}

func (r *radInvocation) applyMaps(radFields []radField) {
	for _, field := range radFields {
		mods, hasMods := r.colToMods[field.name]
		if !hasMods || mods.lambda == nil {
			continue
		}

		fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
		column := fieldVals.RequireList(r.i, field.node)

		wantsContext := lambdaWantsContext(mods.lambda)

		var srcList *RadList
		if wantsContext {
			originalValues := make([]RadValue, len(column.Values))
			copy(originalValues, column.Values)
			srcList = &RadList{Values: originalValues}
		}

		reprSpan := mods.lambda.ReprSpan
		var reprNode rl.Node
		if reprSpan != nil {
			reprNode = rl.NewLitNull(*reprSpan) // dummy node for span
		}

		newValues := make([]RadValue, len(column.Values))
		for i, val := range column.Values {
			var args []PosArg
			args = append(args, NewPosArg(reprNode, val))

			if wantsContext {
				ctx := newRadBlockContext(r.i, reprNode, int64(i), srcList, field.name)
				args = append(args, NewPosArg(reprNode, ctx))
			}

			mapped := mods.lambda.Execute(
				NewFnInvocation(
					r.i,
					reprNode,
					FUNC_MAP,
					args,
					NO_NAMED_ARGS_INPUT,
					mods.lambda.IsBuiltIn(),
				),
			)
			newValues[i] = mapped
		}

		column.Values = newValues
		r.i.env.SetVar(field.name, newRadValue(r.i, field.node, column))
	}
}

func lambdaWantsContext(fn *RadFn) bool {
	return fn != nil && fn.ParamCount() >= 2
}

func newRadBlockContext(i *Interpreter, node rl.Node, idx int64, src *RadList, fieldName string) RadValue {
	ctx := NewRadMap()
	ctx.Set(newRadValue(i, node, "idx"), newRadValue(i, node, idx))
	ctx.Set(newRadValue(i, node, "src"), newRadValue(i, node, src))
	ctx.Set(newRadValue(i, node, "field"), newRadValue(i, node, fieldName))
	return newRadValue(i, node, ctx)
}

func (r *radInvocation) resolveData() (data interface{}, err error) {
	if r.srcExprNode == nil {
		return nil, nil
	}

	src := r.i.eval(r.srcExprNode).Val

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
			r.i.emitErrorf(rl.ErrTypeMismatch, r.srcExprNode, "Display block source can only be a list or a map. Got %q", TypeAsString(val))
		}).Visit(src)
		return
	} else {
		r.i.emitErrorf(rl.ErrInternalBug, r.srcExprNode, "Bug: Unknown rad block type %q", r.blockType)
		panic(UNREACHABLE)
	}
}

func (r *radInvocation) resolveLambdaForModifier(lambdaNode rl.Node, modifierName string) RadFn {
	var lambda RadFn

	switch n := lambdaNode.(type) {
	case *rl.Lambda:
		lambda = NewFnFromAST(r.i, n.Typing, n.Body, n.IsBlock, &n.DefSpan)
	case *rl.Identifier:
		val, ok := r.i.env.GetVar(n.Name)
		if !ok {
			r.i.emitErrorf(rl.ErrUndefinedVariable, lambdaNode, "Undefined lambda %q", n.Name)
		}
		lambda, ok = val.TryGetFn()
		if !ok {
			r.i.emitErrorf(rl.ErrTypeMismatch, lambdaNode, "Expected function for %s modifier, got '%s'",
				modifierName, TypeAsString(val))
		}
	default:
		r.i.emitErrorf(rl.ErrInternalBug, lambdaNode, "Bug: Unknown lambda type %T for %s modifier",
			lambdaNode, modifierName)
	}

	return lambda
}

func applySorting(i *Interpreter, fields []radField, generalSort *GeneralSort, colWiseSort []ColumnSort) {
	if generalSort != nil {
		if len(colWiseSort) > 0 {
			i.emitErrorf(rl.ErrInternalBug, nil, "Bug: General and column-wise sort expected to be mutually exclusive")
		}
		for _, field := range fields {
			colWiseSort = append(colWiseSort, ColumnSort{
				ColName: field.name,
				Span:    generalSort.Span,
				Dir:     generalSort.Dir,
			})
		}
	}

	sortColumns(i, fields, colWiseSort)
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
