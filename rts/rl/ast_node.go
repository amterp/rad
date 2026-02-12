package rl

import "fmt"

// Node is the interface all AST nodes implement.
type Node interface {
	Kind() NodeKind
	Span() Span
	Children() []Node
}

// NodeKind identifies the type of an AST node.
type NodeKind uint16

const (
	// Statements
	NAssign    NodeKind = iota // also handles desugared += -= *= /= ++ --
	NExprStmt                  // expression used as a statement
	NIf                        // if/elif/else
	NSwitch                    // switch/case
	NForLoop                   // for ... in ...
	NWhileLoop                 // while ...
	NShell                     // shell statement ($...)
	NDel                       // del statement
	NDefer                     // defer/errdefer block
	NBreak                     // break
	NContinue                  // continue
	NReturn                    // return [values]
	NYield                     // yield [values]
	NPass                      // pass (no-op)
	NFnDef                     // named function definition

	// Expressions
	NOpBinary   // binary operation (a + b, a and b, etc.)
	NOpUnary    // unary operation (-x, not x)
	NTernary    // condition ? true : false
	NFallback   // left ?? right
	NCatchExpr  // left catch right
	NCall       // function call
	NVarPath    // a.b[c].d - flat segment chain
	NIdentifier // variable name
	NLambda     // anonymous function

	// Literals
	NLitInt    // integer literal
	NLitFloat  // float literal
	NLitBool   // true/false
	NLitNull   // null
	NLitString // string (simple or interpolated)
	NLitList   // [a, b, c]
	NLitMap    // {k: v}

	// Comprehension
	NListComp // [expr for vars in iter if cond]

	// Switch case alternatives
	NSwitchCaseExpr  // single-expression case (case X -> expr)
	NSwitchCaseBlock // multi-statement case (case X:\n  stmts)

	// Rad block internals
	NRadBlock    // rad/request/display block
	NRadField    // field declaration in a rad block
	NRadSort     // sort specifier in a rad block
	NRadFieldMod // field modifier (color, map, filter)
	NRadIf       // conditional in a rad block

	// JSON
	NJsonPath // json[].field path expression

	// Script metadata
	NFileHeader // --- block with description + metadata
	NArgBlock   // args: block
	NArgDecl    // single arg declaration
	NCmdBlock   // command definition

	// Structural
	NSourceFile // root node of a script
)

var nodeKindNames = [...]string{
	NAssign:          "Assign",
	NExprStmt:        "ExprStmt",
	NIf:              "If",
	NSwitch:          "Switch",
	NForLoop:         "ForLoop",
	NWhileLoop:       "WhileLoop",
	NShell:           "Shell",
	NDel:             "Del",
	NDefer:           "Defer",
	NBreak:           "Break",
	NContinue:        "Continue",
	NReturn:          "Return",
	NYield:           "Yield",
	NPass:            "Pass",
	NFnDef:           "FnDef",
	NOpBinary:        "OpBinary",
	NOpUnary:         "OpUnary",
	NTernary:         "Ternary",
	NFallback:        "Fallback",
	NCatchExpr:       "CatchExpr",
	NCall:            "Call",
	NVarPath:         "VarPath",
	NIdentifier:      "Identifier",
	NLambda:          "Lambda",
	NLitInt:          "LitInt",
	NLitFloat:        "LitFloat",
	NLitBool:         "LitBool",
	NLitNull:         "LitNull",
	NLitString:       "LitString",
	NLitList:         "LitList",
	NLitMap:          "LitMap",
	NListComp:        "ListComp",
	NSwitchCaseExpr:  "SwitchCaseExpr",
	NSwitchCaseBlock: "SwitchCaseBlock",
	NRadBlock:        "RadBlock",
	NRadField:        "RadField",
	NRadSort:         "RadSort",
	NRadFieldMod:     "RadFieldMod",
	NRadIf:           "RadIf",
	NJsonPath:        "JsonPath",
	NFileHeader:      "FileHeader",
	NArgBlock:        "ArgBlock",
	NArgDecl:         "ArgDecl",
	NCmdBlock:        "CmdBlock",
	NSourceFile:      "SourceFile",
}

func (k NodeKind) String() string {
	if int(k) < len(nodeKindNames) {
		return nodeKindNames[k]
	}
	return fmt.Sprintf("NodeKind(%d)", k)
}

// Operator identifies a binary or unary operator.
type Operator uint8

const (
	// Arithmetic
	OpAdd Operator = iota
	OpSub
	OpMul
	OpDiv
	OpMod

	// Comparison
	OpEq
	OpNeq
	OpLt
	OpLte
	OpGt
	OpGte

	// Logical
	OpAnd
	OpOr

	// Membership
	OpIn    // in
	OpNotIn // not in

	// Unary
	OpNeg // -x
	OpNot // not x
)

var operatorNames = [...]string{
	OpAdd:   "+",
	OpSub:   "-",
	OpMul:   "*",
	OpDiv:   "/",
	OpMod:   "%",
	OpEq:    "==",
	OpNeq:   "!=",
	OpLt:    "<",
	OpLte:   "<=",
	OpGt:    ">",
	OpGte:   ">=",
	OpAnd:   "and",
	OpOr:    "or",
	OpIn:    "in",
	OpNotIn: "not in",
	OpNeg:   "-",
	OpNot:   "not",
}

func (op Operator) String() string {
	if int(op) < len(operatorNames) {
		return operatorNames[op]
	}
	return fmt.Sprintf("Operator(%d)", op)
}
