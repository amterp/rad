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
	VisitBlockStmt(Block)
	VisitIfStmtStmt(IfStmt)
	VisitIfCaseStmt(IfCase)
	VisitForStmtStmt(ForStmt)
	VisitBreakStmtStmt(BreakStmt)
	VisitContinueStmtStmt(ContinueStmt)
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
	VarType     *RslType
	Initializer Expr
}

func (e PrimaryAssign) Accept(visitor StmtVisitor) {
	visitor.VisitPrimaryAssignStmt(e)
}
func (e PrimaryAssign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Name: %v", e.Name))
	parts = append(parts, fmt.Sprintf("VarType: %v", e.VarType))
	parts = append(parts, fmt.Sprintf("Initializer: %v", e.Initializer))
	return fmt.Sprintf("PrimaryAssign(%s)", strings.Join(parts, ", "))
}

type FileHeader struct {
	FhToken FilerHeaderToken
}

func (e FileHeader) Accept(visitor StmtVisitor) {
	visitor.VisitFileHeaderStmt(e)
}
func (e FileHeader) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("FhToken: %v", e.FhToken))
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
	VarTypes    []*RslType
	Block       SwitchBlock
}

func (e SwitchAssignment) Accept(visitor StmtVisitor) {
	visitor.VisitSwitchAssignmentStmt(e)
}
func (e SwitchAssignment) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifiers: %v", e.Identifiers))
	parts = append(parts, fmt.Sprintf("VarTypes: %v", e.VarTypes))
	parts = append(parts, fmt.Sprintf("Block: %v", e.Block))
	return fmt.Sprintf("SwitchAssignment(%s)", strings.Join(parts, ", "))
}

type Block struct {
	Stmts []Stmt
}

func (e Block) Accept(visitor StmtVisitor) {
	visitor.VisitBlockStmt(e)
}
func (e Block) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Stmts: %v", e.Stmts))
	return fmt.Sprintf("Block(%s)", strings.Join(parts, ", "))
}

type IfStmt struct {
	Cases     []IfCase
	ElseBlock *Block
}

func (e IfStmt) Accept(visitor StmtVisitor) {
	visitor.VisitIfStmtStmt(e)
}
func (e IfStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Cases: %v", e.Cases))
	parts = append(parts, fmt.Sprintf("ElseBlock: %v", e.ElseBlock))
	return fmt.Sprintf("IfStmt(%s)", strings.Join(parts, ", "))
}

type IfCase struct {
	IfToken   Token
	Condition Expr
	Body      Block
}

func (e IfCase) Accept(visitor StmtVisitor) {
	visitor.VisitIfCaseStmt(e)
}
func (e IfCase) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("IfToken: %v", e.IfToken))
	parts = append(parts, fmt.Sprintf("Condition: %v", e.Condition))
	parts = append(parts, fmt.Sprintf("Body: %v", e.Body))
	return fmt.Sprintf("IfCase(%s)", strings.Join(parts, ", "))
}

type ForStmt struct {
	ForToken    Token
	Identifier1 Token
	Identifier2 *Token
	Range       Expr
	Body        Block
}

func (e ForStmt) Accept(visitor StmtVisitor) {
	visitor.VisitForStmtStmt(e)
}
func (e ForStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("ForToken: %v", e.ForToken))
	parts = append(parts, fmt.Sprintf("Identifier1: %v", e.Identifier1))
	parts = append(parts, fmt.Sprintf("Identifier2: %v", e.Identifier2))
	parts = append(parts, fmt.Sprintf("Range: %v", e.Range))
	parts = append(parts, fmt.Sprintf("Body: %v", e.Body))
	return fmt.Sprintf("ForStmt(%s)", strings.Join(parts, ", "))
}

type BreakStmt struct {
	BreakToken Token
}

func (e BreakStmt) Accept(visitor StmtVisitor) {
	visitor.VisitBreakStmtStmt(e)
}
func (e BreakStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("BreakToken: %v", e.BreakToken))
	return fmt.Sprintf("BreakStmt(%s)", strings.Join(parts, ", "))
}

type ContinueStmt struct {
	ContinueToken Token
}

func (e ContinueStmt) Accept(visitor StmtVisitor) {
	visitor.VisitContinueStmtStmt(e)
}
func (e ContinueStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("ContinueToken: %v", e.ContinueToken))
	return fmt.Sprintf("ContinueStmt(%s)", strings.Join(parts, ", "))
}
