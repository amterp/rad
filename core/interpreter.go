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

// CallFrame represents a function call in the Rad call stack.
// Used for providing stack traces in error messages.
type CallFrame struct {
	FunctionName string   // Name of the function (or "<anonymous>" for lambdas)
	CallSite     *rl.Span // Where the function was called from
	DefSite      *rl.Span // Where the function is defined
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

	// Call stack for Rad function calls (not Go stack)
	callStack []CallFrame
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
		// arg.GetSpan() provides span info from the arg declaration for error reporting.

		// Special handling for variadic arguments with list defaults
		if arg.IsVariadic() {
			var hasListDefaults bool
			var defaultValue interface{}

			if stringListArg, ok := arg.(*StringListRadArg); ok && stringListArg.scriptArg != nil && stringListArg.scriptArg.DefaultStringList != nil {
				hasListDefaults = true
				defaultValue = NewRadListFromGeneric(i, nil, *stringListArg.scriptArg.DefaultStringList)
			} else if intListArg, ok := arg.(*IntListRadArg); ok && intListArg.scriptArg != nil && intListArg.scriptArg.DefaultIntList != nil {
				hasListDefaults = true
				defaultValue = NewRadListFromGeneric(i, nil, *intListArg.scriptArg.DefaultIntList)
			} else if floatListArg, ok := arg.(*FloatListRadArg); ok && floatListArg.scriptArg != nil && floatListArg.scriptArg.DefaultFloatList != nil {
				hasListDefaults = true
				defaultValue = NewRadListFromGeneric(i, nil, *floatListArg.scriptArg.DefaultFloatList)
			} else if boolListArg, ok := arg.(*BoolListRadArg); ok && boolListArg.scriptArg != nil && boolListArg.scriptArg.DefaultBoolList != nil {
				hasListDefaults = true
				defaultValue = NewRadListFromGeneric(i, nil, *boolListArg.scriptArg.DefaultBoolList)
			}

			if hasListDefaults {
				if !arg.Configured() {
					env.SetVar(arg.GetIdentifier(), newRadValue(i, nil, defaultValue))
				} else {
					goto normalProcessing
				}
				continue
			}
		}

		if !arg.IsDefined() {
			if arg.IsVariadic() {
				env.SetVar(arg.GetIdentifier(), newRadValue(i, nil, NewRadList()))
			} else {
				env.SetVar(arg.GetIdentifier(), RAD_NULL_VAL)
			}
			continue
		}

	normalProcessing:
		switch coerced := arg.(type) {
		case *BoolRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, nil, coerced.Value))
		case *BoolListRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, nil, NewRadListFromGeneric(i, nil, coerced.Value)))
		case *StringRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, nil, coerced.Value))
		case *StringListRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, nil, NewRadListFromGeneric(i, nil, coerced.Value)))
		case *IntRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, nil, coerced.Value))
		case *IntListRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, nil, NewRadListFromGeneric(i, nil, coerced.Value)))
		case *FloatRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, nil, coerced.Value))
		case *FloatListRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, nil, NewRadListFromGeneric(i, nil, coerced.Value)))
		default:
			i.emitErrorf(rl.ErrInternalBug, nil, "Unsupported arg type, cannot init: %T", arg)
		}
	}
}

func (i *Interpreter) Run() {
	// Convert CST to AST once, then interpret the AST
	astRoot := rts.ConvertCST(i.sd.Tree.Root(), i.sd.Src, i.sd.ScriptName)

	// PHASE 1: Execute top-level code (always)
	res := i.safelyExecuteTopLevel(astRoot)
	if res.Ctrl != CtrlNormal {
		i.emitErrorf(rl.ErrInternalBug, astRoot, "Bug: Unexpected control flow: %v", res.Ctrl)
	}

	// PHASE 2: Execute command callback (if command was invoked)
	if i.invokedCommand != nil {
		i.safelyExecuteCommandCallback(i.invokedCommand)
	}
}

func (i *Interpreter) safelyExecuteTopLevel(root *rl.SourceFile) EvalResult {
	defer func() {
		i.handlePanicRecovery(recover(), root)
	}()

	// First pass: define custom named functions (function hoisting)
	for _, stmt := range root.Stmts {
		if fnDef, ok := stmt.(*rl.FnDef); ok {
			i.defineCustomNamedFunction(fnDef)
		}
	}

	// Second pass: evaluate all statements
	var lastResult EvalResult
	for _, stmt := range root.Stmts {
		lastResult = i.eval(stmt)
	}

	return lastResult
}

func (i *Interpreter) safelyExecuteCommandCallback(cmd *ScriptCommand) {
	defer func() {
		// Use nil node since we don't have a specific node for the command invocation
		i.handlePanicRecovery(recover(), nil, cmd.ExternalName)
	}()

	if cmd.IsLambdaCallback {
		lambdaAST := cmd.CallbackLambda
		if lambdaAST == nil {
			panic(fmt.Sprintf("Bug! Lambda AST is nil for command: %s", cmd.ExternalName))
		}
		fn := NewFnFromAST(i, lambdaAST.Typing, lambdaAST.Body, lambdaAST.IsBlock, &lambdaAST.DefSpan)

		// Execute the lambda with no arguments
		_ = fn.Execute(NewFnInvocation(i, lambdaAST, "<lambda>", []PosArg{}, make(map[string]namedArg), false))
	} else {
		// Look up the function by name and call it
		if cmd.CallbackName == nil {
			panic(fmt.Sprintf("Bug! Command '%s' has no callback", cmd.ExternalName))
		}
		funcName := *cmd.CallbackName
		val, exist := i.env.GetVar(funcName)
		if !exist {
			i.emitErrorf(rl.ErrUnknownFunction, nil, "Cannot invoke unknown function '%s' for command '%s'", funcName, cmd.ExternalName)
		}

		fn, ok := val.TryGetFn()
		if !ok {
			i.emitErrorf(rl.ErrTypeMismatch, nil, "Cannot invoke '%s' as a function for command '%s': it is a %s",
				funcName, cmd.ExternalName, val.Type().AsString())
		}

		// Execute the function with no arguments
		_ = fn.Execute(NewFnInvocation(i, nil, funcName, []PosArg{}, make(map[string]namedArg), fn.IsBuiltIn()))
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
	if tree.HasInvalidNodes() {
		return RAD_NULL_VAL, fmt.Errorf("parse error in statement: %s", input)
	}

	// Convert CST to AST
	astRoot := rts.ConvertCST(tree.Root(), input, "<repl>")

	// Update the interpreter's ScriptData to point to this new tree and source
	// This ensures that GetSrc and other methods work correctly
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

	// REPL: evaluate each statement from the AST
	if len(astRoot.Stmts) == 1 {
		result := i.safelyEvaluate(astRoot.Stmts[0])
		if result.Ctrl != CtrlNormal {
			return RAD_NULL_VAL, fmt.Errorf("unexpected control flow: %v", result.Ctrl)
		}
		return result.Val, nil
	}

	// Multiple statements: evaluate all, return last
	var lastResult EvalResult
	for _, stmt := range astRoot.Stmts {
		lastResult = i.safelyEvaluate(stmt)
	}

	// Check for any evaluation errors from defer
	if evalErr != nil {
		return RAD_NULL_VAL, evalErr
	}

	if lastResult.Ctrl != CtrlNormal {
		return RAD_NULL_VAL, fmt.Errorf("unexpected control flow: %v", lastResult.Ctrl)
	}
	return lastResult.Val, nil
}

func (i *Interpreter) safelyEvaluate(node rl.Node) EvalResult {
	defer func() {
		i.handlePanicRecovery(recover(), node)
	}()
	return i.eval(node)
}

func (i *Interpreter) eval(node rl.Node) (out EvalResult) {
	out = VoidNormal

	switch n := node.(type) {
	// --- Statements ---
	case *rl.SourceFile:
		// First pass: function hoisting
		for _, stmt := range n.Stmts {
			if fnDef, ok := stmt.(*rl.FnDef); ok {
				i.defineCustomNamedFunction(fnDef)
			}
		}
		for _, stmt := range n.Stmts {
			i.eval(stmt)
		}

	case *rl.Assign:
		out = i.withCatch(n.Catch, func(rp *RadPanic) EvalResult {
			// Assign error to first variable, null to rest
			if len(n.Targets) > 0 {
				i.doVarPathAssign(n.Targets[0], rp.ErrV, false)
				for j := 1; j < len(n.Targets); j++ {
					i.doVarPathAssign(n.Targets[j], RAD_NULL_VAL, false)
				}
			}

			// Run catch block and propagate control flow
			res := i.runBlock(n.Catch.Stmts)
			if res.Ctrl != CtrlNormal {
				return res
			}
			return VoidNormal
		}, func() EvalResult {
			return i.evalAssign(n)
		})

	case *rl.ExprStmt:
		out = i.withCatch(n.Catch, func(rp *RadPanic) EvalResult {
			res := i.runBlock(n.Catch.Stmts)
			if res.Ctrl != CtrlNormal {
				return res
			}
			return VoidNormal
		}, func() EvalResult {
			return i.eval(n.Expr)
		})

	case *rl.Pass:
		// no-op

	case *rl.Return:
		return ReturnVal(i.evalValues(n, n.Values))

	case *rl.Yield:
		return YieldVal(i.evalValues(n, n.Values))

	case *rl.Break:
		if i.forWhileLoopLevel > 0 {
			return VoidBreak
		}
		i.emitError(rl.ErrBreakOutsideLoop, n, "Cannot 'break' outside of a loop")

	case *rl.Continue:
		if i.forWhileLoopLevel > 0 {
			return VoidContinue
		}
		i.emitError(rl.ErrContinueOutsideLoop, n, "Cannot 'continue' outside of a loop")

	case *rl.ForLoop:
		i.forWhileLoopLevel++
		defer func() { i.forWhileLoopLevel-- }()
		return i.executeForLoop(n, func() EvalResult { return i.runBlock(n.Body) })

	case *rl.WhileLoop:
		i.forWhileLoopLevel++
		defer func() { i.forWhileLoopLevel-- }()
		for {
			condValue := true
			if n.Condition != nil {
				condValue = i.eval(n.Condition).Val.TruthyFalsy()
			}
			if condValue {
				res := i.runBlock(n.Body)
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

	case *rl.If:
		for _, branch := range n.Branches {
			shouldExecute := true
			if branch.Condition != nil {
				shouldExecute = i.eval(branch.Condition).Val.TruthyFalsy()
			}
			if shouldExecute {
				return i.runBlock(branch.Body)
			}
		}

	case *rl.Switch:
		discriminantVal := i.eval(n.Discriminant).Val

		var matchedCases []rl.SwitchCase
		for _, sc := range n.Cases {
			for _, key := range sc.Keys {
				caseKey := i.eval(key).Val
				if caseKey.Equals(discriminantVal) {
					matchedCases = append(matchedCases, sc)
					break
				}
			}
		}

		if len(matchedCases) == 0 {
			if n.Default != nil {
				return i.executeSwitchCaseAlt(n.Default.Alt)
			}
			i.emitError(rl.ErrSwitchNoMatch, n.Discriminant, "No matching case found for switch")
		}

		if len(matchedCases) > 1 {
			i.emitError(rl.ErrSwitchMultipleMatch, n.Discriminant, "Multiple matching cases found for switch")
		}

		return i.executeSwitchCaseAlt(matchedCases[0].Alt)

	case *rl.FnDef:
		// Non-top-level function definitions (top-level ones are hoisted in SourceFile)
		i.defineCustomNamedFunction(n)

	case *rl.Defer:
		i.deferBlocks = append(i.deferBlocks, NewDeferBlock(n.Span(), n.Body, n.IsErrDefer))

	case *rl.Shell:
		out = i.executeShellStmt(n)

	case *rl.Del:
		for _, target := range n.Targets {
			i.doVarPathAssign(target, VOID_SENTINEL, true)
		}

	case *rl.RadBlock:
		i.runRadBlock(n)

	// --- Expressions ---

	case *rl.OpBinary:
		return NormalVal(newRadValues(i, n, i.executeOp(n, n.Left, n.Right, n.Op, n.IsCompound)))

	case *rl.OpUnary:
		return NormalVal(i.executeUnaryOp(n))

	case *rl.Ternary:
		condition := i.eval(n.Condition).Val.TruthyFalsy()
		return i.eval(lo.Ternary(condition, n.True, n.False))

	case *rl.Fallback:
		var leftResult EvalResult
		panicked := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(*RadPanic); ok {
						panicked = true
					} else {
						panic(r)
					}
				}
			}()
			leftResult = i.eval(n.Left)
		}()
		if panicked {
			return i.eval(n.Right)
		}
		return leftResult

	case *rl.Call:
		return NormalVal(i.callFunction(n, nil))

	case *rl.Identifier:
		if n.Name == "_" {
			i.emitError(rl.ErrUnsupportedOperation, n, "Cannot use '_' as a value")
		}
		val, ok := i.env.GetVar(n.Name)
		if !ok {
			i.emitUndefinedVariableError(n, n.Name)
		}
		return NormalVal(newRadValues(i, n, val))

	case *rl.VarPath:
		val := i.eval(n.Root).Val
		for _, seg := range n.Segments {
			val = i.evaluatePathSegment(n.Root, seg, val)
		}
		return NormalVal(newRadValues(i, n, val))

	case *rl.LitInt:
		return NormalVal(newRadValues(i, n, n.Value))

	case *rl.LitFloat:
		return NormalVal(newRadValues(i, n, n.Value))

	case *rl.LitBool:
		return NormalVal(newRadValues(i, n, n.Value))

	case *rl.LitNull:
		return NormalVal(newRadValues(i, n, nil))

	case *rl.LitString:
		return NormalVal(newRadValues(i, n, i.evalString(n)))

	case *rl.LitList:
		list := NewRadList()
		for _, elem := range n.Elements {
			list.Append(i.eval(elem).Val)
		}
		return NormalVal(newRadValues(i, n, list))

	case *rl.LitMap:
		radMap := NewRadMap()
		for _, entry := range n.Entries {
			key := i.eval(entry.Key).Val
			radMap.Set(key, i.eval(entry.Value).Val)
		}
		return NormalVal(newRadValues(i, n, radMap))

	case *rl.Lambda:
		fn := NewFnFromAST(i, n.Typing, n.Body, n.IsBlock, &n.DefSpan)
		return NormalVal(newRadValues(i, n, fn))

	case *rl.ListComp:
		resultList := NewRadList()
		doOneLoop := func() EvalResult {
			if n.Condition == nil || i.eval(n.Condition).Val.TruthyFalsy() {
				resultList.Append(i.eval(n.Expr).Val)
			}
			return VoidNormal
		}
		i.executeForLoop(n, doOneLoop)
		return NormalVal(newRadValues(i, n, resultList))

	default:
		i.emitErrorf(rl.ErrInternalBug, node, "Unsupported AST node type: %T", node)
	}
	return
}

// evalValues evaluates a list of value nodes. Returns VOID_SENTINEL if empty,
// the single value if one, or packs into a list if multiple.
func (i *Interpreter) evalValues(parent rl.Node, values []rl.Node) RadValue {
	if len(values) == 0 {
		return VOID_SENTINEL
	}
	if len(values) == 1 {
		return i.eval(values[0]).Val
	}
	list := NewRadList()
	for _, v := range values {
		list.Append(i.eval(v).Val)
	}
	return newRadValue(i, parent, list)
}

// evalString evaluates an AST string node. Simple strings return their
// pre-resolved value; interpolated strings concatenate segments.
func (i *Interpreter) evalString(n *rl.LitString) RadString {
	if n.Simple {
		return NewRadString(n.Value)
	}

	str := NewRadString("")
	for _, seg := range n.Segments {
		if seg.IsLiteral {
			str = str.ConcatStr(seg.Text)
		} else {
			exprResult := i.eval(seg.Expr).Val
			if seg.Format != nil {
				str = str.Concat(i.formatInterpolation(seg, exprResult))
			} else {
				switch exprResult.Type() {
				case rl.RadStrT:
					str = str.Concat(exprResult.RequireStr(i, seg.Expr))
				case rl.RadErrorT:
					str = str.Concat(exprResult.RequireError(i, seg.Expr).Msg())
				default:
					str = str.ConcatStr(ToPrintable(exprResult))
				}
			}
		}
	}
	return str
}

// formatInterpolation applies format specifiers to an interpolation expression result.
func (i *Interpreter) formatInterpolation(seg rl.StringSegment, exprResult RadValue) RadString {
	f := seg.Format
	resultType := exprResult.Type()
	// Use the interpolation segment's span for error reporting
	strNode := rl.NewIdentifier(seg.Span(), "")

	var goFmt strings.Builder
	goFmt.WriteString("%")

	if f.Alignment == "<" {
		goFmt.WriteString("-")
	}

	if f.Padding != nil {
		padding := i.eval(f.Padding).Val.RequireInt(i, f.Padding)
		if exprStr, ok := exprResult.TryGetStr(); ok {
			plainLen := exprStr.Len()
			coloredLen := int64(com.StrLen(exprStr.String()))
			diff := coloredLen - plainLen
			padding += diff
		}
		goFmt.WriteString(fmt.Sprint(padding))
	}

	if f.Precision != nil {
		precision := i.eval(f.Precision).Val.RequireInt(i, f.Precision)
		if resultType != rl.RadIntT && resultType != rl.RadFloatT {
			i.emitErrorf(rl.ErrCannotFormat, strNode, "Cannot format %s with a precision", TypeAsString(exprResult))
		}
		goFmt.WriteString(fmt.Sprintf(".%d", precision))
	}

	formatted := func() string {
		if f.ThousandsSeparator {
			if resultType != rl.RadIntT && resultType != rl.RadFloatT {
				i.emitErrorf(rl.ErrCannotFormat, strNode, "Cannot format %s with thousands separator ','", TypeAsString(exprResult))
			}

			var s string
			if f.Precision != nil {
				p := int(i.eval(f.Precision).Val.RequireInt(i, f.Precision))
				if p < 0 {
					i.emitErrorf(rl.ErrNumInvalidRange, strNode, "Precision cannot be negative: %d", p)
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

			s = addThousands(s)

			if f.Padding != nil {
				pad := int(i.eval(f.Padding).Val.RequireInt(i, f.Padding))
				if f.Alignment == "<" {
					return fmt.Sprintf("%-*s", pad, s)
				}
				return fmt.Sprintf("%*s", pad, s)
			}
			return s
		}

		switch resultType {
		case rl.RadIntT:
			if f.Precision == nil {
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

	return NewRadString(formatted)
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

// evalAssign handles normal assignment (no catch) for an Assign node.
func (i *Interpreter) evalAssign(n *rl.Assign) EvalResult {
	if n.IsUnpacking {
		return i.assignRightsToLefts(n.Targets, n.Values, true, n.UpdateEnclosing)
	}
	return i.assignRightsToLefts(n.Targets, n.Values, false, n.UpdateEnclosing)
}

// emitDiagnostic renders a diagnostic and exits with error code 1.
// This is the new error reporting method that replaces errorf/errorDetailsf.
func (i *Interpreter) emitDiagnostic(d Diagnostic) {
	// Automatically attach call stack if not already present
	if len(d.CallStack) == 0 && len(i.callStack) > 0 {
		d = d.WithCallStack(i.CallStack())
	}
	renderer := NewDiagnosticRenderer(RIo.StdErr)
	renderer.Render(d)
	RExit.Exit(1)
}

// emitError creates and emits an error diagnostic with a single span.
// If node is nil, the diagnostic will have no source location.
func (i *Interpreter) emitError(code rl.Error, node rl.Node, message string) {
	if node != nil {
		span := node.Span()
		i.emitErrorSpan(code, &span, message)
	} else {
		i.emitErrorSpan(code, nil, message)
	}
}

// emitErrorSpan creates and emits an error diagnostic from a span pointer.
// Used when we have a span but no node (e.g. from RadError).
func (i *Interpreter) emitErrorSpan(code rl.Error, span *rl.Span, message string) {
	var diag Diagnostic
	if span != nil {
		diag = NewDiagnostic(SeverityError, code, message, i.GetSrc(), *span)
	} else {
		diag = Diagnostic{
			Severity: SeverityError,
			Code:     code,
			Message:  message,
			Source:   i.GetSrc(),
		}
	}
	i.emitDiagnostic(diag)
}

// emitErrorf creates and emits an error diagnostic with formatted message.
func (i *Interpreter) emitErrorf(code rl.Error, node rl.Node, format string, args ...interface{}) {
	i.emitError(code, node, fmt.Sprintf(format, args...))
}

// emitErrorWithHint creates and emits an error diagnostic with a hint.
// If node is nil, the diagnostic will have no source location.
func (i *Interpreter) emitErrorWithHint(code rl.Error, node rl.Node, message string, hint string) {
	var diag Diagnostic
	if node != nil {
		span := node.Span()
		diag = NewDiagnostic(SeverityError, code, message, i.GetSrc(), span).WithHint(hint)
	} else {
		diag = Diagnostic{
			Severity: SeverityError,
			Code:     code,
			Message:  message,
			Source:   i.GetSrc(),
			Hints:    []string{hint},
		}
	}
	i.emitDiagnostic(diag)
}

// emitErrorWithSecondary creates an error diagnostic with a secondary span (e.g., "assigned here").
// If primaryNode is nil, the diagnostic will only have the secondary span (if provided).
func (i *Interpreter) emitErrorWithSecondary(code rl.Error, primaryNode rl.Node, message string, secondarySpan *rl.Span, secondaryMsg string) {
	var labels []Label
	if primaryNode != nil {
		primarySpan := primaryNode.Span()
		labels = append(labels, NewPrimaryLabel(primarySpan, ""))
	}
	if secondarySpan != nil {
		labels = append(labels, NewSecondaryLabel(*secondarySpan, secondaryMsg))
	}
	diag := NewDiagnosticWithLabels(SeverityError, code, message, i.GetSrc(), labels)
	i.emitDiagnostic(diag)
}

// pushCallFrame pushes a new frame onto the call stack.
func (i *Interpreter) pushCallFrame(name string, callSite, defSite *rl.Span) {
	i.callStack = append(i.callStack, CallFrame{
		FunctionName: name,
		CallSite:     callSite,
		DefSite:      defSite,
	})
}

// popCallFrame removes the top frame from the call stack.
func (i *Interpreter) popCallFrame() {
	if len(i.callStack) > 0 {
		i.callStack = i.callStack[:len(i.callStack)-1]
	}
}

// CallStack returns a copy of the current call stack (most recent first).
func (i *Interpreter) CallStack() []CallFrame {
	result := make([]CallFrame, len(i.callStack))
	// Reverse so most recent is first
	for idx, frame := range i.callStack {
		result[len(i.callStack)-1-idx] = frame
	}
	return result
}

// emitUndefinedVariableError emits an error for an undefined variable with
// "did you mean?" suggestions for similar variable names.
func (i *Interpreter) emitUndefinedVariableError(node rl.Node, name string) {
	similar := i.env.FindSimilarVars(name, 3)

	if len(similar) > 0 {
		// Include suggestion hint
		hint := fmt.Sprintf("a variable with a similar name exists: %s", similar[0])
		if len(similar) > 1 {
			hint = fmt.Sprintf("variables with similar names exist: %s", strings.Join(similar, ", "))
		}
		i.emitErrorWithHint(rl.ErrUndefinedVariable, node,
			fmt.Sprintf("Undefined variable: %s", name), hint)
	} else {
		i.emitErrorf(rl.ErrUndefinedVariable, node, "Undefined variable: %s", name)
	}
}

func (i *Interpreter) doVarPathAssign(target rl.Node, rightValue RadValue, updateEnclosing bool) {
	switch n := target.(type) {
	case *rl.Identifier:
		i.env.SetVarUpdatingEnclosing(n.Name, rightValue, updateEnclosing)
	case *rl.VarPath:
		rootId, ok := n.Root.(*rl.Identifier)
		if !ok {
			i.emitError(rl.ErrInternalBug, target, "Bug: expected identifier as VarPath root in assignment")
			return
		}

		if len(n.Segments) == 0 {
			i.env.SetVarUpdatingEnclosing(rootId.Name, rightValue, updateEnclosing)
			return
		}

		val, exists := i.env.GetVar(rootId.Name)
		if !exists {
			i.emitUndefinedVariableError(rootId, rootId.Name)
		}
		// Navigate to the penultimate segment
		for _, seg := range n.Segments[:len(n.Segments)-1] {
			val = i.evaluatePathSegment(n.Root, seg, val)
		}
		// Apply the last segment as a modification
		lastSeg := n.Segments[len(n.Segments)-1]
		i.modifyByPathSegment(val, lastSeg, rightValue)
	default:
		i.emitErrorf(rl.ErrInternalBug, target, "Bug: unexpected assignment target type: %T", target)
	}
}

// modifyByPathSegment modifies a collection value using a path segment.
func (i *Interpreter) modifyByPathSegment(val RadValue, seg rl.PathSegment, rightValue RadValue) {
	if seg.Field != nil {
		key := newRadValueStr(*seg.Field)
		val.ModifyByKey(i, rl.NewIdentifier(seg.Span(), *seg.Field), key, rightValue)
	} else if seg.IsSlice {
		i.modifyBySlice(val, seg, rightValue)
	} else if seg.Index != nil {
		key := i.eval(seg.Index).Val
		val.ModifyByKey(i, seg.Index, key, rightValue)
	} else {
		i.emitError(rl.ErrInternalBug, nil, "Bug: path segment has no field, index, or slice")
	}
}

// modifyBySlice replaces a slice range in a list with the given value.
func (i *Interpreter) modifyBySlice(val RadValue, seg rl.PathSegment, rightValue RadValue) {
	list, ok := val.Val.(*RadList)
	if !ok {
		contextNode := seg.Start
		if contextNode == nil {
			contextNode = seg.End
		}
		i.emitErrorf(rl.ErrTypeMismatch, contextNode, "Cannot slice-assign to a %s", TypeAsString(val))
		return
	}

	start, end := i.resolveSliceBounds(seg, list.Len())
	if start < end {
		newList := NewRadList()
		newList.Values = append(newList.Values, list.Values[:start]...)
		if replacement, ok := rightValue.TryGetList(); ok {
			newList.Values = append(newList.Values, replacement.Values...)
		} else if rightValue == VOID_SENTINEL {
			// delete mode (from del statement)
		} else {
			i.emitError(rl.ErrTypeMismatch, seg.Start, "Cannot assign list slice to a non-list type")
		}
		newList.Values = append(newList.Values, list.Values[end:]...)
		list.Values = newList.Values
	}
}

// setLoopContext creates and sets the loop context variable if the for-loop has a context name.
func (i *Interpreter) setLoopContext(contextName *string, loopNode rl.Node, idx int64, srcValue RadValue) {
	if contextName == nil {
		return
	}
	ctx := NewRadMap()
	ctx.Set(newRadValue(i, loopNode, "idx"), newRadValue(i, loopNode, idx))
	ctx.Set(newRadValue(i, loopNode, "src"), srcValue)
	i.env.SetVar(*contextName, newRadValue(i, loopNode, ctx))
}

// executeForLoop dispatches a for-loop to the appropriate list or map iterator.
// Works for both ForLoop and ListComp AST nodes.
func (i *Interpreter) executeForLoop(node rl.Node, doOneLoop func() EvalResult) EvalResult {
	var vars []string
	var iterNode rl.Node
	var context *string

	switch n := node.(type) {
	case *rl.ForLoop:
		vars = n.Vars
		iterNode = n.Iter
		context = n.Context
	case *rl.ListComp:
		vars = n.Vars
		iterNode = n.Iter
		context = n.Context
	default:
		i.emitErrorf(rl.ErrInternalBug, node, "Bug: executeForLoop called with %T", node)
		panic(UNREACHABLE)
	}

	res := i.eval(iterNode)
	switch coercedRight := res.Val.Val.(type) {
	case RadString:
		return runForLoopList(i, node, vars, iterNode, context, coercedRight.ToRuneList(), res.Val, doOneLoop)
	case *RadList:
		return runForLoopList(i, node, vars, iterNode, context, coercedRight, res.Val, doOneLoop)
	case *RadMap:
		return runForLoopMap(i, node, vars, context, coercedRight, doOneLoop)
	default:
		i.emitErrorf(rl.ErrNotIterable, iterNode, "Cannot iterate through a %s", TypeAsString(res.Val))
		panic(UNREACHABLE)
	}
}

func runForLoopList(
	i *Interpreter,
	loopNode rl.Node,
	vars []string,
	iterNode rl.Node,
	context *string,
	list *RadList,
	srcValue RadValue,
	doOneLoop func() EvalResult,
) EvalResult {
	if len(vars) == 0 {
		i.emitError(rl.ErrInvalidSyntax, loopNode, "Expected at least one variable on the left side of for loop")
	}

	// Copy source for context.src to ensure it's an immutable snapshot
	var srcCopy RadValue
	if context != nil {
		if srcList, ok := srcValue.TryGetList(); ok {
			srcCopy = newRadValue(i, loopNode, srcList.ShallowCopy())
		} else {
			srcCopy = srcValue
		}
	}

Loop:
	for idx, val := range list.Values {
		i.setLoopContext(context, loopNode, int64(idx), srcCopy)

		if len(vars) == 1 {
			i.env.SetVar(vars[0], val)
		} else {
			listInList, ok := val.TryGetList()
			if !ok {
				if vars[0] == "idx" || vars[0] == "index" || vars[0] == "i" || vars[0] == "_" {
					i.emitErrorWithHint(rl.ErrUnpackMismatch, iterNode,
						fmt.Sprintf("Cannot unpack %q into %d values", TypeAsString(val), len(vars)),
						"The for-loop syntax changed. Use: for item in items with loop: print(loop.idx, item). See: https://amterp.github.io/rad/migrations/v0.7/")
				}
				i.emitErrorf(rl.ErrUnpackMismatch, iterNode, "Cannot unpack %q into %d values", TypeAsString(val), len(vars))
			}

			if listInList.LenInt() < len(vars) {
				i.emitErrorf(rl.ErrUnpackMismatch, iterNode, "Expected at least %s in inner list, got %d",
					com.Pluralize(len(vars), "value"), listInList.LenInt())
			}

			for j, varName := range vars {
				i.env.SetVar(varName, listInList.Values[j])
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
	loopNode rl.Node,
	vars []string,
	context *string,
	radMap *RadMap,
	doOneLoop func() EvalResult,
) EvalResult {
	numVars := len(vars)
	if numVars == 0 || numVars > 2 {
		i.emitError(rl.ErrInvalidSyntax, loopNode, "Expected 1 or 2 variables on left side of for loop")
	}

	var srcCopy RadValue
	if context != nil {
		srcCopy = newRadValue(i, loopNode, radMap.ShallowCopy())
	}

	idx := int64(0)
Loop:
	for _, key := range radMap.Keys() {
		i.setLoopContext(context, loopNode, idx, srcCopy)

		i.env.SetVar(vars[0], key)
		if numVars == 2 {
			value, _ := radMap.Get(key)
			i.env.SetVar(vars[1], value)
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

func (i *Interpreter) runBlock(stmtNodes []rl.Node) EvalResult {
	var res EvalResult
	for _, stmtNode := range stmtNodes {
		res = i.eval(stmtNode)
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

// evaluatePathSegment evaluates a single path segment (field, index, or slice) on a value.
func (i *Interpreter) evaluatePathSegment(rootNode rl.Node, seg rl.PathSegment, val RadValue) RadValue {
	if seg.Field != nil {
		// Dot access: a.b
		fieldNode := rl.NewIdentifier(seg.Span(), *seg.Field)
		key := newRadValueStr(*seg.Field)
		if m, ok := val.TryGetMap(); ok {
			// Use unquoted field name in "Key not found" errors
			result, exists := m.Get(key)
			if !exists {
				errVal := newRadValue(i, fieldNode, NewErrorStrf("Key not found: %s", *seg.Field).SetCode(rl.ErrKeyNotFound))
				i.NewRadPanic(fieldNode, errVal).Panic()
			}
			return result
		}
		return val.Index(i, fieldNode, key)
	}

	if seg.IsSlice {
		// Slice access: a[start:end]
		return i.evalSlice(rootNode, seg, val)
	}

	// Bracket access: a[expr]
	// Check if it's a UFCS call (e.g. mylist.sort(), not mylist[clamp(1, 0, 4)])
	if seg.IsUFCS {
		call := seg.Index.(*rl.Call)
		ufcsArg := &PosArg{
			node:  rootNode,
			value: val,
		}
		return i.callFunction(call, ufcsArg)
	}

	key := i.eval(seg.Index).Val
	return val.Index(i, seg.Index, key)
}

func (i *Interpreter) assignRightsToLefts(targets []rl.Node, values []rl.Node, destructure bool, updateEnclosing bool) EvalResult {
	if destructure {
		if len(values) == 1 {
			if jp, ok := values[0].(*rl.JsonPath); ok {
				i.registerJsonFieldVar(targets[0], jp)
			} else {
				val := i.eval(values[0]).Val
				list, ok := val.TryGetList()
				if ok {
					for idx, target := range targets {
						if len(list.Values) > idx {
							i.doVarPathAssign(target, list.Values[idx], updateEnclosing)
						} else {
							i.doVarPathAssign(target, RAD_NULL_VAL, updateEnclosing)
						}
					}
					return VoidNormal
				} else {
					i.doVarPathAssign(targets[0], val, updateEnclosing)
				}
			}

			for _, target := range targets[1:] {
				i.doVarPathAssign(target, RAD_NULL_VAL, updateEnclosing)
			}
			return VoidNormal
		}

		for idx, target := range targets {
			if len(values) > idx {
				if jp, ok := values[idx].(*rl.JsonPath); ok {
					i.registerJsonFieldVar(target, jp)
				} else {
					val := i.eval(values[idx]).Val
					i.doVarPathAssign(target, val, updateEnclosing)
				}
			} else {
				i.doVarPathAssign(target, RAD_NULL_VAL, updateEnclosing)
			}
		}
		return VoidNormal
	}

	// not destructuring, means exactly 1 target

	if len(values) == 1 {
		if jp, ok := values[0].(*rl.JsonPath); ok {
			i.registerJsonFieldVar(targets[0], jp)
			return VoidNormal
		}

		res := i.eval(values[0])
		if res.Ctrl != CtrlNormal {
			return res
		}
		if res.Val == VOID_SENTINEL {
			i.emitError(rl.ErrVoidValue, values[0], "Cannot assign to a void value")
		}
		i.doVarPathAssign(targets[0], res.Val, updateEnclosing)
		return VoidNormal
	}

	// not destructuring (so 1 target) & not 1 value;
	// means at least 2 values -> pack into list and assign to 1 target

	list := NewRadList()
	for _, value := range values {
		val := i.eval(value).Val
		list.Append(val)
	}
	i.doVarPathAssign(targets[0], newRadValueList(list), updateEnclosing)
	return VoidNormal
}

// registerJsonFieldVar creates and registers a JsonFieldVar from a json path expression.
// The target node provides the variable name; the json path provides the data extraction path.
func (i *Interpreter) registerJsonFieldVar(target rl.Node, jp *rl.JsonPath) {
	// Json field vars must be assigned to plain identifiers, not indexed paths.
	ident, isIdent := target.(*rl.Identifier)
	if !isIdent {
		i.emitError(rl.ErrInvalidSyntax, target, "Json paths must be defined to plain identifiers")
	}
	name := ident.Name

	segments := make([]JsonPathSegment, 0, len(jp.Segments))
	for _, seg := range jp.Segments {
		var idxSegments []JsonPathSegmentIdx
		for _, idx := range seg.Indexes {
			if idx.Expr == nil {
				// wildcard []
				idxSegments = append(idxSegments, JsonPathSegmentIdx{Span: idx.Span})
			} else {
				val := i.eval(idx.Expr).Val
				val.RequireType(i, idx.Expr, fmt.Sprintf("Json path indexes must be ints, was %s", TypeAsString(val)), rl.RadIntT)
				idxSegments = append(idxSegments, JsonPathSegmentIdx{Span: idx.Span, Idx: &val})
			}
		}
		segments = append(segments, JsonPathSegment{
			Identifier:  seg.Key,
			SegmentSpan: seg.KeySpan,
			IdxSegments: idxSegments,
		})
	}

	span := jp.Span()
	fieldVar := NewJsonFieldVar(i, name, span, segments)
	i.env.SetJsonFieldVar(fieldVar)
}

func (i *Interpreter) defineCustomNamedFunction(fnDef *rl.FnDef) {
	fn := NewFnFromAST(i, fnDef.Typing, fnDef.Body, fnDef.IsBlock, &fnDef.DefSpan)
	i.env.SetVar(fnDef.Name, newRadValueFn(fn))
}

func (i *Interpreter) GetSrc() string {
	if i.tmpSrc != nil {
		return *i.tmpSrc
	}
	return i.sd.Src
}

// GetScriptName returns the name/path of the current script.
func (i *Interpreter) GetScriptName() string {
	return i.sd.ScriptName
}

func (i *Interpreter) GetSrcForSpan(span rl.Span) string {
	return i.GetSrc()[span.StartByte:span.EndByte]
}

// evalSlice evaluates a slice operation (a[start:end]) on a value.
func (i *Interpreter) evalSlice(contextNode rl.Node, seg rl.PathSegment, val RadValue) RadValue {
	switch coerced := val.Val.(type) {
	case RadString:
		length := coerced.Len()
		start, end := i.resolveSliceBounds(seg, length)
		return newRadValues(i, contextNode, NewRadString(string(coerced.Runes()[start:end])))
	case *RadList:
		length := coerced.Len()
		start, end := i.resolveSliceBounds(seg, length)
		sliced := NewRadList()
		sliced.Values = coerced.Values[start:end]
		return newRadValues(i, contextNode, sliced)
	default:
		i.emitErrorf(rl.ErrTypeMismatch, contextNode, "Cannot slice a %s", TypeAsString(val))
		panic(UNREACHABLE)
	}
}

func (i *Interpreter) resolveSliceBounds(seg rl.PathSegment, length int64) (int64, int64) {
	start := int64(0)
	end := length

	if seg.Start != nil {
		start = i.eval(seg.Start).Val.RequireInt(i, seg.Start)
		start = CalculateCorrectedIndex(start, length, true)
	}

	if seg.End != nil {
		end = i.eval(seg.End).Val.RequireInt(i, seg.End)
		end = CalculateCorrectedIndex(end, length, true)
	}

	if start > end {
		start = end
	}

	return start, end
}

// todo this is somewhat hacky, not a fan. only use when you're extremely sure fn won't panic
func (i *Interpreter) WithTmpSrc(tmpSrc string, fn func()) {
	i.tmpSrc = &tmpSrc
	defer func() {
		i.tmpSrc = nil
	}()
	fn()
}

// executeSwitchCaseAlt handles a switch case alternative (expr or block).
func (i *Interpreter) executeSwitchCaseAlt(alt rl.Node) EvalResult {
	switch n := alt.(type) {
	case *rl.SwitchCaseExpr:
		return NormalVal(i.evalValues(n, n.Values))
	case *rl.SwitchCaseBlock:
		res := i.runBlock(n.Stmts)
		switch res.Ctrl {
		case CtrlNormal, CtrlBreak, CtrlContinue, CtrlReturn:
			return res
		case CtrlYield:
			return NormalVal(res.Val)
		}
	default:
		i.emitErrorf(rl.ErrInternalBug, alt, "Bug: Unsupported switch case alt type: %T", alt)
	}
	return VoidNormal
}

// handlePanicRecovery handles panic recovery with RadPanic-aware error reporting.
// Should be called from a deferred function with the result of recover().
// fallbackNode is used for error reporting when the panic is not a RadPanic.
// msgArgs are optional context values to include in the error message before the panic value.
func (i *Interpreter) handlePanicRecovery(r interface{}, fallbackNode rl.Node, msgArgs ...interface{}) {
	if r == nil {
		return
	}

	// RadPanic is expected - it's how Rad propagates user-facing errors
	if radPanic, ok := r.(*RadPanic); ok {
		err := radPanic.Err()
		msg := err.Msg().Plain()
		code := rl.ErrGenericRuntime
		if !com.IsBlank(string(err.Code)) {
			code = err.Code
		}
		// Use err.Span if available, otherwise fall back to fallbackNode's span
		if err.Span != nil {
			i.emitErrorSpan(code, err.Span, msg)
		} else {
			i.emitError(code, fallbackNode, msg)
		}
		return
	}

	// Non-RadPanic means an internal bug - this shouldn't happen
	// Skip in tests since test framework may use panics for control flow (e.g., exit simulation)
	if !IsTest {
		i.emitInternalBug(r, fallbackNode, msgArgs...)
	}
}

// emitInternalBug reports an internal Rad bug to the user.
// This should only be called for unexpected panics that are NOT RadPanic.
func (i *Interpreter) emitInternalBug(panicValue interface{}, fallbackNode rl.Node, msgArgs ...interface{}) {
	var msgBuilder strings.Builder
	msgBuilder.WriteString("This is a bug in Rad. Please report it at:\n")
	msgBuilder.WriteString("  https://github.com/amterp/rad/issues\n\n")

	msgBuilder.WriteString("Panic: ")
	for _, arg := range msgArgs {
		msgBuilder.WriteString(fmt.Sprintf("%v ", arg))
	}
	msgBuilder.WriteString(fmt.Sprintf("%v\n\n", panicValue))

	if !IsTest {
		// Include Go stack trace for debugging (skip in tests for deterministic output)
		msgBuilder.WriteString("Go stack trace:\n")
		msgBuilder.WriteString(string(debug.Stack()))
	}

	i.emitError(rl.ErrInternalBug, fallbackNode, msgBuilder.String())
}

// withCatch wraps body execution with panic catching. If catch is nil, just executes body.
// On RadPanic, calls onErr callback to handle the error (assign variables, run catch block, etc.).
// Propagates control flow (return/break/continue/yield) from the catch block.
// Re-panics non-RadPanic errors to preserve Go's panic semantics (e.g., runtime errors, bugs).
func (i *Interpreter) withCatch(
	catch *rl.CatchBlock,
	onErr func(rp *RadPanic) EvalResult,
	body func() EvalResult,
) (out EvalResult) {
	if catch == nil {
		return body()
	}

	defer func() {
		if r := recover(); r != nil {
			if rp, ok := r.(*RadPanic); ok {
				out = onErr(rp)
			} else {
				panic(r)
			}
		}
	}()

	return body()
}
