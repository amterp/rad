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
	VisitFunctionStmtStmt(FunctionStmt)
	VisitPrimaryAssignStmt(PrimaryAssign)
	VisitFileHeaderStmt(FileHeader)
	VisitArgBlockStmt(ArgBlock)
	VisitRadBlockStmt(RadBlock)
	VisitJsonPathAssignStmt(JsonPathAssign)
	VisitSwitchBlockStmtStmt(SwitchBlockStmt)
	VisitSwitchAssignmentStmt(SwitchAssignment)
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

type FunctionStmt struct {
	Call FunctionCall
}

func (e FunctionStmt) Accept(visitor StmtVisitor) {
	visitor.VisitFunctionStmtStmt(e)
}
func (e FunctionStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Call: %v", e.Call))
	return fmt.Sprintf("FunctionStmt(%s)", strings.Join(parts, ", "))
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
	Url        Expr
	Stmts      []RadStmt
}

func (e RadBlock) Accept(visitor StmtVisitor) {
	visitor.VisitRadBlockStmt(e)
}
func (e RadBlock) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("RadKeyword: %v", e.RadKeyword))
	parts = append(parts, fmt.Sprintf("Url: %v", e.Url))
	parts = append(parts, fmt.Sprintf("Stmts: %v", e.Stmts))
	return fmt.Sprintf("RadBlock(%s)", strings.Join(parts, ", "))
}

type JsonPathAssign struct {
	Identifier Token
	Path       JsonPath
}

func (e JsonPathAssign) Accept(visitor StmtVisitor) {
	visitor.VisitJsonPathAssignStmt(e)
}
func (e JsonPathAssign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifier: %v", e.Identifier))
	parts = append(parts, fmt.Sprintf("Path: %v", e.Path))
	return fmt.Sprintf("JsonPathAssign(%s)", strings.Join(parts, ", "))
}

type SwitchBlockStmt struct {
	Block SwitchBlock
}

func (e SwitchBlockStmt) Accept(visitor StmtVisitor) {
	visitor.VisitSwitchBlockStmtStmt(e)
}
func (e SwitchBlockStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Block: %v", e.Block))
	return fmt.Sprintf("SwitchBlockStmt(%s)", strings.Join(parts, ", "))
}

type SwitchAssignment struct {
	Identifiers []Token
	Block       SwitchBlock
}

func (e SwitchAssignment) Accept(visitor StmtVisitor) {
	visitor.VisitSwitchAssignmentStmt(e)
}
func (e SwitchAssignment) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifiers: %v", e.Identifiers))
	parts = append(parts, fmt.Sprintf("Block: %v", e.Block))
	return fmt.Sprintf("SwitchAssignment(%s)", strings.Join(parts, ", "))
}
