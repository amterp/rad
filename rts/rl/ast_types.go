package rl

// SourceFile is the root AST node for a script.
type SourceFile struct {
	span   Span
	Header *FileHeader
	Args   *ArgBlock
	Cmds   []*CmdBlock
	Stmts  []Node
}

func NewSourceFile(span Span, stmts []Node) *SourceFile {
	return &SourceFile{span: span, Stmts: stmts}
}
func (n *SourceFile) Kind() NodeKind { return NSourceFile }
func (n *SourceFile) Span() Span     { return n.span }

// --- Statements ---

// Assign handles simple assignment, unpacking, and desugared compound
// assign (+=) and increment/decrement (++).
type Assign struct {
	span            Span
	Targets         []Node      // left-hand sides (var paths)
	Values          []Node      // right-hand sides
	IsUnpacking     bool        // true if `a, b = ...` syntax
	UpdateEnclosing bool        // true for compound assign/incr-decr (updates enclosing scope)
	Catch           *CatchBlock // optional catch block
}

func NewAssign(span Span, targets, values []Node, isUnpacking bool, catch *CatchBlock) *Assign {
	return &Assign{span: span, Targets: targets, Values: values, IsUnpacking: isUnpacking, Catch: catch}
}
func (n *Assign) Kind() NodeKind { return NAssign }
func (n *Assign) Span() Span     { return n.span }

// CatchBlock represents an error-catching block attached to a statement.
type CatchBlock struct {
	span  Span
	Stmts []Node
}

func NewCatchBlock(span Span, stmts []Node) *CatchBlock {
	return &CatchBlock{span: span, Stmts: stmts}
}
func (n *CatchBlock) Span() Span { return n.span }

// ExprStmt wraps an expression used as a statement.
type ExprStmt struct {
	span  Span
	Expr  Node
	Catch *CatchBlock
}

func NewExprStmt(span Span, expr Node, catch *CatchBlock) *ExprStmt {
	return &ExprStmt{span: span, Expr: expr, Catch: catch}
}
func (n *ExprStmt) Kind() NodeKind { return NExprStmt }
func (n *ExprStmt) Span() Span     { return n.span }

// If represents an if/elif/else chain.
type If struct {
	span     Span
	Branches []IfBranch
}

// IfBranch is a single branch (if, elif, or else).
// Condition is nil for the else branch.
type IfBranch struct {
	Condition Node   // nil for else
	Body      []Node // statements
}

func NewIf(span Span, branches []IfBranch) *If {
	return &If{span: span, Branches: branches}
}
func (n *If) Kind() NodeKind { return NIf }
func (n *If) Span() Span     { return n.span }

// Switch represents a switch/case statement.
type Switch struct {
	span         Span
	Discriminant Node         // the value being switched on
	Cases        []SwitchCase // case branches
	Default      *SwitchDefault
}

// SwitchCase is a single case branch.
type SwitchCase struct {
	Keys []Node // match values (can be multiple: case 1, 2:)
	Alt  Node   // either SwitchCaseExpr or SwitchCaseBlock
}

// SwitchDefault is the default branch of a switch.
type SwitchDefault struct {
	Alt Node // either SwitchCaseExpr or SwitchCaseBlock
}

// SwitchCaseExpr is a single-expression case value (case X: expr).
type SwitchCaseExpr struct {
	span   Span
	Values []Node // right-hand side values
}

func NewSwitchCaseExpr(span Span, values []Node) *SwitchCaseExpr {
	return &SwitchCaseExpr{span: span, Values: values}
}
func (n *SwitchCaseExpr) Kind() NodeKind { return NSwitchCaseExpr }
func (n *SwitchCaseExpr) Span() Span     { return n.span }

// SwitchCaseBlock is a multi-statement case body (case X:\n  stmts).
type SwitchCaseBlock struct {
	span  Span
	Stmts []Node
}

func NewSwitchCaseBlock(span Span, stmts []Node) *SwitchCaseBlock {
	return &SwitchCaseBlock{span: span, Stmts: stmts}
}
func (n *SwitchCaseBlock) Kind() NodeKind { return NSwitchCaseBlock }
func (n *SwitchCaseBlock) Span() Span     { return n.span }

func NewSwitch(span Span, discriminant Node, cases []SwitchCase, dflt *SwitchDefault) *Switch {
	return &Switch{span: span, Discriminant: discriminant, Cases: cases, Default: dflt}
}
func (n *Switch) Kind() NodeKind { return NSwitch }
func (n *Switch) Span() Span     { return n.span }

// ForLoop represents a for ... in ... loop.
type ForLoop struct {
	span    Span
	Vars    []string // loop variable names
	Iter    Node     // the iterable expression
	Body    []Node   // statement list
	Context *string  // optional "with x" context name
}

func NewForLoop(span Span, vars []string, iter Node, body []Node, context *string) *ForLoop {
	return &ForLoop{span: span, Vars: vars, Iter: iter, Body: body, Context: context}
}
func (n *ForLoop) Kind() NodeKind { return NForLoop }
func (n *ForLoop) Span() Span     { return n.span }

// WhileLoop represents a while loop. Condition is nil for infinite loops.
type WhileLoop struct {
	span      Span
	Condition Node   // nil = infinite loop
	Body      []Node // statement list
}

func NewWhileLoop(span Span, condition Node, body []Node) *WhileLoop {
	return &WhileLoop{span: span, Condition: condition, Body: body}
}
func (n *WhileLoop) Kind() NodeKind { return NWhileLoop }
func (n *WhileLoop) Span() Span     { return n.span }

// Shell represents a shell statement ($...).
type Shell struct {
	span      Span
	Targets   []Node      // assignment targets (if shell assigns to vars)
	Cmd       Node        // the command expression
	Catch     *CatchBlock // optional catch block
	IsQuiet   bool        // quiet modifier
	IsConfirm bool        // confirm modifier
}

func NewShell(span Span, targets []Node, cmd Node, catch *CatchBlock, isQuiet, isConfirm bool) *Shell {
	return &Shell{span: span, Targets: targets, Cmd: cmd, Catch: catch, IsQuiet: isQuiet, IsConfirm: isConfirm}
}
func (n *Shell) Kind() NodeKind { return NShell }
func (n *Shell) Span() Span     { return n.span }

// Del represents a del statement.
type Del struct {
	span    Span
	Targets []Node // var paths to delete
}

func NewDel(span Span, targets []Node) *Del {
	return &Del{span: span, Targets: targets}
}
func (n *Del) Kind() NodeKind { return NDel }
func (n *Del) Span() Span     { return n.span }

// Defer represents a defer or errdefer block.
type Defer struct {
	span       Span
	IsErrDefer bool   // true for errdefer
	Body       []Node // deferred statements
}

func NewDefer(span Span, isErrDefer bool, body []Node) *Defer {
	return &Defer{span: span, IsErrDefer: isErrDefer, Body: body}
}
func (n *Defer) Kind() NodeKind { return NDefer }
func (n *Defer) Span() Span     { return n.span }

// Break represents a break statement.
type Break struct{ span Span }

func NewBreak(span Span) *Break { return &Break{span: span} }
func (n *Break) Kind() NodeKind { return NBreak }
func (n *Break) Span() Span     { return n.span }

// Continue represents a continue statement.
type Continue struct{ span Span }

func NewContinue(span Span) *Continue { return &Continue{span: span} }
func (n *Continue) Kind() NodeKind    { return NContinue }
func (n *Continue) Span() Span        { return n.span }

// Return represents a return statement with optional values.
type Return struct {
	span   Span
	Values []Node
}

func NewReturn(span Span, values []Node) *Return {
	return &Return{span: span, Values: values}
}
func (n *Return) Kind() NodeKind { return NReturn }
func (n *Return) Span() Span     { return n.span }

// Yield represents a yield statement with optional values.
type Yield struct {
	span   Span
	Values []Node
}

func NewYield(span Span, values []Node) *Yield {
	return &Yield{span: span, Values: values}
}
func (n *Yield) Kind() NodeKind { return NYield }
func (n *Yield) Span() Span     { return n.span }

// Pass is a no-op statement.
type Pass struct{ span Span }

func NewPass(span Span) *Pass  { return &Pass{span: span} }
func (n *Pass) Kind() NodeKind { return NPass }
func (n *Pass) Span() Span     { return n.span }

// FnDef represents a named function definition.
type FnDef struct {
	span    Span
	Name    string
	Typing  *TypingFnT // parameter types and return type
	Body    []Node
	IsBlock bool // block function (uses return stmt) vs expression function
	DefSpan Span // span of the keyword or name for error reporting
}

func NewFnDef(span Span, name string, typing *TypingFnT, body []Node, isBlock bool, defSpan Span) *FnDef {
	return &FnDef{span: span, Name: name, Typing: typing, Body: body, IsBlock: isBlock, DefSpan: defSpan}
}
func (n *FnDef) Kind() NodeKind { return NFnDef }
func (n *FnDef) Span() Span     { return n.span }

// --- Expressions ---

// OpBinary represents a binary operation.
type OpBinary struct {
	span       Span
	Op         Operator
	Left       Node
	Right      Node
	IsCompound bool // true when desugared from compound assign (e.g. +=)
}

func NewOpBinary(span Span, op Operator, left, right Node) *OpBinary {
	return &OpBinary{span: span, Op: op, Left: left, Right: right}
}
func (n *OpBinary) Kind() NodeKind { return NOpBinary }
func (n *OpBinary) Span() Span     { return n.span }

// OpUnary represents a unary operation.
type OpUnary struct {
	span    Span
	Op      Operator
	Operand Node
}

func NewOpUnary(span Span, op Operator, operand Node) *OpUnary {
	return &OpUnary{span: span, Op: op, Operand: operand}
}
func (n *OpUnary) Kind() NodeKind { return NOpUnary }
func (n *OpUnary) Span() Span     { return n.span }

// Ternary represents a ternary expression (cond ? true : false).
type Ternary struct {
	span      Span
	Condition Node
	True      Node
	False     Node
}

func NewTernary(span Span, condition, trueExpr, falseExpr Node) *Ternary {
	return &Ternary{span: span, Condition: condition, True: trueExpr, False: falseExpr}
}
func (n *Ternary) Kind() NodeKind { return NTernary }
func (n *Ternary) Span() Span     { return n.span }

// Fallback represents the null-coalescing operator (left ?? right).
type Fallback struct {
	span  Span
	Left  Node
	Right Node
}

func NewFallback(span Span, left, right Node) *Fallback {
	return &Fallback{span: span, Left: left, Right: right}
}
func (n *Fallback) Kind() NodeKind { return NFallback }
func (n *Fallback) Span() Span     { return n.span }

// CatchExpr represents the error-catching operator (left catch right).
// Unlike Fallback (??), CatchExpr only catches errors, not null values.
type CatchExpr struct {
	span  Span
	Left  Node
	Right Node
}

func NewCatchExpr(span Span, left, right Node) *CatchExpr {
	return &CatchExpr{span: span, Left: left, Right: right}
}
func (n *CatchExpr) Kind() NodeKind { return NCatchExpr }
func (n *CatchExpr) Span() Span     { return n.span }

// Call represents a function call.
type Call struct {
	span      Span
	Func      Node           // the function expression
	Args      []Node         // positional arguments
	NamedArgs []CallNamedArg // named arguments
}

// CallNamedArg is a named argument in a function call.
type CallNamedArg struct {
	Name      string
	NameSpan  Span // span of the argument name
	Value     Node
	ValueSpan Span // span of the argument value
}

func NewCall(span Span, fn Node, args []Node, namedArgs []CallNamedArg) *Call {
	return &Call{span: span, Func: fn, Args: args, NamedArgs: namedArgs}
}
func (n *Call) Kind() NodeKind { return NCall }
func (n *Call) Span() Span     { return n.span }

// VarPath represents chained access: a.b[c].d
type VarPath struct {
	span     Span
	Root     Node          // the base identifier or expression
	Segments []PathSegment // .field, [expr], [start:end]
}

func NewVarPath(span Span, root Node, segments []PathSegment) *VarPath {
	return &VarPath{span: span, Root: root, Segments: segments}
}
func (n *VarPath) Kind() NodeKind { return NVarPath }
func (n *VarPath) Span() Span     { return n.span }

// PathSegment represents one segment in a var path.
type PathSegment struct {
	span    Span
	Field   *string // dot access: .name (nil if bracket)
	Index   Node    // bracket access: [expr] (nil if dot or slice)
	IsSlice bool    // [start:end] slice syntax
	IsUFCS  bool    // true when indexing node was directly a K_CALL (UFCS syntax)
	Start   Node    // slice start (may be nil)
	End     Node    // slice end (may be nil)
}

// Span returns the span of this path segment.
func (p PathSegment) Span() Span { return p.span }

func NewPathSegmentField(span Span, field string) PathSegment {
	return PathSegment{span: span, Field: &field}
}

func NewPathSegmentIndex(span Span, index Node) PathSegment {
	return PathSegment{span: span, Index: index}
}

func NewPathSegmentSlice(span Span, start, end Node) PathSegment {
	return PathSegment{span: span, IsSlice: true, Start: start, End: end}
}

// Identifier represents a variable reference.
type Identifier struct {
	span Span
	Name string
}

func NewIdentifier(span Span, name string) *Identifier {
	return &Identifier{span: span, Name: name}
}
func (n *Identifier) Kind() NodeKind { return NIdentifier }
func (n *Identifier) Span() Span     { return n.span }

// Lambda represents an anonymous function.
type Lambda struct {
	span    Span
	Typing  *TypingFnT // parameter types and return type
	Body    []Node
	IsBlock bool // block function vs expression function
	DefSpan Span // span of the keyword for error reporting
}

func NewLambda(span Span, typing *TypingFnT, body []Node, isBlock bool, defSpan Span) *Lambda {
	return &Lambda{span: span, Typing: typing, Body: body, IsBlock: isBlock, DefSpan: defSpan}
}
func (n *Lambda) Kind() NodeKind { return NLambda }
func (n *Lambda) Span() Span     { return n.span }

// --- Literals ---

// LitInt represents an integer literal.
type LitInt struct {
	span  Span
	Value int64
}

func NewLitInt(span Span, value int64) *LitInt {
	return &LitInt{span: span, Value: value}
}
func (n *LitInt) Kind() NodeKind { return NLitInt }
func (n *LitInt) Span() Span     { return n.span }

// LitFloat represents a floating-point literal.
type LitFloat struct {
	span  Span
	Value float64
}

func NewLitFloat(span Span, value float64) *LitFloat {
	return &LitFloat{span: span, Value: value}
}
func (n *LitFloat) Kind() NodeKind { return NLitFloat }
func (n *LitFloat) Span() Span     { return n.span }

// LitBool represents a boolean literal.
type LitBool struct {
	span  Span
	Value bool
}

func NewLitBool(span Span, value bool) *LitBool {
	return &LitBool{span: span, Value: value}
}
func (n *LitBool) Kind() NodeKind { return NLitBool }
func (n *LitBool) Span() Span     { return n.span }

// LitNull represents a null literal.
type LitNull struct{ span Span }

func NewLitNull(span Span) *LitNull { return &LitNull{span: span} }
func (n *LitNull) Kind() NodeKind   { return NLitNull }
func (n *LitNull) Span() Span       { return n.span }

// LitString uses a hybrid representation:
// simple strings (no interpolation) store just Value;
// interpolated strings use Segments.
type LitString struct {
	span     Span
	Simple   bool            // true = use Value, false = use Segments
	Value    string          // resolved string (when Simple)
	Segments []StringSegment // literal text + interpolation exprs
}

func NewLitStringSimple(span Span, value string) *LitString {
	return &LitString{span: span, Simple: true, Value: value}
}

func NewLitStringInterpolated(span Span, segments []StringSegment) *LitString {
	return &LitString{span: span, Simple: false, Segments: segments}
}
func (n *LitString) Kind() NodeKind { return NLitString }
func (n *LitString) Span() Span     { return n.span }

// StringSegment is either literal text or an interpolation expression.
type StringSegment struct {
	IsLiteral bool                 // true = Text only, false = Expr
	Text      string               // literal text (resolved escapes)
	Expr      Node                 // interpolation expression (when !IsLiteral)
	Format    *InterpolationFormat // optional format specifiers
	span      Span                 // span of the interpolation {expr:fmt} (only for !IsLiteral)
}

// NewStringSegmentInterp creates an interpolation string segment.
func NewStringSegmentInterp(expr Node, format *InterpolationFormat, span Span) StringSegment {
	return StringSegment{IsLiteral: false, Expr: expr, Format: format, span: span}
}

// Span returns the span of this string segment.
func (s StringSegment) Span() Span { return s.span }

// InterpolationFormat holds pre-extracted format specifiers for
// string interpolation (alignment, padding, precision, etc.).
type InterpolationFormat struct {
	ThousandsSeparator bool
	Alignment          string // "<" for left, "" for right (default)
	FillChar           string // explicit fill char from fill_alignment (e.g. "*" from :*>), "" if not specified
	ZeroPad            bool   // :05 shorthand (any type; sign-aware for numbers) vs :0>5 (explicit fill)
	Padding            Node   // padding width expression
	Precision          Node   // precision expression
}

// LitList represents a list literal [a, b, c].
type LitList struct {
	span     Span
	Elements []Node
}

func NewLitList(span Span, elements []Node) *LitList {
	return &LitList{span: span, Elements: elements}
}
func (n *LitList) Kind() NodeKind { return NLitList }
func (n *LitList) Span() Span     { return n.span }

// LitMap represents a map literal {k: v, ...}.
type LitMap struct {
	span    Span
	Entries []MapEntry
}

// MapEntry is a key-value pair in a map literal.
type MapEntry struct {
	Key   Node
	Value Node
}

func NewLitMap(span Span, entries []MapEntry) *LitMap {
	return &LitMap{span: span, Entries: entries}
}
func (n *LitMap) Kind() NodeKind { return NLitMap }
func (n *LitMap) Span() Span     { return n.span }

// --- Comprehension ---

// ListComp represents a list comprehension [expr for vars in iter if cond].
type ListComp struct {
	span      Span
	Expr      Node     // the expression to evaluate per iteration
	Vars      []string // loop variable names
	Iter      Node     // the iterable expression
	Condition Node     // optional filter condition (nil if absent)
	Context   *string  // optional "with x" context name
}

func NewListComp(span Span, expr Node, vars []string, iter Node, condition Node, context *string) *ListComp {
	return &ListComp{span: span, Expr: expr, Vars: vars, Iter: iter, Condition: condition, Context: context}
}
func (n *ListComp) Kind() NodeKind { return NListComp }
func (n *ListComp) Span() Span     { return n.span }

// --- Rad block internals ---

// RadBlock represents a rad/request/display block.
type RadBlock struct {
	span      Span
	BlockType string // "rad", "request", "display"
	Source    Node   // the source expression
	Stmts     []Node // block statements (RadField, RadSort, etc.)
}

func NewRadBlock(span Span, blockType string, source Node, stmts []Node) *RadBlock {
	return &RadBlock{span: span, BlockType: blockType, Source: source, Stmts: stmts}
}
func (n *RadBlock) Kind() NodeKind { return NRadBlock }
func (n *RadBlock) Span() Span     { return n.span }

// RadField represents a field declaration in a rad block.
// A single statement can declare multiple fields (e.g. "name age email").
type RadField struct {
	span        Span
	Identifiers []Node // field name Identifier nodes
}

func NewRadField(span Span, identifiers []Node) *RadField {
	return &RadField{span: span, Identifiers: identifiers}
}
func (n *RadField) Kind() NodeKind { return NRadField }
func (n *RadField) Span() Span     { return n.span }

// RadSort represents a sort specifier in a rad block.
type RadSort struct {
	span       Span
	Specifiers []RadSortSpecifier
}

// RadSortSpecifier is a single sort field + direction.
type RadSortSpecifier struct {
	Field     string // field name to sort by
	Ascending bool   // true = asc, false = desc
}

func NewRadSort(span Span, specifiers []RadSortSpecifier) *RadSort {
	return &RadSort{span: span, Specifiers: specifiers}
}
func (n *RadSort) Kind() NodeKind { return NRadSort }
func (n *RadSort) Span() Span     { return n.span }

// RadFieldMod represents a field modifier statement.
// At the container level (K_RAD_FIELD_MODIFIER_STMT): Fields holds the target
// field identifiers, ModType is empty, and Args holds the child modifier nodes.
// At the individual level (color/map/filter): Fields is nil, ModType is set,
// and Args holds the modifier arguments (expressions/lambdas).
type RadFieldMod struct {
	span    Span
	Fields  []Node // target field identifiers (nil for individual modifiers)
	ModType string // "color", "map", "filter" (empty for container)
	Args    []Node // modifier arguments or child modifier nodes
}

func NewRadFieldMod(span Span, fields []Node, modType string, args []Node) *RadFieldMod {
	return &RadFieldMod{span: span, Fields: fields, ModType: modType, Args: args}
}
func (n *RadFieldMod) Kind() NodeKind { return NRadFieldMod }
func (n *RadFieldMod) Span() Span     { return n.span }

// RadIf represents a conditional in a rad block (if/elif/else).
// Mirrors the top-level If node's branch structure.
type RadIf struct {
	span     Span
	Branches []IfBranch // reuses IfBranch from the top-level If node
}

func NewRadIf(span Span, branches []IfBranch) *RadIf {
	return &RadIf{span: span, Branches: branches}
}
func (n *RadIf) Kind() NodeKind { return NRadIf }
func (n *RadIf) Span() Span     { return n.span }

// JsonPath represents a json field path expression (e.g. json[].age, json[0].name.first).
type JsonPath struct {
	span     Span
	Segments []JsonPathSeg
}

// JsonPathSeg is one segment in a json path (e.g. "json[]" or ".age" or ".names[0]").
type JsonPathSeg struct {
	Key     string // the identifier part (e.g. "json", "age", "names")
	KeySpan Span   // span of the key for error reporting
	Indexes []JsonPathIdx
}

// JsonPathIdx is an index within a json path segment. Expr is nil for wildcard [].
type JsonPathIdx struct {
	Span Span
	Expr Node // nil = wildcard
}

func NewJsonPath(span Span, segments []JsonPathSeg) *JsonPath {
	return &JsonPath{span: span, Segments: segments}
}
func (n *JsonPath) Kind() NodeKind { return NJsonPath }
func (n *JsonPath) Span() Span     { return n.span }

// --- Script metadata blocks ---

// FileHeader represents the --- block with description and metadata entries.
type FileHeader struct {
	span            Span
	Contents        string
	MetadataEntries map[string]string
}

func NewFileHeader(span Span, contents string, metadata map[string]string) *FileHeader {
	return &FileHeader{span: span, Contents: contents, MetadataEntries: metadata}
}
func (n *FileHeader) Kind() NodeKind { return NFileHeader }
func (n *FileHeader) Span() Span     { return n.span }

// ArgBlock represents the top-level args: block.
type ArgBlock struct {
	span             Span
	Decls            []ArgDecl
	EnumConstraints  map[string]*ArgEnumConstraint
	RegexConstraints map[string]*ArgRegexConstraint
	RangeConstraints map[string]*ArgRangeConstraint
	Requirements     []ArgRelation
	Exclusions       []ArgRelation
}

func NewArgBlock(span Span, decls []ArgDecl) *ArgBlock {
	return &ArgBlock{span: span, Decls: decls}
}
func (n *ArgBlock) Kind() NodeKind { return NArgBlock }
func (n *ArgBlock) Span() Span     { return n.span }

// ArgDecl represents a single arg declaration (e.g. "name str? = "default").
type ArgDecl struct {
	span       Span
	Name       string
	TypeName   string // "int", "str", "int[]", etc.
	IsOptional bool
	IsVariadic bool
	Rename     *string
	Shorthand  *string
	Default    Node // expression node or nil
	Comment    *string
	// Pre-parsed typed default values for metadata consumers
	DefaultString     *string
	DefaultInt        *int64
	DefaultFloat      *float64
	DefaultBool       *bool
	DefaultStringList *[]string
	DefaultIntList    *[]int64
	DefaultFloatList  *[]float64
	DefaultBoolList   *[]bool
}

func NewArgDecl(span Span, name, typeName string) *ArgDecl {
	return &ArgDecl{span: span, Name: name, TypeName: typeName}
}
func (n *ArgDecl) Kind() NodeKind { return NArgDecl }
func (n *ArgDecl) Span() Span     { return n.span }

// ExternalName returns the CLI-visible name (hyphenated form), or the rename if set.
func (n *ArgDecl) ExternalName() string {
	if n.Rename != nil {
		return *n.Rename
	}
	return toExternalArgName(n.Name)
}

// toExternalArgName converts snake_case to kebab-case for CLI flag names.
func toExternalArgName(name string) string {
	result := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		if name[i] == '_' {
			result[i] = '-'
		} else {
			result[i] = name[i]
		}
	}
	return string(result)
}

// ArgEnumConstraint restricts an arg to a set of string values.
type ArgEnumConstraint struct {
	Span_  Span
	Values []string
}

// ArgRegexConstraint restricts an arg value to match a regex.
type ArgRegexConstraint struct {
	Span_ Span
	Value string
}

// ArgRangeConstraint restricts an arg to a numeric range.
type ArgRangeConstraint struct {
	Span_       Span
	OpenerToken string // "[" or "("
	CloserToken string // "]" or ")"
	Min         *float64
	Max         *float64
}

// ArgRelation represents a requires or excludes relation between args.
type ArgRelation struct {
	Span_    Span
	Arg      string
	IsMutual bool
	Related  []string // required or excluded arg names
}

// CmdBlock represents a command definition block.
type CmdBlock struct {
	span             Span
	Name             string
	Description      *string
	Decls            []ArgDecl
	EnumConstraints  map[string]*ArgEnumConstraint
	RegexConstraints map[string]*ArgRegexConstraint
	RangeConstraints map[string]*ArgRangeConstraint
	Requirements     []ArgRelation
	Exclusions       []ArgRelation
	Callback         CmdCallback
}

func NewCmdBlock(span Span, name string) *CmdBlock {
	return &CmdBlock{span: span, Name: name}
}
func (n *CmdBlock) Kind() NodeKind { return NCmdBlock }
func (n *CmdBlock) Span() Span     { return n.span }

// CmdCallback represents a command's callback - either a function name or inline lambda.
type CmdCallback struct {
	Span_          Span
	IsLambda       bool
	IdentifierName *string
	IdentifierSpan *Span
	Lambda         *Lambda
}
