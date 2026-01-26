package core

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts"

	"github.com/amterp/rad/rts/rl"

	"github.com/samber/lo"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type Delimiter struct {
	Open string
}

type CtrlKind int

const (
	CtrlNormal CtrlKind = iota
	CtrlBreak
	CtrlContinue
	CtrlReturn
	CtrlYield
)

type EvalResult struct {
	Val  RadValue
	Ctrl CtrlKind
}

func NewEvalResult(val RadValue, ctrl CtrlKind) EvalResult {
	return EvalResult{
		Val:  val,
		Ctrl: ctrl,
	}
}

func NormalVal(val RadValue) EvalResult {
	return NewEvalResult(val, CtrlNormal)
}

func ReturnVal(val RadValue) EvalResult {
	return NewEvalResult(val, CtrlReturn)
}

func YieldVal(val RadValue) EvalResult {
	return NewEvalResult(val, CtrlYield)
}

var VoidNormal = NewEvalResult(VOID_SENTINEL, CtrlNormal)
var VoidBreak = NewEvalResult(VOID_SENTINEL, CtrlBreak)
var VoidContinue = NewEvalResult(VOID_SENTINEL, CtrlContinue)

type InterpreterInput struct {
	Src            string
	Tree           *rts.RadTree
	ScriptName     string
	InvokedCommand *ScriptCommand
}

type Interpreter struct {
	sd             *ScriptData
	invokedCommand *ScriptCommand
	env            *Env
	deferBlocks    []*DeferBlock
	tmpSrc         *string

	forWhileLoopLevel int
	// Used to track current delimiter, currently for correct delimiter escaping handling
	delimiterStack *com.Stack[Delimiter]
}

func NewInterpreter(input InterpreterInput) *Interpreter {
	// Construct ScriptData from input
	scriptData := &ScriptData{
		ScriptName: input.ScriptName,
		Tree:       input.Tree,
		Src:        input.Src,
	}

	i := &Interpreter{
		sd:             scriptData,
		invokedCommand: input.InvokedCommand,
		delimiterStack: com.NewStack[Delimiter](),
	}
	i.env = NewEnv(i)
	return i
}

func (i *Interpreter) InitBuiltIns() {
	for name, fn := range FunctionsByName {
		fnVal := newRadValueFn(NewBuiltIn(fn))
		i.env.SetVar(name, fnVal)
	}
}

func (i *Interpreter) InitArgs(args []RadArg) {
	env := i.env

	for _, arg := range args {
		// Special handling for variadic arguments with list defaults
		if arg.IsVariadic() {
			var hasListDefaults bool
			var defaultValue interface{}

			if stringListArg, ok := arg.(*StringListRadArg); ok && stringListArg.scriptArg != nil && stringListArg.scriptArg.DefaultStringList != nil {
				hasListDefaults = true
				defaultValue = NewRadListFromGeneric(i, arg.GetNode(), *stringListArg.scriptArg.DefaultStringList)
			} else if intListArg, ok := arg.(*IntListRadArg); ok && intListArg.scriptArg != nil && intListArg.scriptArg.DefaultIntList != nil {
				hasListDefaults = true
				defaultValue = NewRadListFromGeneric(i, arg.GetNode(), *intListArg.scriptArg.DefaultIntList)
			} else if floatListArg, ok := arg.(*FloatListRadArg); ok && floatListArg.scriptArg != nil && floatListArg.scriptArg.DefaultFloatList != nil {
				hasListDefaults = true
				defaultValue = NewRadListFromGeneric(i, arg.GetNode(), *floatListArg.scriptArg.DefaultFloatList)
			} else if boolListArg, ok := arg.(*BoolListRadArg); ok && boolListArg.scriptArg != nil && boolListArg.scriptArg.DefaultBoolList != nil {
				hasListDefaults = true
				defaultValue = NewRadListFromGeneric(i, arg.GetNode(), *boolListArg.scriptArg.DefaultBoolList)
			}

			if hasListDefaults {
				if !arg.Configured() {
					// Use defaults only if user didn't provide values
					env.SetVar(arg.GetIdentifier(), newRadValue(i, arg.GetNode(), defaultValue))
				} else {
					// User provided values, use them normally
					goto normalProcessing
				}
				continue
			}
		}

		if !arg.IsDefined() {
			if arg.IsVariadic() {
				// Variadic args should be empty lists when undefined, not null
				env.SetVar(arg.GetIdentifier(), newRadValue(i, arg.GetNode(), NewRadList()))
			} else {
				env.SetVar(arg.GetIdentifier(), RAD_NULL_VAL)
			}
			continue
		}

	normalProcessing:
		switch coerced := arg.(type) {
		case *BoolRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), coerced.Value))
		case *BoolListRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), NewRadListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *StringRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), coerced.Value))
		case *StringListRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), NewRadListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *IntRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), coerced.Value))
		case *IntListRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), NewRadListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *FloatRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), coerced.Value))
		case *FloatListRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), NewRadListFromGeneric(i, arg.GetNode(), coerced.Value)))
		default:
			i.errorf(arg.GetNode(), "Unsupported arg type, cannot init: %T", arg)
		}
	}
}

func (i *Interpreter) Run() {
	root := i.sd.Tree.Root()

	// PHASE 1: Execute top-level code (always)
	res := i.safelyExecuteTopLevel(root)
	if res.Ctrl != CtrlNormal {
		i.errorf(root, "Bug? Unexpected control flow: %s", res.Ctrl)
	}

	// PHASE 2: Execute command callback (if command was invoked)
	if i.invokedCommand != nil {
		i.safelyExecuteCommandCallback(i.invokedCommand)
	}
}

func (i *Interpreter) safelyExecuteTopLevel(root *ts.Node) EvalResult {
	defer func() {
		i.handlePanicRecovery(recover(), root)
	}()

	children := root.Children(root.Walk())

	// First pass: define custom named functions (function hoisting)
	for _, child := range children {
		if child.Kind() == rl.K_FN_NAMED {
			i.defineCustomNamedFunction(child)
		}
	}

	// Second pass: evaluate all children except command blocks
	var lastResult EvalResult
	for _, child := range children {
		kind := child.Kind()
		// Skip command blocks and other non-executable nodes
		if kind == rl.K_CMD_BLOCK || kind == rl.K_COMMENT || kind == rl.K_SHEBANG ||
			kind == rl.K_FILE_HEADER || kind == rl.K_ARG_BLOCK {
			continue
		}
		lastResult = i.eval(&child)
	}

	return lastResult
}

func (i *Interpreter) safelyExecuteCommandCallback(cmd *ScriptCommand) {
	defer func() {
		// Use nil node since we don't have a specific node for the command invocation
		i.handlePanicRecovery(recover(), nil, cmd.Name)
	}()

	switch cmd.CallbackType {
	case rts.CallbackIdentifier:
		// Look up the function by name and call it
		funcName := *cmd.CallbackName
		val, exist := i.env.GetVar(funcName)
		if !exist {
			// Find the root node for error reporting
			root := i.sd.Tree.Root()
			i.errorf(root, "Cannot invoke unknown function '%s' for command '%s'", funcName, cmd.Name)
		}

		fn, ok := val.TryGetFn()
		if !ok {
			root := i.sd.Tree.Root()
			i.errorf(root, "Cannot invoke '%s' as a function for command '%s': it is a %s",
				funcName, cmd.Name, val.Type().AsString())
		}

		// Execute the function with no arguments
		_ = fn.Execute(NewFnInvocation(i, nil, funcName, []PosArg{}, make(map[string]namedArg), fn.IsBuiltIn()))

	case rts.CallbackLambda:
		// Create a function from the lambda node
		fn := NewFn(i, cmd.CallbackLambda)

		// Execute the lambda with no arguments
		_ = fn.Execute(NewFnInvocation(i, cmd.CallbackLambda, "<lambda>", []PosArg{}, make(map[string]namedArg), false))

	default:
		panic(fmt.Sprintf("Bug! Unknown callback type %d for command: %s", cmd.CallbackType, cmd.Name))
	}
}

// EvaluateStatement evaluates a single statement string and returns the result
// This is designed for REPL use where individual statements are evaluated
// against a persistent interpreter environment
func (i *Interpreter) EvaluateStatement(input string) (RadValue, error) {
	// Parse the input statement
	parser, err := rts.NewRadParser()
	if err != nil {
		return RAD_NULL_VAL, fmt.Errorf("failed to create parser: %w", err)
	}
	defer parser.Close()

	tree := parser.Parse(input)
	// Check for parse errors using FindInvalidNodes
	if invalidNodes := tree.FindInvalidNodes(); len(invalidNodes) > 0 {
		return RAD_NULL_VAL, fmt.Errorf("parse error in statement: %s", input)
	}

	// Update the interpreter's ScriptData to point to this new tree and source
	// This ensures that GetSrcForNode and other methods work correctly
	originalScriptData := i.sd
	i.sd = &ScriptData{
		ScriptName:        "<repl>",
		Tree:              tree,
		Src:               input,
		DisableGlobalOpts: true,
		DisableArgsBlock:  true,
	}

	// Restore original ScriptData after evaluation and handle any panics
	var evalErr error
	defer func() {
		i.sd = originalScriptData
		if r := recover(); r != nil {
			// Convert panic to error instead of crashing REPL
			if radPanic, ok := r.(*RadPanic); ok {
				evalErr = fmt.Errorf("%v", radPanic.Err().Msg().Plain())
			} else {
				evalErr = fmt.Errorf("Runtime error: %v", r)
			}
		}
	}()

	node := tree.Root()
	children := node.Children(node.Walk())

	// REPL: unwrap the source_file container and evaluate the actual statement/expression
	// This allows both expressions ("2 + 3") and statements ("x = 5") to work correctly
	if len(children) == 1 {
		// Evaluate the child directly, bypassing the source_file wrapper
		result := i.safelyEvaluate(&children[0])
		if result.Ctrl != CtrlNormal {
			return RAD_NULL_VAL, fmt.Errorf("unexpected control flow: %s", result.Ctrl)
		}
		return result.Val, nil
	}

	// Fallback: evaluate normally (shouldn't happen for single statements)
	res := i.safelyEvaluate(node)
	// Check for any evaluation errors from defer
	if evalErr != nil {
		return RAD_NULL_VAL, evalErr
	}

	if res.Ctrl != CtrlNormal {
		return RAD_NULL_VAL, fmt.Errorf("unexpected control flow: %s", res.Ctrl)
	}
	return res.Val, nil
}

func (i *Interpreter) safelyEvaluate(node *ts.Node) EvalResult {
	defer func() {
		i.handlePanicRecovery(recover(), node)
	}()
	return i.eval(node)
}

func (i *Interpreter) eval(node *ts.Node) (out EvalResult) {
	out = VoidNormal

	switch node.Kind() {
	// no-ops
	case rl.K_SOURCE_FILE:
		children := node.Children(node.Walk())
		// first define custom named functions
		for _, child := range children {
			if child.Kind() == rl.K_FN_NAMED {
				i.defineCustomNamedFunction(child)
			}
		}
		for _, child := range children {
			i.eval(&child)
		}
	case rl.K_COMMENT, rl.K_SHEBANG, rl.K_FILE_HEADER, rl.K_ARG_BLOCK, rl.K_CMD_BLOCK:
		// no-op
	case rl.K_ERROR:
		i.errorf(node, "Bug! Error pre-check should've prevented running into this node")
	case rl.K_ASSIGN:
		catchBlockNode := rl.GetChild(node, rl.F_CATCH)

		out = i.withCatch(catchBlockNode, func(rp *RadPanic) EvalResult {
			// Assign error to first variable, null to rest
			leftNodes := i.getAssignLeftNodes(node)
			if len(leftNodes) > 0 {
				i.doVarPathAssign(&leftNodes[0], rp.ErrV, false)
				for j := 1; j < len(leftNodes); j++ {
					i.doVarPathAssign(&leftNodes[j], RAD_NULL_VAL, false)
				}
			}

			// Run catch block and propagate control flow
			stmtNodes := rl.GetChildren(catchBlockNode, rl.F_STMT)
			res := i.runBlock(stmtNodes)
			if res.Ctrl != CtrlNormal {
				return res // Propagate return/break/continue/yield
			}
			return VoidNormal
		}, func() EvalResult {
			// Normal assignment execution
			rightNodes := rl.GetChildren(node, rl.F_RIGHT)
			leftNodes := rl.GetChildren(node, rl.F_LEFT)
			if len(leftNodes) > 0 {
				return i.assignRightsToLefts(leftNodes, rightNodes, false)
			} else {
				leftsNodes := rl.GetChildren(node, rl.F_LEFTS)
				return i.assignRightsToLefts(leftsNodes, rightNodes, true)
			}
		})
	case rl.K_COMPOUND_ASSIGN:
		leftVarPathNode := rl.GetChild(node, rl.F_LEFT)
		rightNode := rl.GetChild(node, rl.F_RIGHT)
		opNode := rl.GetChild(node, rl.F_OP)
		newValue := i.executeCompoundOp(node, leftVarPathNode, rightNode, opNode)
		i.doVarPathAssign(leftVarPathNode, newValue, true)
	case rl.K_EXPR:
		out = i.eval(rl.GetChild(node, rl.F_DELEGATE))
	case rl.K_EXPR_STMT:
		exprNode := rl.GetChild(node, rl.F_EXPR)
		catchBlockNode := rl.GetChild(node, rl.F_CATCH)

		out = i.withCatch(catchBlockNode, func(rp *RadPanic) EvalResult {
			// Run catch block and propagate control flow
			stmtNodes := rl.GetChildren(catchBlockNode, rl.F_STMT)
			res := i.runBlock(stmtNodes)
			if res.Ctrl != CtrlNormal {
				return res // Propagate return/break/continue/yield
			}
			return VoidNormal // Statement with catch returns void, not error value
		}, func() EvalResult {
			return i.eval(exprNode)
		})
	case rl.K_PASS:
		// no-op
	case rl.K_RETURN_STMT:
		return ReturnVal(i.evalRights(node))
	case rl.K_YIELD_STMT:
		return YieldVal(i.evalRights(node))
	case rl.K_BREAK_STMT:
		if i.forWhileLoopLevel > 0 {
			return VoidBreak
		}
		i.errorf(node, "Cannot 'break' outside of a for loop")
	case rl.K_CONTINUE_STMT:
		if i.forWhileLoopLevel > 0 {
			return VoidContinue
		}
		i.errorf(node, "Cannot 'continue' outside of a for loop")
	case rl.K_FOR_LOOP:
		i.forWhileLoopLevel++
		defer func() {
			i.forWhileLoopLevel--
		}()
		stmts := rl.GetChildren(node, rl.F_STMT)
		return i.executeForLoop(node, func() EvalResult { return i.runBlock(stmts) })
	case rl.K_WHILE_LOOP:
		i.forWhileLoopLevel++
		defer func() {
			i.forWhileLoopLevel--
		}()
		condNode := rl.GetChild(node, rl.F_CONDITION)
		stmtNodes := rl.GetChildren(node, rl.F_STMT)
		for {
			condValue := true
			if condNode != nil {
				condValue = i.eval(condNode).Val.TruthyFalsy()
			}
			if condValue {
				res := i.runBlock(stmtNodes)
				switch res.Ctrl {
				case CtrlBreak:
					return VoidNormal
				case CtrlReturn, CtrlYield:
					return res
				}
			} else {
				break
			}
		}
	case rl.K_IF_STMT:
		altNodes := rl.GetChildren(node, rl.F_ALT)
		for _, altNode := range altNodes {
			condNode := rl.GetChild(&altNode, rl.F_CONDITION)

			shouldExecute := true
			if condNode != nil {
				condResult := i.eval(condNode).Val.TruthyFalsy()
				shouldExecute = condResult
			}

			if shouldExecute {
				stmtNodes := rl.GetChildren(&altNode, rl.F_STMT)
				return i.runBlock(stmtNodes)
			}
		}
	case rl.K_SWITCH_STMT:
		discriminantNode := rl.GetChild(node, rl.F_DISCRIMINANT)
		caseNodes := rl.GetChildren(node, rl.F_CASE)
		defaultNode := rl.GetChild(node, rl.F_DEFAULT)

		discriminantVal := i.eval(discriminantNode).Val

		matchedCaseNodes := make([]ts.Node, 0)
		for _, caseNode := range caseNodes {
			caseKeyNodes := rl.GetChildren(&caseNode, rl.F_CASE_KEY)
			for _, caseKeyNode := range caseKeyNodes {
				caseKey := i.eval(&caseKeyNode).Val
				if caseKey.Equals(discriminantVal) {
					matchedCaseNodes = append(matchedCaseNodes, caseNode)
					break
				}
			}
		}

		if len(matchedCaseNodes) == 0 {
			if defaultNode != nil {
				caseValueAltNode := rl.GetChild(defaultNode, rl.F_ALT)
				return i.executeSwitchCase(caseValueAltNode)
			}
			i.errorf(discriminantNode, "No matching case found for switch")
		}

		if len(matchedCaseNodes) > 1 {
			// todo fancier error msg: should point at the cases that matched, example:
			// 13 | case 1, "one": blah blah
			//           ^  ^^^^^ MATCHED        << in red
			// 14 | case 2, "two": blah blah
			// 15 | case 3, val: blah blah
			//              ^^^ MATCHED          << in red
			// 16 | case 4, "four":  blah blah
			i.errorf(discriminantNode, "Multiple matching cases found for switch")
		}

		matchedCaseNode := matchedCaseNodes[0]
		caseValueAltNode := rl.GetChild(&matchedCaseNode, rl.F_ALT)
		return i.executeSwitchCase(caseValueAltNode)
	case rl.K_FN_NAMED:
		// do not redefine top-level functions as they're already defined
		if node.Parent().Kind() != rl.K_SOURCE_FILE {
			i.defineCustomNamedFunction(*node)
		}
	case rl.K_DEFER_BLOCK:
		keywordNode := rl.GetChild(node, rl.F_KEYWORD)
		stmtNodes := rl.GetChildren(node, rl.F_STMT)
		i.deferBlocks = append(i.deferBlocks, NewDeferBlock(i, keywordNode, stmtNodes))
	case rl.K_SHELL_STMT:
		out = i.executeShellStmt(node)
	case rl.K_DEL_STMT:
		rightVarPathNodes := rl.GetChildren(node, rl.F_RIGHT)
		for _, rightVarPathNode := range rightVarPathNodes {
			i.doVarPathAssign(&rightVarPathNode, VOID_SENTINEL, true)
		}
	case rl.K_RAD_BLOCK:
		i.runRadBlock(node)
	case rl.K_INCR_DECR:
		leftVarPathNode := rl.GetChild(node, rl.F_LEFT)
		opNode := rl.GetChild(node, rl.F_OP)
		newValue := i.executeUnaryOp(node, leftVarPathNode, opNode)
		i.doVarPathAssign(leftVarPathNode, newValue, true)
	case rl.K_PRIMARY_EXPR, rl.K_LITERAL:
		return i.eval(i.getOnlyChild(node))
	case rl.K_PARENTHESIZED_EXPR:
		return i.eval(rl.GetChild(node, rl.F_EXPR))
	case rl.K_UNARY_EXPR:
		delegateNode := rl.GetChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.eval(delegateNode)
		}

		opNode := rl.GetChild(node, rl.F_OP)
		argNode := rl.GetChild(node, rl.F_ARG)
		return NormalVal(newRadValues(i, node, i.executeUnaryOp(node, argNode, opNode)))
	case rl.K_OR_EXPR, rl.K_AND_EXPR, rl.K_COMPARE_EXPR, rl.K_ADD_EXPR, rl.K_MULT_EXPR:
		delegateNode := rl.GetChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.eval(delegateNode)
		}

		left := rl.GetChild(node, rl.F_LEFT)
		op := rl.GetChild(node, rl.F_OP)
		right := rl.GetChild(node, rl.F_RIGHT)
		return NormalVal(newRadValues(i, node, i.executeBinary(node, left, right, op)))
	case rl.K_FALLBACK_EXPR:
		delegateNode := rl.GetChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.eval(delegateNode)
		}

		leftNode := rl.GetChild(node, rl.F_LEFT)
		rightNode := rl.GetChild(node, rl.F_RIGHT)

		// Evaluate left with panic catching
		var leftResult EvalResult
		panicked := false

		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(*RadPanic); ok {
						panicked = true
					} else {
						panic(r) // Re-panic non-RadPanic errors
					}
				}
			}()
			leftResult = i.eval(leftNode)
		}()

		if panicked {
			// Left side errored, evaluate and return right side
			return i.eval(rightNode)
		}

		return leftResult

	// LEAF NODES
	case rl.K_IDENTIFIER:
		identifier := i.GetSrcForNode(node)
		if identifier == "_" {
			i.errorf(node, "Cannot use '_' as a value")
		}
		val, ok := i.env.GetVar(identifier)
		if !ok {
			i.errorf(node, "Undefined variable: %s", identifier)
		}
		return NormalVal(newRadValues(i, node, val))
	case rl.K_VAR_PATH:
		rootNode := rl.GetChild(node, rl.F_ROOT)
		indexingNodes := rl.GetChildren(node, rl.F_INDEXING)
		val := i.eval(rootNode).Val
		for _, indexNode := range indexingNodes {
			val = i.evaluateIndexing(rootNode, indexNode, val)
		}
		return NormalVal(newRadValues(i, node, val))
	case rl.K_INDEXED_EXPR:
		rootNode := rl.GetChild(node, rl.F_ROOT)
		indexingNodes := rl.GetChildren(node, rl.F_INDEXING)
		if len(indexingNodes) > 0 {
			val := i.eval(rootNode).Val
			for _, index := range indexingNodes {
				val = i.evaluateIndexing(rootNode, index, val)
			}
			return NormalVal(newRadValues(i, node, val))
		} else {
			return i.eval(rootNode)
		}
	case rl.K_INT:
		asStr := i.GetSrcForNode(node)
		asInt, _ := rts.ParseInt(asStr) // todo unhandled err
		return NormalVal(newRadValues(i, node, asInt))
	case rl.K_FLOAT:
		asStr := i.GetSrcForNode(node)
		asFloat, _ := rts.ParseFloat(asStr) // todo unhandled err
		return NormalVal(newRadValues(i, node, asFloat))
	case rl.K_SCIENTIFIC_NUMBER:
		asStr := i.GetSrcForNode(node)
		asFloat, _ := rts.ParseFloat(asStr) // todo unhandled err
		// Evaluate as int if it's a whole number, float otherwise
		// RadChecker validates that int-typed params only use whole numbers
		if asFloat == float64(int64(asFloat)) {
			return NormalVal(newRadValues(i, node, int64(asFloat)))
		}
		return NormalVal(newRadValues(i, node, asFloat))
	case rl.K_STRING:
		str := NewRadString("")

		contentsNode := rl.GetChild(node, rl.F_CONTENTS)

		// With current TS grammar, last character of closing delimiter is always the delimiter
		// Admittedly bad, very white boxy and brittle
		endNode := rl.GetChild(node, rl.F_END)
		endStr := i.GetSrcForNode(endNode)
		delimiterStr := endStr[len(endStr)-1]
		i.delimiterStack.Push(Delimiter{Open: string(delimiterStr)})

		if contentsNode != nil {
			for _, child := range contentsNode.Children(contentsNode.Walk()) {
				str = str.Concat(i.eval(&child).Val.RequireStr(i, &child))
			}
		}

		i.delimiterStack.Pop()

		return NormalVal(newRadValues(i, node, str))
	case rl.K_BOOL:
		asStr := i.GetSrcForNode(node)
		asBool, _ := strconv.ParseBool(asStr)
		return NormalVal(newRadValues(i, node, asBool))
	case rl.K_NULL:
		return NormalVal(newRadValues(i, node, nil))
	case rl.K_STRING_CONTENT:
		src := i.GetSrcForNode(node)
		return NormalVal(newRadValues(i, node, src))
	case rl.K_BACKSLASH:
		return NormalVal(newRadValues(i, node, "\\"))
	case rl.K_ESC_SINGLE_QUOTE:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == "'" {
			return NormalVal(newRadValues(i, node, "'"))
		} else {
			return NormalVal(newRadValues(i, node, `\'`))
		}
	case rl.K_ESC_DOUBLE_QUOTE:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == `"` {
			return NormalVal(newRadValues(i, node, `"`))
		} else {
			return NormalVal(newRadValues(i, node, `\"`))
		}
	case rl.K_ESC_BACKTICK:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == "`" {
			return NormalVal(newRadValues(i, node, "`"))
		} else {
			return NormalVal(newRadValues(i, node, "\\`"))
		}
	case rl.K_ESC_NEWLINE:
		return NormalVal(newRadValues(i, node, "\n"))
	case rl.K_ESC_TAB:
		return NormalVal(newRadValues(i, node, "\t"))
	case rl.K_ESC_OPEN_BRACKET:
		return NormalVal(newRadValues(i, node, "{"))
	case rl.K_INTERPOLATION:
		exprResult := evaluateInterpolation(i, node)
		return NormalVal(newRadValues(i, node, exprResult))
	case rl.K_ESC_BACKSLASH:
		return NormalVal(newRadValues(i, node, "\\"))
	case rl.K_LIST:
		entries := rl.GetChildren(node, rl.F_LIST_ENTRY)
		list := NewRadList()
		for _, entry := range entries {
			list.Append(i.eval(&entry).Val)
		}
		return NormalVal(newRadValues(i, node, list))
	case rl.K_MAP:
		radMap := NewRadMap()
		entryNodes := rl.GetChildren(node, rl.F_MAP_ENTRY)
		for _, entryNode := range entryNodes {
			keyNode := rl.GetChild(&entryNode, rl.F_KEY)
			valueNode := rl.GetChild(&entryNode, rl.F_VALUE)
			key := evalMapKey(i, keyNode)
			radMap.Set(key, i.eval(valueNode).Val)
		}
		return NormalVal(newRadValues(i, node, radMap))
	case rl.K_CALL:
		return NormalVal(i.callFunction(node, nil))
	case rl.K_FN_LAMBDA:
		return NormalVal(newRadValues(i, node, NewFn(i, node)))
	case rl.K_LIST_COMPREHENSION:
		resultExprNode := rl.GetChild(node, rl.F_EXPR)
		conditionNode := rl.GetChild(node, rl.F_CONDITION)

		resultList := NewRadList()
		doOneLoop := func() EvalResult {
			if conditionNode == nil || i.eval(conditionNode).Val.TruthyFalsy() {
				resultList.Append(i.eval(resultExprNode).Val)
			}
			return VoidNormal
		}
		i.executeForLoop(node, doOneLoop)

		return NormalVal(newRadValues(i, node, resultList))
	case rl.K_TERNARY_EXPR:
		delegateNode := rl.GetChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.eval(delegateNode)
		}

		conditionNode := rl.GetChild(node, rl.F_CONDITION)
		trueNode := rl.GetChild(node, rl.F_TRUE_BRANCH)
		falseNode := rl.GetChild(node, rl.F_FALSE_BRANCH)
		condition := i.eval(conditionNode).Val.TruthyFalsy()
		return i.eval(lo.Ternary(condition, trueNode, falseNode))
	default:
		i.errorf(node, "Unsupported node kind: %s", node.Kind())
	}
	return
}

func (i *Interpreter) evalRights(node *ts.Node) RadValue {
	rightNodes := rl.GetChildren(node, rl.F_RIGHT)
	if len(rightNodes) == 0 {
		return VOID_SENTINEL
	}
	if len(rightNodes) == 1 {
		return i.eval(&rightNodes[0]).Val
	}
	list := NewRadList()
	for _, rightNode := range rightNodes {
		list.Append(i.eval(&rightNode).Val)
	}
	return newRadValue(i, node, list)
}

func evaluateInterpolation(i *Interpreter, interpNode *ts.Node) RadValue {
	exprNode := rl.GetChild(interpNode, rl.F_EXPR)
	formatNode := rl.GetChild(interpNode, rl.F_FORMAT)

	exprResult := i.eval(exprNode).Val
	resultType := exprResult.Type()

	if formatNode == nil {
		switch resultType {
		case rl.RadStrT:
			// to maintain attributes
			return exprResult
		case rl.RadErrorT:
			return newRadValue(i, exprNode, exprResult.RequireError(i, interpNode).Msg())
		default:
			return newRadValue(i, exprNode, NewRadString(ToPrintable(exprResult)))
		}
	}

	thousandsSeparatorNode := rl.GetChild(formatNode, rl.F_THOUSANDS_SEPARATOR)
	alignmentNode := rl.GetChild(formatNode, rl.F_ALIGNMENT)
	paddingNode := rl.GetChild(formatNode, rl.F_PADDING)
	precisionNode := rl.GetChild(formatNode, rl.F_PRECISION)

	var goFmt strings.Builder
	goFmt.WriteString("%")

	if alignmentNode != nil {
		alignment := i.GetSrcForNode(alignmentNode)
		if alignment == "<" {
			goFmt.WriteString("-")
		}
	}

	if paddingNode != nil {
		padding := i.eval(paddingNode).Val.RequireInt(i, paddingNode)
		if exprStr, ok := exprResult.TryGetStr(); ok {
			// is string, need to account for color chars (increase padding len if present)
			plainLen := exprStr.Len()
			coloredLen := int64(com.StrLen(exprStr.String()))
			diff := coloredLen - plainLen
			padding += diff
		}

		goFmt.WriteString(fmt.Sprint(padding))
	}

	if precisionNode != nil {
		precision := i.eval(precisionNode).Val.RequireInt(i, precisionNode)

		if resultType != rl.RadIntT && resultType != rl.RadFloatT {
			precisionStr := "." + i.GetSrcForNode(precisionNode)
			i.errorf(interpNode, "Cannot format %s with a precision %q", TypeAsString(exprResult), precisionStr)
		}

		goFmt.WriteString(fmt.Sprintf(".%d", precision))
	}

	formatted := func() string {
		// If thousands separator is requested, render with fmt first, then group digits.
		if thousandsSeparatorNode != nil {
			if resultType != rl.RadIntT && resultType != rl.RadFloatT {
				i.errorf(interpNode, "Cannot format %s with thousands separator ','", TypeAsString(exprResult))
			}

			// 1) Render once with the right precision (if any)
			var s string
			if precisionNode != nil {
				p := int(i.eval(precisionNode).Val.RequireInt(i, precisionNode))
				if p < 0 {
					i.errorf(interpNode, "Precision cannot be negative: %d", p)
				}
				if resultType == rl.RadIntT {
					s = fmt.Sprintf("%.*f", p, float64(exprResult.Val.(int64)))
				} else {
					s = fmt.Sprintf("%.*f", p, exprResult.Val.(float64))
				}
			} else {
				if resultType == rl.RadIntT {
					s = fmt.Sprintf("%d", exprResult.Val.(int64))
				} else {
					s = strconv.FormatFloat(exprResult.Val.(float64), 'f', -1, 64)
				}
			}

			// 2) Add commas to the integer part only
			s = addThousands(s)

			// 3) Optional padding/alignment
			if paddingNode != nil {
				pad := int(i.eval(paddingNode).Val.RequireInt(i, paddingNode))
				align := ""
				if alignmentNode != nil {
					align = i.GetSrcForNode(alignmentNode)
				}
				if align == "<" {
					return fmt.Sprintf("%-*s", pad, s)
				}
				return fmt.Sprintf("%*s", pad, s)
			}

			return s
		}

		// Use existing formatting for non-comma cases
		switch resultType {
		case rl.RadIntT:
			if precisionNode == nil {
				goFmt.WriteString("d")
				return fmt.Sprintf(goFmt.String(), int(exprResult.Val.(int64)))
			} else {
				goFmt.WriteString("f")
				return fmt.Sprintf(goFmt.String(), float64(exprResult.Val.(int64)))
			}
		case rl.RadFloatT:
			goFmt.WriteString("f")
			return fmt.Sprintf(goFmt.String(), exprResult.Val)
		default:
			goFmt.WriteString("s")
			return fmt.Sprintf(goFmt.String(), ToPrintableQuoteStr(exprResult, false))
		}
	}()

	return newRadValue(i, interpNode, formatted)
}

func addThousands(num string) string {
	// Keep sign
	sign := ""
	if len(num) > 0 && (num[0] == '-' || num[0] == '+') {
		sign, num = num[:1], num[1:]
	}

	// Split on decimal point (keep '.' in frac)
	dot := strings.IndexByte(num, '.')
	intPart, frac := num, ""
	if dot >= 0 {
		intPart, frac = num[:dot], num[dot:]
	}

	n := len(intPart)
	if n <= 3 {
		return sign + intPart + frac
	}

	var b strings.Builder
	// First chunk can be 1–3 digits
	rem := n % 3
	if rem == 0 {
		rem = 3
	}
	b.WriteString(intPart[:rem])
	for i := rem; i < n; i += 3 {
		b.WriteByte(',')
		b.WriteString(intPart[i : i+3])
	}
	return sign + b.String() + frac
}

func (i *Interpreter) getOnlyChild(node *ts.Node) *ts.Node {
	count := node.ChildCount()
	if count != 1 {
		i.errorf(node, "Bug? Expected exactly one child, got %d", count)
	}
	return node.Child(0)
}

func (i *Interpreter) errorf(node *ts.Node, oneLinerFmt string, args ...interface{}) {
	RP.CtxErrorExit(NewCtx(i.GetSrc(), node, fmt.Sprintf(oneLinerFmt, args...), ""))
}

func (i *Interpreter) errorDetailsf(node *ts.Node, details string, oneLinerFmt string, args ...interface{}) {
	RP.CtxErrorExit(NewCtx(i.GetSrc(), node, fmt.Sprintf(oneLinerFmt, args...), details))
}

func (i *Interpreter) doVarPathAssign(varPathNode *ts.Node, rightValue RadValue, updateEnclosing bool) {
	rootIdentifier := rl.GetChild(varPathNode, rl.F_ROOT) // identifier required by grammar
	rootIdentifierName := i.GetSrcForNode(rootIdentifier)
	indexings := rl.GetChildren(varPathNode, rl.F_INDEXING)
	val, ok := i.env.GetVar(rootIdentifierName)

	if len(indexings) == 0 {
		// simple assignment, no collection lookups
		i.env.SetVarUpdatingEnclosing(rootIdentifierName, rightValue, updateEnclosing)
		return
	}

	// modifying collection
	if !ok {
		// modifying collection must exist
		i.errorf(rootIdentifier, "Undefined variable: %s", rootIdentifierName)
	}
	for _, index := range indexings[:len(indexings)-1] {
		val = val.Index(i, &index)
	}
	// val is now the collection to modify, using the last index
	lastIndex := indexings[len(indexings)-1]
	val.ModifyIdx(i, &lastIndex, rightValue)
}

// setLoopContext creates and sets the loop context variable if the user specified 'with <name>'.
// The context is a map containing 'idx' (current iteration index) and 'src' (original collection).
func (i *Interpreter) setLoopContext(contextNode *ts.Node, idx int64, srcValue RadValue) {
	if contextNode == nil {
		return
	}
	ctxName := i.GetSrcForNode(contextNode)
	ctx := NewRadMap()
	ctx.Set(newRadValue(i, contextNode, "idx"), newRadValue(i, contextNode, idx))
	ctx.Set(newRadValue(i, contextNode, "src"), srcValue)
	i.env.SetVar(ctxName, newRadValue(i, contextNode, ctx))
}

func (i *Interpreter) executeForLoop(node *ts.Node, doOneLoop func() EvalResult) EvalResult {
	leftsNode := rl.GetChild(node, rl.F_LEFTS)
	rightNode := rl.GetChild(node, rl.F_RIGHT)
	contextNode := rl.GetChild(node, rl.F_CONTEXT)

	res := i.eval(rightNode)
	switch coercedRight := res.Val.Val.(type) {
	case RadString:
		return runForLoopList(i, leftsNode, rightNode, contextNode, coercedRight.ToRuneList(), res.Val, doOneLoop)
	case *RadList:
		return runForLoopList(i, leftsNode, rightNode, contextNode, coercedRight, res.Val, doOneLoop)
	case *RadMap:
		return runForLoopMap(i, leftsNode, contextNode, coercedRight, res.Val, doOneLoop)
	default:
		i.errorf(rightNode, "Cannot iterate through a %s", TypeAsString(res.Val))
		panic(UNREACHABLE)
	}
}

func runForLoopList(
	i *Interpreter,
	leftsNode, rightNode, contextNode *ts.Node,
	list *RadList,
	srcValue RadValue,
	doOneLoop func() EvalResult,
) EvalResult {
	leftNodes := rl.GetChildren(leftsNode, rl.F_LEFT)

	if len(leftNodes) == 0 {
		i.errorf(leftsNode, "Expected at least one variable on the left side of for loop")
	}

	// Copy source for context.src to ensure it's an immutable snapshot
	var srcCopy RadValue
	if contextNode != nil {
		if srcList, ok := srcValue.TryGetList(); ok {
			srcCopy = newRadValue(i, contextNode, srcList.ShallowCopy())
		} else {
			srcCopy = srcValue // Fallback (shouldn't happen for list iteration)
		}
	}

Loop:
	for idx, val := range list.Values {
		i.setLoopContext(contextNode, int64(idx), srcCopy)

		if len(leftNodes) == 1 {
			itemNode := &leftNodes[0]
			itemName := i.GetSrcForNode(itemNode)
			i.env.SetVar(itemName, val)
		} else {
			// Multiple variables = unpacking (expecting list of lists)
			listInList, ok := val.TryGetList()
			if !ok {
				// Migration hint for old syntax
				firstName := i.GetSrcForNode(&leftNodes[0])
				if firstName == "idx" || firstName == "index" || firstName == "i" || firstName == "_" {
					i.errorf(rightNode, "Cannot unpack %q into %d values\n\n"+
						"Note: The for-loop syntax changed. It looks like you may be using the old syntax.\n"+
						"Old: for idx, item in items:\n"+
						"New: for item in items with loop:\n"+
						"         print(loop.idx, item)\n\n"+
						"See: https://amterp.github.io/rad/migrations/v0.7/",
						TypeAsString(val), len(leftNodes))
				}
				i.errorf(rightNode, "Cannot unpack %q into %d values", TypeAsString(val), len(leftNodes))
			}

			if listInList.LenInt() < len(leftNodes) {
				i.errorf(rightNode, "Expected at least %s in inner list, got %d",
					com.Pluralize(len(leftNodes), "value"), listInList.LenInt())
			}

			for j, itemNode := range leftNodes {
				itemName := i.GetSrcForNode(&itemNode)
				i.env.SetVar(itemName, listInList.Values[j])
			}
		}

		res := doOneLoop()
		switch res.Ctrl {
		case CtrlBreak:
			break Loop
		case CtrlReturn, CtrlYield:
			return res
		}
	}
	return VoidNormal
}

func runForLoopMap(
	i *Interpreter,
	leftsNode, contextNode *ts.Node,
	radMap *RadMap,
	srcValue RadValue,
	doOneLoop func() EvalResult,
) EvalResult {
	var keyNode *ts.Node
	var valueNode *ts.Node

	leftNodes := rl.GetChildren(leftsNode, rl.F_LEFT)
	numLefts := len(leftNodes)

	if numLefts == 0 || numLefts > 2 {
		i.errorf(leftsNode, "Expected 1 or 2 variables on left side of for loop")
	}

	keyNode = &leftNodes[0]
	if numLefts == 2 {
		valueNode = &leftNodes[1]
	}

	// Copy source for context.src to ensure it's an immutable snapshot
	var srcCopy RadValue
	if contextNode != nil {
		srcCopy = newRadValue(i, contextNode, radMap.ShallowCopy())
	}

	idx := int64(0)
Loop:
	for _, key := range radMap.Keys() {
		i.setLoopContext(contextNode, idx, srcCopy)

		keyName := i.GetSrcForNode(keyNode)
		i.env.SetVar(keyName, key)

		if valueNode != nil {
			valueName := i.GetSrcForNode(valueNode)
			value, _ := radMap.Get(key)
			i.env.SetVar(valueName, value)
		}

		res := doOneLoop()
		switch res.Ctrl {
		case CtrlBreak:
			break Loop
		case CtrlReturn, CtrlYield:
			return res
		}
		idx++
	}
	return VoidNormal
}

// if stmts, for loops
func (i *Interpreter) runBlock(stmtNodes []ts.Node) EvalResult {
	var res EvalResult
	for _, stmtNode := range stmtNodes {
		res = i.eval(&stmtNode)
		if res.Ctrl != CtrlNormal {
			break
		}
	}
	return res
}

func (i *Interpreter) runWithChildEnv(runnable func()) {
	originalEnv := i.env
	env := originalEnv.NewChildEnv()
	i.env = &env
	runnable()
	i.env = originalEnv
}

func (i *Interpreter) evaluateIndexing(rootNode *ts.Node, index ts.Node, val RadValue) RadValue {
	if index.Kind() == rl.K_CALL {
		// ufcs
		ufcsArg := &PosArg{
			// todo 'rootNode' is not great to use, it misses indexes in between that and this call,
			//  resulting in bad error pointing. could potentially replace ts.Node with interface
			//  'Pointable' i.e. a range we can point to in an error, that's ultimately all we need (?)
			node:  rootNode,
			value: val,
		}
		return i.callFunction(&index, ufcsArg)
	} else {
		return val.Index(i, &index)
	}
}

func (i *Interpreter) assignRightsToLefts(leftNodes []ts.Node, rightNodes []ts.Node, destructure bool) EvalResult {
	if destructure {
		if len(rightNodes) == 1 {
			rightNode := rightNodes[0]
			if rightNode.Kind() == rl.K_JSON_PATH {
				jsonFieldVar := NewJsonFieldVar(i, &leftNodes[0], &rightNode)
				i.env.SetJsonFieldVar(jsonFieldVar)
			} else {
				val := i.eval(&rightNode).Val
				list, ok := val.TryGetList()
				if ok {
					for idx, leftNode := range leftNodes {
						if len(list.Values) > idx {
							val := list.Values[idx]
							i.doVarPathAssign(&leftNode, val, false)
						} else {
							i.doVarPathAssign(&leftNode, RAD_NULL_VAL, false)
						}
					}
					return VoidNormal
				} else {
					i.doVarPathAssign(&leftNodes[0], val, false)
				}
			}

			for _, leftNode := range leftNodes[1:] {
				i.doVarPathAssign(&leftNode, RAD_NULL_VAL, false)
			}
			return VoidNormal
		}

		for idx, leftNode := range leftNodes {
			if len(rightNodes) > idx {
				rightNode := rightNodes[idx]
				if rightNode.Kind() == rl.K_JSON_PATH {
					jsonFieldVar := NewJsonFieldVar(i, &leftNode, &rightNode)
					i.env.SetJsonFieldVar(jsonFieldVar)
				} else {
					val := i.eval(&rightNode).Val
					i.doVarPathAssign(&leftNode, val, false)
				}
			} else {
				i.doVarPathAssign(&leftNode, RAD_NULL_VAL, false)
			}
		}
		return VoidNormal
	}

	// not destructuring, means exactly 1 left node

	if len(rightNodes) == 1 {
		rightNode := rightNodes[0]
		if rightNode.Kind() == rl.K_JSON_PATH {
			jsonFieldVar := NewJsonFieldVar(i, &leftNodes[0], &rightNode)
			i.env.SetJsonFieldVar(jsonFieldVar)
		} else {
			res := i.eval(&rightNodes[0])
			if res.Ctrl != CtrlNormal {
				return res
			}
			if res.Val == VOID_SENTINEL {
				i.errorf(&rightNode, "Cannot assign to a void value")
			}
			i.doVarPathAssign(&leftNodes[0], res.Val, false)
		}
		return VoidNormal
	}

	// not destructuring (so 1 left) & not 1 right node;
	// means at least 2 right nodes -> pack into list and assign to 1 left

	list := NewRadList()
	for _, rightNode := range rightNodes {
		val := i.eval(&rightNode).Val
		list.Append(val)
	}
	i.doVarPathAssign(&leftNodes[0], newRadValueList(list), false)
	return VoidNormal
}

func (i *Interpreter) defineCustomNamedFunction(fnNamedNode ts.Node) {
	nameNode := rl.GetChild(&fnNamedNode, rl.F_NAME)
	name := i.GetSrcForNode(nameNode)
	lambda := NewFn(i, &fnNamedNode)
	i.env.SetVar(name, newRadValueFn(lambda))
}

func (i *Interpreter) GetSrc() string {
	if i.tmpSrc != nil {
		return *i.tmpSrc
	}
	return i.sd.Src
}

func (i *Interpreter) GetSrcForNode(node *ts.Node) string {
	return i.GetSrc()[node.StartByte():node.EndByte()]
}

// todo this is somewhat hacky, not a fan. only use when you're extremely sure fn won't panic
func (i *Interpreter) WithTmpSrc(tmpSrc string, fn func()) {
	i.tmpSrc = &tmpSrc
	defer func() {
		i.tmpSrc = nil
	}()
	fn()
}

func (i *Interpreter) executeSwitchCase(caseValueAltNode *ts.Node) EvalResult {
	switch caseValueAltNode.Kind() {
	case rl.K_SWITCH_CASE_EXPR:
		return NormalVal(i.evalRights(caseValueAltNode))
	case rl.K_SWITCH_CASE_BLOCK:
		stmtNodes := rl.GetChildren(caseValueAltNode, rl.F_STMT)
		res := i.runBlock(stmtNodes)
		switch res.Ctrl {
		case CtrlNormal, CtrlBreak, CtrlContinue, CtrlReturn:
			return res
		case CtrlYield:
			return NormalVal(res.Val)
		}
	default:
		i.errorf(caseValueAltNode, "Bug! Unsupported switch case value node kind: %s", caseValueAltNode.Kind())
	}
	return VoidNormal
}

// getAssignLeftNodes returns the left-hand side nodes from an assignment,
// trying F_LEFT first, then F_LEFTS.
func (i *Interpreter) getAssignLeftNodes(node *ts.Node) []ts.Node {
	leftNodes := rl.GetChildren(node, rl.F_LEFT)
	if len(leftNodes) == 0 {
		leftNodes = rl.GetChildren(node, rl.F_LEFTS)
	}
	return leftNodes
}

// handlePanicRecovery handles panic recovery with RadPanic-aware error reporting.
// Should be called from a deferred function with the result of recover().
// fallbackNode is used for error reporting when the panic is not a RadPanic.
// msgArgs are optional context values to include in the error message before the panic value.
func (i *Interpreter) handlePanicRecovery(r interface{}, fallbackNode *ts.Node, msgArgs ...interface{}) {
	if r != nil {
		radPanic, ok := r.(*RadPanic)
		if ok {
			err := radPanic.Err()
			msg := err.Msg().Plain()
			if !com.IsBlank(string(err.Code)) {
				msg += fmt.Sprintf(" (%s)", err.Code)
			}
			i.errorf(err.Node, msg)
		}
		if !IsTest {
			// Build format string: "Bug! Panic: %v %v ... %v\n%s"
			// One %v for each msgArg, one for panic value, one %s for stack trace
			var fmtStr strings.Builder
			fmtStr.WriteString("Bug! Panic:")
			for range msgArgs {
				fmtStr.WriteString(" %v")
			}
			fmtStr.WriteString(" %v\n%s") // panic value and stack trace

			// Append panic value and stack trace to the provided args
			allArgs := append(msgArgs, r, debug.Stack())
			i.errorf(fallbackNode, fmtStr.String(), allArgs...)
		}
	}
}

// withCatch wraps body execution with panic catching. If catchNode is nil, just executes body.
// On RadPanic, calls onErr callback to handle the error (assign variables, run catch block, etc.).
// Propagates control flow (return/break/continue/yield) from the catch block.
// Re-panics non-RadPanic errors to preserve Go's panic semantics (e.g., runtime errors, bugs).
func (i *Interpreter) withCatch(
	catchNode *ts.Node,
	onErr func(rp *RadPanic) EvalResult,
	body func() EvalResult,
) (out EvalResult) {
	if catchNode == nil {
		return body()
	}

	defer func() {
		if r := recover(); r != nil {
			if rp, ok := r.(*RadPanic); ok {
				// Rad error occurred - run error handler
				out = onErr(rp)
			} else {
				// Non-RadPanic errors (runtime panics, bugs) must propagate unchanged
				panic(r)
			}
		}
	}()

	return body()
}
