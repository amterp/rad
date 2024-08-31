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
	VisitEmptyStmt(Empty)
	VisitExprStmtStmt(ExprStmt)
	VisitPrimaryAssignStmt(PrimaryAssign)
	VisitFileHeaderStmt(FileHeader)
	VisitArgBlockStmt(ArgBlock)
	VisitRadBlockStmt(RadBlock)
	VisitJsonPathAssignStmt(JsonPathAssign)
}
type Empty struct {
}

func (e Empty) Accept(visitor StmtVisitor) {
	visitor.VisitEmptyStmt(e)
}
func (e Empty) String() string {
	return "Empty()"
}

type ExprStmt struct {
	Expression Expr
}

func (e ExprStmt) Accept(visitor StmtVisitor) {
	visitor.VisitExprStmtStmt(e)
}
func (e ExprStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Expression: %v", e.Expression))
	return fmt.Sprintf("ExprStmt(%s)", strings.Join(parts, ", "))
}

type PrimaryAssign struct {
	Name        Token
	Initializer Expr
}

func (e PrimaryAssign) Accept(visitor StmtVisitor) {
	visitor.VisitPrimaryAssignStmt(e)
}
func (e PrimaryAssign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Name: %v", e.Name))
	parts = append(parts, fmt.Sprintf("Initializer: %v", e.Initializer))
	return fmt.Sprintf("PrimaryAssign(%s)", strings.Join(parts, ", "))
}

type FileHeader struct {
	FileHeaderToken Token
}

func (e FileHeader) Accept(visitor StmtVisitor) {
	visitor.VisitFileHeaderStmt(e)
}
func (e FileHeader) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("FileHeaderToken: %v", e.FileHeaderToken))
	return fmt.Sprintf("FileHeader(%s)", strings.Join(parts, ", "))
}

type ArgBlock struct {
	ArgsKeyword Token
	Stmts       []ArgStmt
}

func (e ArgBlock) Accept(visitor StmtVisitor) {
	visitor.VisitArgBlockStmt(e)
}
func (e ArgBlock) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("ArgsKeyword: %v", e.ArgsKeyword))
	parts = append(parts, fmt.Sprintf("Stmts: %v", e.Stmts))
	return fmt.Sprintf("ArgBlock(%s)", strings.Join(parts, ", "))
}

type RadBlock struct {
	RadKeyword Token
	Url        *Expr
	RadStmts   []RadStmt
}

func (e RadBlock) Accept(visitor StmtVisitor) {
	visitor.VisitRadBlockStmt(e)
}
func (e RadBlock) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("RadKeyword: %v", e.RadKeyword))
	parts = append(parts, fmt.Sprintf("Url: %v", e.Url))
	parts = append(parts, fmt.Sprintf("RadStmts: %v", e.RadStmts))
	return fmt.Sprintf("RadBlock(%s)", strings.Join(parts, ", "))
}

type JsonPathAssign struct {
	Identifier Token
	Elements   []JsonPathElement
}

func (e JsonPathAssign) Accept(visitor StmtVisitor) {
	visitor.VisitJsonPathAssignStmt(e)
}
func (e JsonPathAssign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifier: %v", e.Identifier))
	parts = append(parts, fmt.Sprintf("Elements: %v", e.Elements))
	return fmt.Sprintf("JsonPathAssign(%s)", strings.Join(parts, ", "))
}
