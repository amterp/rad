// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Accept(visitor StmtVisitor)
}
type StmtVisitor interface {
	VisitExpressionStmt(*Expression)
	VisitPrimaryAssignStmt(*PrimaryAssign)
	VisitFileHeaderStmt(*FileHeader)
	VisitEmptyStmt(*Empty)
	VisitArgBlockStmt(*ArgBlock)
	VisitJsonPathAssignStmt(*JsonPathAssign)
}
type Expression struct {
	expression Expr
}

func (e *Expression) Accept(visitor StmtVisitor) {
	visitor.VisitExpressionStmt(e)
}
func (e *Expression) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("expression: %v", e.expression))
	return fmt.Sprintf("Expression(%s)", strings.Join(parts, ", "))
}

type PrimaryAssign struct {
	name        Token
	initializer Expr
}

func (e *PrimaryAssign) Accept(visitor StmtVisitor) {
	visitor.VisitPrimaryAssignStmt(e)
}
func (e *PrimaryAssign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("name: %v", e.name))
	parts = append(parts, fmt.Sprintf("initializer: %v", e.initializer))
	return fmt.Sprintf("PrimaryAssign(%s)", strings.Join(parts, ", "))
}

type FileHeader struct {
	fileHeaderToken Token
}

func (e *FileHeader) Accept(visitor StmtVisitor) {
	visitor.VisitFileHeaderStmt(e)
}
func (e *FileHeader) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("fileHeaderToken: %v", e.fileHeaderToken))
	return fmt.Sprintf("FileHeader(%s)", strings.Join(parts, ", "))
}

type Empty struct {
}

func (e *Empty) Accept(visitor StmtVisitor) {
	visitor.VisitEmptyStmt(e)
}
func (e *Empty) String() string {
	return "Empty()"
}

type ArgBlock struct {
	argsKeyword Token
	argStmts    []ArgStmt
}

func (e *ArgBlock) Accept(visitor StmtVisitor) {
	visitor.VisitArgBlockStmt(e)
}
func (e *ArgBlock) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("argsKeyword: %v", e.argsKeyword))
	parts = append(parts, fmt.Sprintf("argStmts: %v", e.argStmts))
	return fmt.Sprintf("ArgBlock(%s)", strings.Join(parts, ", "))
}

type JsonPathAssign struct {
	identifier Token
	elements   []JsonPathElement
}

func (e *JsonPathAssign) Accept(visitor StmtVisitor) {
	visitor.VisitJsonPathAssignStmt(e)
}
func (e *JsonPathAssign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("identifier: %v", e.identifier))
	parts = append(parts, fmt.Sprintf("elements: %v", e.elements))
	return fmt.Sprintf("JsonPathAssign(%s)", strings.Join(parts, ", "))
}
