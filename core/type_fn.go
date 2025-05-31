package core

import (
	"fmt"
	"strings"

	"github.com/amterp/rad/rts/rsl"
	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslFn struct {
	BuiltInFunc *BuiltInFunc // if this represents a built-in function
	// below for non-built-in functions
	Params     []string
	ReprNode   *ts.Node  // representative node (can point at this for errors)
	Exprs      []ts.Node // for returning lambdas
	Stmt       *ts.Node  // for stmt lambdas
	Body       []ts.Node // for fn blocks
	ReturnStmt *ts.Node  // for fn blocks
	Env        *Env      // for closures
}

func NewLambda(i *Interpreter, lambdaNode *ts.Node) RslFn {
	paramNodes := i.getChildren(lambdaNode, rsl.F_PARAM)
	params := lo.Map(paramNodes, func(n ts.Node, _ int) string { return GetSrc(i.sd.Src, &n) })
	return RslFn{
		Params:   params,
		ReprNode: lambdaNode,
		Exprs:    i.getChildren(lambdaNode, rsl.F_EXPR),
		Stmt:     i.getChild(lambdaNode, rsl.F_STMT),
		Env:      i.env,
	}
}

func NewFnBlock(i *Interpreter, fnBlockNode *ts.Node) RslFn {
	keywordNode := i.getChild(fnBlockNode, rsl.F_KEYWORD)
	paramNodes := i.getChildren(fnBlockNode, rsl.F_PARAM)
	params := lo.Map(paramNodes, func(n ts.Node, _ int) string { return GetSrc(i.sd.Src, &n) })
	stmtNodes := i.getChildren(fnBlockNode, rsl.F_STMT)
	returnNode := i.getChild(fnBlockNode, rsl.F_RETURN_STMT)
	return RslFn{
		Params:     params,
		ReprNode:   keywordNode,
		Body:       stmtNodes,
		ReturnStmt: returnNode,
		Env:        i.env,
	}
}

func NewBuiltIn(inFunc BuiltInFunc) RslFn {
	return RslFn{
		BuiltInFunc: &inFunc,
	}
}

func (fn RslFn) IsLambda() bool {
	return len(fn.Exprs) > 0 || fn.Stmt != nil
}

func (fn RslFn) Execute(f FuncInvocationArgs) []RslValue {
	if fn.BuiltInFunc != nil {
		assertMinNumPosArgs(f, fn.BuiltInFunc)
		fn.BuiltInFunc.PosArgValidator.validate(f, fn.BuiltInFunc)
		assertAllowedNamedArgs(f, fn.BuiltInFunc)
		assertCorrectNumReturnValues(f, fn.BuiltInFunc)
		return fn.BuiltInFunc.Execute(f)
	}

	i := f.i
	output := make([]RslValue, 0)
	i.runWithChildEnv(func() {
		args := f.args
		// custom funcs don't support namedArgs, so we ignore them. Parser doesn't allow them anyway.
		for idx, arg := range args {
			i.env.SetVar(fn.Params[idx], arg.value)
		}

		if fn.IsLambda() {
			if len(fn.Exprs) > 0 {
				for _, exprNode := range fn.Exprs {
					val := i.evaluate(&exprNode, NO_NUM_RETURN_VALUES_CONSTRAINT)
					output = append(output, val...) // todo dunno about this splatter
				}
			} else {
				i.recursivelyRun(fn.Stmt)
			}
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
