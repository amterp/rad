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
	VisitAssignStmt(Assign)
	VisitCollectionEntryAssignStmt(CollectionEntryAssign)
	VisitFileHeaderStmt(FileHeader)
	VisitArgBlockStmt(ArgBlock)
	VisitRadBlockStmt(RadBlock)
	VisitJsonPathAssignStmt(JsonPathAssign)
	VisitSwitchBlockStmtStmt(SwitchBlockStmt)
	VisitSwitchAssignmentStmt(SwitchAssignment)
	VisitShellCmdStmt(ShellCmd)
	VisitBlockStmt(Block)
	VisitIfStmtStmt(IfStmt)
	VisitForStmtStmt(ForStmt)
	VisitBreakStmtStmt(BreakStmt)
	VisitContinueStmtStmt(ContinueStmt)
	VisitDeleteStmtStmt(DeleteStmt)
	VisitDeferStmtStmt(DeferStmt)
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

type Assign struct {
	Identifiers []Token
	Initializer Expr
}

func (e Assign) Accept(visitor StmtVisitor) {
	visitor.VisitAssignStmt(e)
}
func (e Assign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifiers: %v", e.Identifiers))
	parts = append(parts, fmt.Sprintf("Initializer: %v", e.Initializer))
	return fmt.Sprintf("Assign(%s)", strings.Join(parts, ", "))
}

type CollectionEntryAssign struct {
	Identifier Token
	Key        Expr
	Operator   Token
	Value      Expr
}

func (e CollectionEntryAssign) Accept(visitor StmtVisitor) {
	visitor.VisitCollectionEntryAssignStmt(e)
}
func (e CollectionEntryAssign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifier: %v", e.Identifier))
	parts = append(parts, fmt.Sprintf("Key: %v", e.Key))
	parts = append(parts, fmt.Sprintf("Operator: %v", e.Operator))
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("CollectionEntryAssign(%s)", strings.Join(parts, ", "))
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
	RadType    RadBlockType
	Source     *Expr
	Stmts      []RadStmt
}

func (e RadBlock) Accept(visitor StmtVisitor) {
	visitor.VisitRadBlockStmt(e)
}
func (e RadBlock) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("RadKeyword: %v", e.RadKeyword))
	parts = append(parts, fmt.Sprintf("RadType: %v", e.RadType))
	parts = append(parts, fmt.Sprintf("Source: %v", e.Source))
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

type ShellCmd struct {
	Identifiers  []Token
	Unsafe       *Token
	Quiet        *Token
	Dollar       Token
	Bang         *Token
	CmdExpr      Expr
	FailBlock    *Block
	RecoverBlock *Block
}

func (e ShellCmd) Accept(visitor StmtVisitor) {
	visitor.VisitShellCmdStmt(e)
}
func (e ShellCmd) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifiers: %v", e.Identifiers))
	parts = append(parts, fmt.Sprintf("Unsafe: %v", e.Unsafe))
	parts = append(parts, fmt.Sprintf("Quiet: %v", e.Quiet))
	parts = append(parts, fmt.Sprintf("Dollar: %v", e.Dollar))
	parts = append(parts, fmt.Sprintf("Bang: %v", e.Bang))
	parts = append(parts, fmt.Sprintf("CmdExpr: %v", e.CmdExpr))
	parts = append(parts, fmt.Sprintf("FailBlock: %v", e.FailBlock))
	parts = append(parts, fmt.Sprintf("RecoverBlock: %v", e.RecoverBlock))
	return fmt.Sprintf("ShellCmd(%s)", strings.Join(parts, ", "))
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

type DeleteStmt struct {
	DeleteToken Token
	Vars        []VarPath
}

func (e DeleteStmt) Accept(visitor StmtVisitor) {
	visitor.VisitDeleteStmtStmt(e)
}
func (e DeleteStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("DeleteToken: %v", e.DeleteToken))
	parts = append(parts, fmt.Sprintf("Vars: %v", e.Vars))
	return fmt.Sprintf("DeleteStmt(%s)", strings.Join(parts, ", "))
}

type DeferStmt struct {
	DeferToken    Token
	IsErrDefer    bool
	DeferredStmt  *Stmt
	DeferredBlock *Block
}

func (e DeferStmt) Accept(visitor StmtVisitor) {
	visitor.VisitDeferStmtStmt(e)
}
func (e DeferStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("DeferToken: %v", e.DeferToken))
	parts = append(parts, fmt.Sprintf("IsErrDefer: %v", e.IsErrDefer))
	parts = append(parts, fmt.Sprintf("DeferredStmt: %v", e.DeferredStmt))
	parts = append(parts, fmt.Sprintf("DeferredBlock: %v", e.DeferredBlock))
	return fmt.Sprintf("DeferStmt(%s)", strings.Join(parts, ", "))
}
