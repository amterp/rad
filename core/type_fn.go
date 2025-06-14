package core

import (
	"fmt"
	"strings"

	"github.com/amterp/rad/rts/rl"
	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadFn struct {
	BuiltInFunc *BuiltInFunc // if this represents a built-in function
	// below for non-built-in functions
	Params   []string
	ReprNode *ts.Node  // representative node (can point at this for errors)
	Expr     *ts.Node  // for returning lambdas
	Stmt     *ts.Node  // for stmt lambdas
	Body     []ts.Node // for fn blocks
	Env      *Env      // for closures
}

func NewLambda(i *Interpreter, lambdaNode *ts.Node) RadFn {
	blockColon := i.getChild(lambdaNode, rl.F_BLOCK_COLON)
	if blockColon == nil {
		return NewLambdaOneLiner(i, lambdaNode)
	} else {
		return NewLambdaBlock(i, lambdaNode)
	}
}

func NewLambdaOneLiner(i *Interpreter, lambdaNode *ts.Node) RadFn {
	params := resolveParamNames(i, lambdaNode)
	return RadFn{
		Params:   params,
		ReprNode: lambdaNode,
		Expr:     i.getChild(lambdaNode, rl.F_EXPR),
		Stmt:     i.getChild(lambdaNode, rl.F_STMT),
		Env:      i.env,
	}
}

func NewLambdaBlock(i *Interpreter, lambdaNode *ts.Node) RadFn {
	keywordNode := i.getChild(lambdaNode, rl.F_KEYWORD)
	params := resolveParamNames(i, lambdaNode)
	stmtNodes := i.getChildren(lambdaNode, rl.F_STMT)
	return RadFn{
		Params:   params,
		ReprNode: keywordNode,
		Body:     stmtNodes,
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

func (fn RadFn) IsLambda() bool { // todo not accurate, can have named func with this
	return fn.Expr != nil || fn.Stmt != nil
}

func (fn RadFn) Execute(f FuncInvocationArgs) (out RadValue) {
	if fn.BuiltInFunc == nil {
		if len(f.args) != len(fn.Params) {
			f.i.errorf(f.callNode, "Expected %d args, but was invoked with %d", len(fn.Params), len(f.args))
		}

		i := f.i
		out = VOID_SENTINEL
		i.runWithChildEnv(func() {
			args := f.args
			// custom funcs don't support namedArgs, so we ignore them. Parser doesn't allow them anyway.
			// todo ^^ no longer accurate
			for idx, arg := range args {
				i.env.SetVar(fn.Params[idx], arg.value)
			}

			if fn.IsLambda() {
				if fn.Expr != nil {
					out = i.eval(fn.Expr).Val
				}
				if fn.Stmt != nil {
					i.eval(fn.Stmt)
				}
			} else {
				out = i.runBlock(fn.Body).Val
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
			f.i.NewRadPanic(f.callNode, out).Panic()
		} else {
			// we'll let this error propagate, so let's clear its node for error pointing, if it has one
			err := out.RequireError(f.i, f.callNode)
			err.SetNode(nil)
		}
	}

	return
}

func (fn RadFn) ToString() string {
	// todo can we include var name if possible?
	return fmt.Sprintf("<fn (%s)>", strings.Join(fn.Params, ", "))
}

func resolveParamNames(i *Interpreter, lambdaNode *ts.Node) []string {
	paramNodes := i.getChildren(lambdaNode, rl.F_PARAM)
	return lo.Map(paramNodes, func(n ts.Node, _ int) string {
		nameNode := i.getChild(&n, rl.F_NAME)
		return GetSrc(i.sd.Src, nameNode)
	})
}
