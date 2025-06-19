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

func NewLambda(i *Interpreter, fnNode *ts.Node) RadFn {
	typing := rl.NewTypingFnT(fnNode, i.sd.Src)
	stmts := rl.GetChildren(fnNode, rl.F_STMT)
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

func (fn RadFn) IsBuiltIn() bool {
	return fn.BuiltInFunc != nil
}

func (fn RadFn) Execute(f FuncInvocationArgs) (out RadValue) {
	i := f.i
	if fn.BuiltInFunc == nil {
		out = VOID_SENTINEL
		i.runWithChildEnv(func() {
			// todo the following checking logic should be in IsCompatibleWith for TypingFnT

			seen := make(map[string]bool)

			for idx, arg := range f.args {
				if idx >= len(fn.Typing.Params) {
					i.errorf(f.callNode,
						"Expected at most %d args, but was invoked with %d", len(fn.Typing.Params), len(f.args))
				}

				param := fn.Typing.Params[idx]
				if param.NamedOnly {
					i.errorf(arg.node, "Too many positional args, remaining args are named-only.")
				}

				if param.Type != nil {
					fn.typeCheck(f.i, param.Type, arg.node, arg.value)
				}
				seen[param.Name] = true
				i.env.SetVar(param.Name, arg.value)
			}

			byName := fn.Typing.ByName()

			names := make([]string, 0, len(f.namedArgs))
			for name := range f.namedArgs {
				names = append(names, name)
			}
			sort.Strings(names) // ascending lexicographic order

			for _, name := range names {
				arg := f.namedArgs[name]

				param, ok := byName[name]
				if !ok {
					i.errorf(arg.nameNode, "Unknown named argument '%s'", name)
				}

				if param.AnonymousOnly() {
					i.errorf(arg.nameNode,
						"Argument '%s' cannot be passed as named arg, only positionally.", name)
				}

				if seen[name] {
					i.errorf(arg.nameNode, "Argument '%s' already specified.", name)
				}

				if param.Type != nil {
					fn.typeCheck(f.i, param.Type, arg.valueNode, arg.value)
				}

				seen[param.Name] = true
				i.env.SetVar(param.Name, arg.value)
			}

			// check for missing required args, or define optional/default args
			for _, param := range fn.Typing.Params {
				_, seenParam := seen[param.Name]

				if seenParam {
					continue
				}

				if param.Default != nil {
					defaultVal := i.eval(param.Default).Val
					if param.Type != nil {
						fn.typeCheck(i, param.Type, param.Default, defaultVal)
					}
					i.env.SetVar(param.Name, defaultVal)
					continue
				}

				if param.IsOptional || (param.Type != nil && (*param.Type).IsCompatibleWith(rl.NewNullSubject())) {
					i.env.SetVar(param.Name, RAD_NULL_VAL)
					continue
				}

				i.errorf(f.callNode, "Missing required argument '%s'", param.Name)
			}

			res := i.runBlock(fn.Stmts)
			fn.typeCheck(i, fn.Typing.ReturnT, f.callNode, res.Val)
			if fn.IsBlock {
				if res.Ctrl == CtrlReturn {
					out = res.Val
				}
			} else {
				out = res.Val
			}
		})
	} else {
		assertMinNumPosArgs(f, fn.BuiltInFunc)
		fn.BuiltInFunc.PosArgValidator.validate(f, fn.BuiltInFunc)
		assertAllowedNamedArgs(f, fn.BuiltInFunc)
		out = fn.BuiltInFunc.Execute(f)
	}

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

func (fn RadFn) typeCheck(i *Interpreter, typing *rl.TypingT, node *ts.Node, val RadValue) {
	if typing == nil {
		return
	}

	isCompat := (*typing).IsCompatibleWith(val.ToCompatSubject())
	if !isCompat {
		if val == VOID_SENTINEL {
			i.errorf(node, "Expected '%s', but got void value.", (*typing).Name())
			return
		}

		i.errorf(node, "Value '%s' (%s) is not compatible with expected type '%s'",
			ToPrintable(val), val.Type().AsString(), (*typing).Name())
	}
}

func (fn RadFn) ToString() string {
	// todo can we include var name if possible?
	return fmt.Sprintf("<fn>") // TODO should add details from signature
}
