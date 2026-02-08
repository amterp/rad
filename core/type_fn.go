package core

import (
	"fmt"
	"sort"

	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadFn struct {
	BuiltInFunc *BuiltInFunc // if this represents a built-in function
	// below for non-built-in functions
	ReprNode *ts.Node // representative node (can point at this for errors)
	Typing   *rl.TypingFnT
	Stmts    []ts.Node
	IsBlock  bool // if this is a block function or expr. Block functions can only return with a 'return' stmt.
	Env      *Env // for closures
}

func NewFn(i *Interpreter, fnNode *ts.Node) RadFn {
	typing := rl.NewTypingFnT(fnNode, i.GetSrc())
	stmts := rl.GetChildren(fnNode, rl.F_STMT, i.cursor)
	reprNode := fnNode
	isBlock := rl.GetChild(fnNode, rl.F_BLOCK_COLON) != nil

	if isBlock {
		reprNode = rl.GetChild(fnNode, rl.F_KEYWORD)
	}

	return RadFn{
		ReprNode: reprNode,
		Typing:   typing,
		Stmts:    stmts,
		IsBlock:  isBlock,
		Env:      i.env,
	}
}

func NewBuiltIn(inFunc BuiltInFunc) RadFn {
	return RadFn{
		BuiltInFunc: &inFunc,
	}
}

func (fn RadFn) Name() string {
	if fn.BuiltInFunc != nil {
		return fn.BuiltInFunc.Name
	}
	return fn.Typing.Name
}

func (fn RadFn) IsBuiltIn() bool {
	return fn.BuiltInFunc != nil
}

// ParamCount returns the number of parameters this function accepts.
// Returns 0 if typing information is unavailable.
func (fn RadFn) ParamCount() int {
	if fn.BuiltInFunc != nil {
		if fn.BuiltInFunc.Signature != nil && fn.BuiltInFunc.Signature.Typing != nil {
			return len(fn.BuiltInFunc.Signature.Typing.Params)
		}
		return 0
	}
	if fn.Typing == nil {
		return 0
	}
	return len(fn.Typing.Params)
}

func (fn RadFn) Execute(f FuncInvocation) (out RadValue) {
	i := f.i

	var typing *rl.TypingFnT
	if fn.BuiltInFunc == nil {
		typing = fn.Typing
	} else {
		sig := fn.BuiltInFunc.Signature
		if sig != nil {
			typing = fn.BuiltInFunc.Signature.Typing
		}
	}

	out = VOID_SENTINEL
	i.runWithChildEnv(func() {
		// todo the following checking logic should be in IsCompatibleWith for TypingFnT

		seen := make(map[string]bool)

		params := typing.Params
		paramCount := len(params)
		// handle positional and variadic args
		for idx, arg := range f.args {
			// if we've consumed all fixed params
			if idx >= paramCount {
				// check if last param is variadic
				if paramCount > 0 && params[paramCount-1].IsVariadic {
					varArg := params[paramCount-1]
					// collect all remaining args into a slice
					radList := NewRadList()
					for j := paramCount - 1; j < len(f.args); j++ {
						elem := f.args[j].value
						typeCheck(i, varArg.Type, arg.node, elem)
						radList.Append(elem)
					}
					i.env.SetVar(varArg.Name, newRadValueList(radList))
					seen[varArg.Name] = true
					break
				} else {
					i.emitErrorf(rl.ErrWrongArgCount, f.callNode,
						"Expected at most %d args, but was invoked with %d", paramCount, len(f.args))
				}
			}

			param := params[idx]

			if param.IsVariadic {
				radList := NewRadList()

				for j := idx; j < len(f.args); j++ {
					elem := f.args[j].value
					typeCheck(i, param.Type, f.args[j].node, elem)
					radList.Append(elem)
				}

				i.env.SetVar(param.Name, newRadValueList(radList))
				seen[param.Name] = true
				break
			}

			if param.NamedOnly {
				i.emitError(rl.ErrWrongArgCount, arg.node, "Too many positional args, remaining args are named-only")
			}

			// normal type check
			if param.Type != nil {
				typeCheck(i, param.Type, arg.node, arg.value)
			}

			seen[param.Name] = true
			i.env.SetVar(param.Name, arg.value)
		}

		// named args (unchanged)
		byName := typing.ByName()

		names := make([]string, 0, len(f.namedArgs))
		for name := range f.namedArgs {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			arg := f.namedArgs[name]

			param, ok := byName[name]
			if !ok {
				i.emitErrorf(rl.ErrInvalidArgType, arg.nameNode, "Unknown named argument '%s'", name)
			}

			if param.AnonymousOnly() {
				i.emitErrorf(rl.ErrInvalidArgType, arg.nameNode,
					"Argument '%s' cannot be passed as named arg, only positionally", name)
			}

			if seen[name] {
				i.emitErrorf(rl.ErrInvalidArgType, arg.nameNode, "Argument '%s' already specified", name)
			}

			if param.Type != nil {
				typeCheck(f.i, param.Type, arg.valueNode, arg.value)
			}

			seen[param.Name] = true
			i.env.SetVar(param.Name, arg.value)
		}

		// check for missing required args, handle defaults or null
		for _, param := range params {
			if seen[param.Name] {
				continue
			}

			if param.Default != nil {
				i.WithTmpSrc(param.Default.Src, func() {
					defaultVal := i.eval(param.Default.Node).Val
					if param.Type != nil {
						typeCheck(i, param.Type, param.Default.Node, defaultVal)
					}
					i.env.SetVar(param.Name, defaultVal)
				})
				continue
			}

			if param.IsVariadic {
				i.env.SetVar(param.Name, newRadValueList(NewRadList()))
				continue
			}

			if param.IsOptional || (param.Type != nil && (*param.Type).IsCompatibleWith(rl.NewNullSubject())) {
				i.env.SetVar(param.Name, RAD_NULL_VAL)
				continue
			}

			// todo below exposes _vars not meant to be. Perhaps just say # of missing args?
			i.emitErrorf(rl.ErrWrongArgCount, f.callNode, "Missing required argument '%s'", param.Name)
		}

		// todo this should be more shared between the two branches?
		if fn.BuiltInFunc == nil {
			// Push call frame for user-defined functions
			var callSite, defSite *Span
			if f.callNode != nil {
				cs := NewSpanFromNode(f.callNode, i.GetScriptName())
				callSite = &cs
			}
			if fn.ReprNode != nil {
				ds := NewSpanFromNode(fn.ReprNode, i.GetScriptName())
				defSite = &ds
			}
			fnName := fn.Name()
			if fnName == "" {
				fnName = "<anonymous>"
			}
			i.pushCallFrame(fnName, callSite, defSite)
			defer i.popCallFrame()

			res := i.runBlock(fn.Stmts)
			typeCheck(i, typing.ReturnT, f.callNode, res.Val)
			if fn.IsBlock {
				if res.Ctrl == CtrlReturn {
					out = res.Val
				}
			} else {
				out = res.Val
			}
		} else {
			// Built-in functions don't get call frames in the stack trace.
			// Users care about their own code, not Rad internals.
			out = fn.BuiltInFunc.Execute(f)
		}
	})

	if out.IsError() {
		if f.panicIfError {
			i.NewRadPanic(f.callNode, out).Panic()
		} else {
			// we'll let this error propagate, so let's clear its node for error pointing, if it has one
			err := out.RequireError(f.i, f.callNode)
			err.SetNode(nil)
		}
	}

	return
}

func typeCheck(i *Interpreter, typing *rl.TypingT, node *ts.Node, val RadValue) {
	if typing == nil {
		return
	}

	isCompat := (*typing).IsCompatibleWith(val.ToCompatSubject(i))
	if !isCompat {
		if val == VOID_SENTINEL {
			i.emitErrorf(rl.ErrVoidValue, node, "Expected '%s', but got void value", (*typing).Name())
			return
		}

		i.emitErrorf(rl.ErrTypeMismatch, node, "Value '%s' (%s) is not compatible with expected type '%s'",
			ToPrintable(val), val.Type().AsString(), (*typing).Name())
	}
}

func (fn RadFn) ToString() string {
	// todo can we include var name if possible?
	return fmt.Sprintf("<fn>") // TODO should add details from signature
}
