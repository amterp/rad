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
	Params     []string
	ReprNode   *ts.Node  // representative node (can point at this for errors)
	Exprs      []ts.Node // for returning lambdas
	Stmt       *ts.Node  // for stmt lambdas
	Body       []ts.Node // for fn blocks
	ReturnStmt *ts.Node  // for fn blocks
	Env        *Env      // for closures
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
		Exprs:    i.getChildren(lambdaNode, rl.F_EXPR),
		Stmt:     i.getChild(lambdaNode, rl.F_STMT),
		Env:      i.env,
	}
}

func NewLambdaBlock(i *Interpreter, lambdaNode *ts.Node) RadFn {
	keywordNode := i.getChild(lambdaNode, rl.F_KEYWORD)
	params := resolveParamNames(i, lambdaNode)
	stmtNodes := i.getChildren(lambdaNode, rl.F_STMT)
	returnNode := i.getChild(lambdaNode, rl.F_RETURN_STMT)
	return RadFn{
		Params:     params,
		ReprNode:   keywordNode,
		Body:       stmtNodes,
		ReturnStmt: returnNode,
		Env:        i.env,
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

func (fn RadFn) IsLambda() bool {
	return len(fn.Exprs) > 0 || fn.Stmt != nil
}

func (fn RadFn) Execute(f FuncInvocationArgs) []RadValue {
	if fn.BuiltInFunc != nil {
		assertMinNumPosArgs(f, fn.BuiltInFunc)
		fn.BuiltInFunc.PosArgValidator.validate(f, fn.BuiltInFunc)
		assertAllowedNamedArgs(f, fn.BuiltInFunc)
		assertCorrectNumReturnValues(f, fn.BuiltInFunc)
		return fn.BuiltInFunc.Execute(f)
	}

	i := f.i
	output := make([]RadValue, 0)
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
				valueNodes := i.getChildren(fn.ReturnStmt, rl.F_VALUE)
				for _, valueNode := range valueNodes {
					val := i.evaluate(&valueNode, NO_NUM_RETURN_VALUES_CONSTRAINT)
					output = append(output, val...) // todo this is probably bad and inconsistent with e.g. switch yields
				}
			}
		}
	})
	return output
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
