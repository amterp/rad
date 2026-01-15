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
	filter         *RadFn
}

func newRadFieldMods(identifierNode *ts.Node) *radFieldMods {
	return &radFieldMods{
		identifierNode: identifierNode,
		colors:         make([]radColorMod, 0),
	}
}

func (i *Interpreter) runRadBlock(radBlockNode *ts.Node) {
	srcNode := rl.GetChild(radBlockNode, rl.F_SOURCE)
	radTypeNode := rl.GetChild(radBlockNode, rl.F_RAD_TYPE)
	typeStr := i.GetSrcForNode(radTypeNode)

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

	radStmtNodes := rl.GetChildren(radBlockNode, rl.F_STMT)
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
		identifierNodes := rl.GetChildren(node, rl.F_IDENTIFIER)
		for _, identifierNode := range identifierNodes {
			r.fields = append(r.fields, &identifierNode)
		}
	case rl.K_RAD_SORT_STMT:
		if r.generalSort != nil || len(r.colWiseSorting) > 0 {
			r.i.errorf(node, "Only one sort statement allowed per rad block")
		}

		specifierNodes := rl.GetChildren(node, rl.F_SPECIFIER)
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
		firstNode := rl.GetChild(node, rl.F_FIRST) // we can assume this non-nil, otherwise this node wouldn't exist
		secondNode := rl.GetChild(node, rl.F_SECOND)

		if secondNode == nil {
			firstNodeSrc := r.i.GetSrcForNode(firstNode)
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
		identifierNodes := rl.GetChildren(node, rl.F_IDENTIFIER)
		stmtNodes := rl.GetChildren(node, rl.F_MOD_STMT)
		var fields []radField
		for _, identifierNode := range identifierNodes {
			identifierStr := r.i.GetSrcForNode(&identifierNode)
			fields = append(fields, radField{
				node: &identifierNode,
				name: identifierStr,
			})
		}
		for _, stmtNode := range stmtNodes {
			switch stmtNode.Kind() {
			case rl.K_RAD_FIELD_MOD_COLOR:
				// todo could I replace this syntax with a 'map' lambda operation?
				clrExprNode := rl.GetChild(&stmtNode, rl.F_COLOR)
				clrStr := r.i.eval(clrExprNode).Val.RequireStr(r.i, clrExprNode)
				clr := AttrFromString(r.i, clrExprNode, clrStr.Plain())
				regexExprNode := rl.GetChild(&stmtNode, rl.F_REGEX)
				regexStr := r.i.eval(regexExprNode).Val.RequireStr(r.i, regexExprNode)
				regex, err := regexp.Compile(regexStr.Plain())
				if err != nil {
					r.i.errorf(regexExprNode, fmt.Sprintf("Invalid regex pattern: %s", err))
				}
				for _, field := range fields {
					mods := r.loadFieldMods(field)
					mods.colors = append(mods.colors, radColorMod{color: clr.ToTblColor(), regex: regex})
				}
			case rl.K_RAD_FIELD_MOD_MAP:
				lambdaNode := rl.GetChild(&stmtNode, rl.F_LAMBDA)
				lambda := r.resolveLambdaForModifier(lambdaNode, "map")

				for _, field := range fields {
					mods := r.loadFieldMods(field)
					mods.lambda = &lambda
				}
			case rl.K_RAD_FIELD_MOD_FILTER:
				lambdaNode := rl.GetChild(&stmtNode, rl.F_LAMBDA)
				lambda := r.resolveLambdaForModifier(lambdaNode, "filter")

				for _, field := range fields {
					mods := r.loadFieldMods(field)
					mods.filter = &lambda
				}
			}
		}
	case rl.K_RAD_IF_STMT:
		altNodes := rl.GetChildren(node, rl.F_ALT)
		for _, altNode := range altNodes {
			condNode := rl.GetChild(&altNode, rl.F_CONDITION)

			shouldExecute := true
			if condNode != nil {
				condResult := r.i.eval(condNode).Val.TruthyFalsy()
				shouldExecute = condResult
			}

			if shouldExecute {
				stmtNodes := rl.GetChildren(&altNode, rl.F_STMT)
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
		name := r.i.GetSrcForNode(fieldIdentifierNode)
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

	// Execution order: filter → sort → map
	//
	// Rationale:
	//   - Filter first: Sort only the rows we'll display (performance)
	//   - Map last: Display transformation shouldn't affect filtering/sorting logic
	//   - Sort middle: Sort filtered results in their original form
	//
	// For rad/request blocks: permanently affect field arrays for the rad block and beyond
	// For display blocks: temporarily apply for rendering, then restore original values
	if r.blockType == RadBlock || r.blockType == RequestBlock {
		indicesToKeep := r.applyFilters(radFields)
		r.filterColumns(radFields, indicesToKeep)
		applySorting(r.i, radFields, r.generalSort, r.colWiseSorting)
		r.applyMaps(radFields)
	} else {
		// Display block: save original values, apply filter/sort/map temporarily
		savedValues := make(map[string]*RadList)
		for _, field := range radFields {
			fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
			column := fieldVals.RequireList(r.i, field.node)
			// Save a copy of the original values
			savedValues[field.name] = &RadList{Values: append([]RadValue{}, column.Values...)}
		}

		// Temporarily mutate: filter → sort → map
		indicesToKeep := r.applyFilters(radFields)
		r.filterColumns(radFields, indicesToKeep)
		applySorting(r.i, radFields, r.generalSort, r.colWiseSorting)
		r.applyMaps(radFields)

		// Restore original values after rendering (deferred)
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
			r.i.errorf(field.node, "Values for field %q not found in environment", field.name)
		}
		columnValues := fieldVals.RequireList(r.i, field.node)
		longestColumnLen = com.IntMax(longestColumnLen, columnValues.LenInt())
		return toStringArrayQuoteStr(columnValues.Values, false), true
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

// applyFilters evaluates filter predicates across all fields and returns indices
// of rows that pass ALL filters (AND logic).
//
// Return value semantics:
//   - nil: No filters present, keep all rows
//   - []int64{}: Filters present but all rows filtered out
//   - []int64{...}: Indices of rows that passed all filters
func (r *radInvocation) applyFilters(radFields []radField) []int64 {
	// Early exit if no filters
	hasFilters := false
	for _, mods := range r.colToMods {
		if mods.filter != nil {
			hasFilters = true
			break
		}
	}
	if !hasFilters {
		return nil // nil = keep all rows
	}

	// Get row count from first field
	if len(radFields) == 0 {
		return nil
	}
	firstFieldVals := r.i.env.GetVarElseBug(r.i, radFields[0].node, radFields[0].name)
	firstColumn := firstFieldVals.RequireList(r.i, radFields[0].node)
	rowCount := firstColumn.LenInt()

	if rowCount == 0 {
		return nil
	}

	// Track which rows pass (AND logic)
	keepRow := make([]bool, rowCount)
	for i := range keepRow {
		keepRow[i] = true
	}

	// Evaluate each field's filter
	for _, field := range radFields {
		mods, hasMods := r.colToMods[field.name]
		if !hasMods || mods.filter == nil {
			continue
		}

		fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
		column := fieldVals.RequireList(r.i, field.node)

		// Check if filter lambda expects context (2+ parameters)
		wantsContext := lambdaWantsContext(mods.filter)

		// Copy original values for context.src to ensure it's an immutable snapshot
		var srcList *RadList
		if wantsContext {
			originalValues := make([]RadValue, len(column.Values))
			copy(originalValues, column.Values)
			srcList = &RadList{Values: originalValues}
		}

		for rowIdx := 0; rowIdx < column.LenInt(); rowIdx++ {
			if !keepRow[rowIdx] {
				continue // Already filtered out
			}

			val := column.Values[rowIdx]
			var args []PosArg
			args = append(args, NewPosArg(mods.filter.ReprNode, val))

			// If lambda expects 2+ params, pass context object as second arg
			if wantsContext {
				ctx := newRadBlockContext(r.i, mods.filter.ReprNode, int64(rowIdx), srcList, field.name)
				args = append(args, NewPosArg(mods.filter.ReprNode, ctx))
			}

			filterResult := mods.filter.Execute(
				NewFnInvocation(
					r.i,
					mods.filter.ReprNode,
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

	// Build list of kept indices
	keptIndices := make([]int64, 0) // Initialize as empty slice, not nil
	for i, keep := range keepRow {
		if keep {
			keptIndices = append(keptIndices, int64(i))
		}
	}

	return keptIndices
}

// filterColumns applies row filtering by removing rows from all field columns.
// Uses indicesToKeep from applyFilters() to determine which rows to keep.
//
// Parameters:
//   - radFields: All fields in the rad block
//   - indicesToKeep: Indices of rows to keep (nil means keep all)
//
// Side effects:
//   - Mutates column.Values for each field in the environment
//   - For rad/request blocks: permanent mutation
//   - For display blocks: caller must save/restore values
func (r *radInvocation) filterColumns(radFields []radField, indicesToKeep []int64) {
	if indicesToKeep == nil {
		return
	}

	if len(radFields) > 0 {
		// ensure all columns are equal length
		expectedLen := -1
		for _, field := range radFields {
			fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
			column := fieldVals.RequireList(r.i, field.node)
			actualLen := column.LenInt()

			if expectedLen == -1 {
				expectedLen = actualLen
			} else if actualLen != expectedLen {
				r.i.errorf(field.node,
					"Bug! Field %q has %d rows but expected %d. All fields must have identical row counts.",
					field.name, actualLen, expectedLen)
			}
		}

		// ensure no indices to keep are out of bounds
		for _, idx := range indicesToKeep {
			if idx < 0 || idx >= int64(expectedLen) {
				r.i.errorf(radFields[0].node,
					"Bug! Filter index %d is out of bounds for %d rows", idx, expectedLen)
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

// applyMaps applies map transformations to field columns, mutating them permanently
func (r *radInvocation) applyMaps(radFields []radField) {
	for _, field := range radFields {
		mods, hasMods := r.colToMods[field.name]
		if !hasMods || mods.lambda == nil {
			continue
		}

		fieldVals := r.i.env.GetVarElseBug(r.i, field.node, field.name)
		column := fieldVals.RequireList(r.i, field.node)

		// Check if lambda expects context (2+ parameters)
		wantsContext := lambdaWantsContext(mods.lambda)

		// Copy original values for context.src to avoid circular reference
		// (since we mutate column.Values below, ctx.src would otherwise point to mutated values)
		var srcList *RadList
		if wantsContext {
			originalValues := make([]RadValue, len(column.Values))
			copy(originalValues, column.Values)
			srcList = &RadList{Values: originalValues}
		}

		newValues := make([]RadValue, len(column.Values))
		for i, val := range column.Values {
			var args []PosArg
			args = append(args, NewPosArg(mods.lambda.ReprNode, val))

			// If lambda expects 2+ params, pass context object as second arg
			if wantsContext {
				ctx := newRadBlockContext(r.i, mods.lambda.ReprNode, int64(i), srcList, field.name)
				args = append(args, NewPosArg(mods.lambda.ReprNode, ctx))
			}

			mapped := mods.lambda.Execute(
				NewFnInvocation(
					r.i,
					mods.lambda.ReprNode,
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

// lambdaWantsContext returns true if the lambda expects 2+ parameters,
// indicating it wants the context object as the second argument.
func lambdaWantsContext(fn *RadFn) bool {
	return fn != nil && fn.ParamCount() >= 2
}

// newRadBlockContext creates a context object for rad block lambdas.
// Contains: idx (int), src (list), field (string)
func newRadBlockContext(i *Interpreter, node *ts.Node, idx int64, src *RadList, fieldName string) RadValue {
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
			r.i.errorf(r.srcExprNode, "Display block source can only be a list or a map. Got %q", TypeAsString(val))
		}).Visit(src)
		return
	} else {
		r.i.errorf(r.srcExprNode, "Bug! Unknown rad block type %q", r.blockType)
		panic(UNREACHABLE)
	}
}

func (r *radInvocation) resolveLambdaForModifier(lambdaNode *ts.Node, modifierName string) RadFn {
	var lambda RadFn

	if lambdaNode.Kind() == rl.K_FN_LAMBDA {
		lambda = NewFn(r.i, lambdaNode)
	} else if lambdaNode.Kind() == rl.K_IDENTIFIER {
		identifier := r.i.GetSrcForNode(lambdaNode)
		val, ok := r.i.env.GetVar(identifier)
		if !ok {
			r.i.errorf(lambdaNode, "Undefined lambda %q", identifier)
		}
		lambda, ok = val.TryGetFn()
		if !ok {
			r.i.errorf(lambdaNode, "Expected function for %s modifier, got '%s'",
				modifierName, TypeAsString(val))
		}
	} else {
		r.i.errorf(lambdaNode, "Bug! Unknown lambda type %q for %s modifier",
			lambdaNode.Kind(), modifierName)
	}

	return lambda
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
