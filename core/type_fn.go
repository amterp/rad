package core

import (
	"fmt"
	"github.com/amterp/rts/rsl"
	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
	"strings"
)

type RslFn struct {
	Params     []string
	Expr       *ts.Node  // for lambdas
	Body       []ts.Node // for fn blocks
	ReturnStmt *ts.Node  // for fn blocks
	Env        *Env      // for closures
	// todo add something in here to store built-in funcs. also need bool for 'IsBuiltIn'.
}

func NewLambda(i *Interpreter, lambdaNode *ts.Node) RslFn {
	paramNodes := i.getChildren(lambdaNode, rsl.F_PARAM)
	params := lo.Map(paramNodes, func(n ts.Node, _ int) string { return GetSrc(i.sd.Src, &n) })
	exprNode := i.getChild(lambdaNode, rsl.F_EXPR)
	return RslFn{
		Params: params,
		Expr:   exprNode,
		Env:    i.env,
	}
}

func NewFnBlock(i *Interpreter, fnBlockNode *ts.Node) RslFn {
	paramNodes := i.getChildren(fnBlockNode, rsl.F_PARAM)
	params := lo.Map(paramNodes, func(n ts.Node, _ int) string { return GetSrc(i.sd.Src, &n) })
	stmtNodes := i.getChildren(fnBlockNode, rsl.F_STMT)
	returnNode := i.getChild(fnBlockNode, rsl.F_RETURN_STMT)
	return RslFn{
		Params:     params,
		Body:       stmtNodes,
		ReturnStmt: returnNode,
		Env:        i.env,
	}
}

func (fn RslFn) IsLambda() bool {
	return fn.Expr != nil
}

// todo will this be re-used for built-in funcs? probably, but we'll fork off early
func (fn RslFn) Execute(f FuncInvocationArgs) []RslValue {
	i := f.i
	output := make([]RslValue, 0)
	i.runWithChildEnv(func() {
		args := f.args
		// custom funcs don't support namedArgs, so we ignore them. Parser doesn't allow them anyway.
		for idx, arg := range args {
			i.env.SetVar(fn.Params[idx], arg.value)
		}

		if fn.IsLambda() {
			output = i.evaluate(fn.Expr, NO_NUM_RETURN_VALUES_CONSTRAINT)
		} else {
			i.runBlock(fn.Body)

			if i.breakingOrContinuing() {
				return
			}

			if f.numExpectedOutputs > 0 && fn.ReturnStmt == nil {
				i.errorf(f.callNode, "Expected %d outputs, but function '%s' is missing a return statement.",
					f.numExpectedOutputs, f.funcName)
			}

			if fn.ReturnStmt != nil {
				valueNodes := i.getChildren(fn.ReturnStmt, rsl.F_VALUE)
				for _, valueNode := range valueNodes {
					val := i.evaluate(&valueNode, NO_NUM_RETURN_VALUES_CONSTRAINT)
					output = append(output, val...) // todo this is probably bad and inconsistent with e.g. switch yields
				}
			}
		}
	})
	return output
}

func (fn RslFn) ToString() string {
	// todo can we include var name if possible?
	return fmt.Sprintf("<fn (%s)>", strings.Join(fn.Params, ", "))
}
