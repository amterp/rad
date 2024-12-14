package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type TypeInfo struct {
	ClassName string
	Fields    string
}

// todo thought on debuggability: ensure *everything* has a Token of some sort, so that we can always point to the source

func main() {
	outputDir := "./core"

	// literal -> STRING | NUMBER | BOOL
	defineAst(outputDir, "Literal", "interface{}", []string{
		"StringLiteral     : []StringLiteralToken Value, []InlineExpr InlineExprs", // expect 1 less InlineExpr than StringLiteralToken
		"IntLiteral        : IntLiteralToken Value, bool IsNegative",
		"FloatLiteral      : FloatLiteralToken Value, bool IsNegative",
		"BoolLiteral       : BoolLiteralToken Value",
		"IdentifierLiteral : Token Tkn", // kinda like 'quoteless' strings. e.g. json.key.etc. Returns as string when visited.
	})

	// arrayLiteral -> "[" ( literal ( "," literal )* )? "]"
	defineAst(outputDir, "ArrayLiteral", "interface{}", []string{
		"MixedArrayLiteral    : []LiteralOrArray Values",
	})

	// literalOrArray -> literal | arrayLiteral
	defineAst(outputDir, "LiteralOrArray", "interface{}", []string{
		"LoaLiteral   : Literal Value",
		"LoaArray     : ArrayLiteral Value",
	})

	// expression       -> logic_or
	// logic_or         -> logic_and ( "or" logic_and )*
	// logic_and        -> equality ( "and" equality )*
	// equality         -> comparison ( ( NOT_EQUAL | EQUAL ) comparison )*
	// comparison       -> term ( ( GT | GTE | LT | LTE ) term )*
	// term             -> factor ( ( "-" | "+" ) factor )*
	// factor           -> unary ( ( "/" | "*" ) unary )*
	// unary            -> ( "!" | "-" ) unary | primary
	// primary          -> "(" expression ")" | literalOrArray | arrayExpr | arrayAccess | functionCall | IDENTIFIER
	// collectionAccess -> expression "[" expression "]"
	// functionCall     -> IDENTIFIER "(" ( ( expression ( "," expression )* )? ( IDENTIFIER "=" expression ( "," IDENTIFIER "=" expression )* )? )? ")"
	defineAst(outputDir, "Expr", "interface{}", []string{
		"ExprLoa           : LiteralOrArray Value",
		"ArrayExpr         : []Expr Values",
		"MapExpr           : []Expr Keys, []Expr Values, Token OpenBraceToken",
		"FunctionCall      : Token Function, []Expr Args, []NamedArg NamedArgs, int NumExpectedReturnValues",
		"Variable          : Token Name",
		"Binary            : Expr Left, Token Operator, Expr Right", // +, -, *, /
		"Ternary           : Expr Condition, Token QuestionMark, Expr True, Expr False",
		"Logical           : Expr Left, Token Operator, Expr Right", // and, or
		"Grouping          : Expr Value",                            // ( expr )
		"Unary             : Token Operator, Expr Right",            // !, -, +
		"ListComprehension : Expr Expression, Token For, Token Identifier1, *Token Identifier2, Expr Range, *Expr Condition",
		// Identifier can be nil
		"VarPath           : Token Identifier, Expr Collection, []CollectionKey Keys", // for e.g. `del myMap["key"][0]`
	})

	defineAst(outputDir, "Stmt", "", []string{
		"Empty                  :",
		"ExprStmt               : Expr Expression",
		"FunctionStmt           : FunctionCall Call",
		// todo can merge Primary and Compound if we treat == as an operator?
		"Assign                 : []Token Identifiers, Expr Initializer", // todo allow []Expr?
		"CompoundAssign         : Token Name, Token Operator, Expr Value",
		"CollectionEntryAssign  : Token Identifier, Expr Key, Token Operator, Expr Value",
		"FileHeader             : FilerHeaderToken FhToken",
		"ArgBlock               : Token ArgsKeyword, []ArgStmt Stmts",
		"RadBlock               : Token RadKeyword, RadBlockType RadType, *Expr Source, []RadStmt Stmts",
		"JsonPathAssign         : Token Identifier, JsonPath Path",
		"SwitchBlockStmt        : SwitchBlock Block",
		"SwitchAssignment       : []Token Identifiers, SwitchBlock Block",
		"ShellCmd               : []Token Identifiers, *Token Unsafe, *Token Quiet, Token Dollar, *Token Bang, Expr CmdExpr, *Block FailBlock, *Block RecoverBlock",
		"Block			        : []Stmt Stmts",
		"IfStmt                 : []IfCase Cases, *Block ElseBlock",
		"ForStmt			    : Token ForToken, Token Identifier1, *Token Identifier2, Expr Range, Block Body",
		"BreakStmt			    : Token BreakToken",
		"ContinueStmt		    : Token ContinueToken",
		"DeleteStmt			    : Token DeleteToken, []VarPath Vars",
		"DeferStmt			    : Token DeferToken, bool IsErrDefer, *Stmt DeferredStmt, *Block DeferredBlock",
	})

	defineAst(outputDir, "ArgStmt", "", []string{
		"ArgDeclaration     : Token Identifier, *Token Rename, *Token Flag, RslArgType ArgType, " + // todo rename 'Rename'?
			"bool IsOptional, *LiteralOrArray Default, *ArgCommentToken Comment",
	})

	defineAst(outputDir, "RadStmt", "", []string{
		"Fields     : []Token Identifiers",
		"Sort	    : Token SortToken, []Token Identifiers, []SortDir Directions, *SortDir GeneralSort",
		"FieldMods  : []Token Identifiers, []RadFieldModStmt Mods",
		"RadIfStmt  : []RadIfCase Cases, *[]RadStmt ElseBlock",
	})

	defineAst(outputDir, "RadFieldModStmt", "", []string{
		"Color      : Token ColorToken, Expr ColorValue, Expr Regex",
		"MapMod     : Token MapToken, Lambda Op",
	})

	defineAst(outputDir, "ValueReturning", "[]interface{}", []string{
		"SwitchBlock  : Token SwitchToken, *Token Discriminator, []SwitchStmt Stmts",
	})

	defineAst(outputDir, "SwitchStmt", "", []string{
		"SwitchCase     : Token CaseKeyword, []StringLiteral Keys, []Expr Values",
		"SwitchDefault  : Token DefaultKeyword, []Expr Values",
	})
}

func defineAst(outputDir, baseName string, returnType string, types []string) {
	path := fmt.Sprintf("%s/gen_%s.go", outputDir, ToLowerSnakeCase(baseName))
	file, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create file %s: %v\n", path, err)
		os.Exit(1)
	}
	defer file.Close()

	funcMap := template.FuncMap{
		"split": strings.Split,
	}

	tmpl, err := template.New("ast").Funcs(funcMap).Parse(astTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse template: %v\n", err)
		os.Exit(1)
	}

	var typeInfos []TypeInfo
	for _, t := range types {
		parts := strings.Split(t, ":")
		typeInfos = append(typeInfos, TypeInfo{
			ClassName: strings.TrimSpace(parts[0]),
			Fields:    strings.TrimSpace(parts[1]),
		})
	}

	data := struct {
		BaseName   string
		ReturnType string
		Types      []TypeInfo
	}{
		BaseName:   baseName,
		ReturnType: returnType,
		Types:      typeInfos,
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not execute template: %v\n", err)
		os.Exit(1)
	}
}

func ToLowerSnakeCase(str string) string {
	pattern := regexp.MustCompile(`([A-Z])`)
	newStr := pattern.ReplaceAllString(str, "_"+strings.ToLower("${1}"))[1:]
	return strings.ToLower(newStr)
}

const astTemplate = `// GENERATED -- DO NOT EDIT
package core
import (
    "fmt"
    "strings"
)
type {{.BaseName}} interface {
    Accept(visitor {{.BaseName}}Visitor){{if .ReturnType}} {{.ReturnType}}{{end}}
}
type {{.BaseName}}Visitor interface {
{{- range .Types }}
    Visit{{.ClassName}}{{$.BaseName}}({{.ClassName}}){{if $.ReturnType}} {{$.ReturnType}}{{end}}
{{- end }}
}
{{- range .Types }}
type {{.ClassName}} struct {
{{- if .Fields }}
{{- $fields := split .Fields ", " }}
{{- range $fields }}
    {{- $parts := split . " " }}
    {{index $parts 1}} {{index $parts 0}}
{{- end }}
{{- end }}
}
func (e {{.ClassName}}) Accept(visitor {{$.BaseName}}Visitor){{if $.ReturnType}} {{$.ReturnType}}{{end}} {
    {{if $.ReturnType}}return {{end}}visitor.Visit{{.ClassName}}{{$.BaseName}}(e)
}
func (e {{.ClassName}}) String() string {
{{- if .Fields }}
    var parts []string
{{- $fields := split .Fields ", " }}
{{- range $fields }}
    {{- $parts := split . " " }}
    parts = append(parts, fmt.Sprintf("{{index $parts 1}}: %v", e.{{index $parts 1}}))
{{- end }}
    return fmt.Sprintf("{{.ClassName}}(%s)", strings.Join(parts, ", "))
{{- else }}
    return "{{.ClassName}}()"
{{- end }}
}
{{- end }}
`
