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

func main() {
	outputDir := "./core"

	defineAst(outputDir, "Literal", "interface{}", []string{
		"StringLiteral   : Token Value",
		"IntLiteral      : Token Value",
		"FloatLiteral    : Token Value",
		"BoolLiteral     : Token Value",
		"NullLiteral     : Token Value",
	})

	defineAst(outputDir, "ArrayLiteral", "interface{}", []string{
		"StringArrayLiteral   : []StringLiteral Values",
		"IntArrayLiteral      : []IntLiteral Values",
		"FloatArrayLiteral    : []FloatLiteral Values",
		"BoolArrayLiteral     : []BoolLiteral Values",
	})

	defineAst(outputDir, "LiteralOrArray", "interface{}", []string{
		"LoaLiteral   : Literal Value",
		"LoaArray     : ArrayLiteral Value",
	})

	defineAst(outputDir, "Expr", "interface{}", []string{
		"LiteralExpr     : Literal Value",
		"ExprLoa	     : LiteralOrArray Value",
		"ArrayExpr	     : []Expr Values",
		"ArrayAccess     : Expr Array, Expr Index",
		"FunctionCall    : Token Function", // todo add args
		"Variable		 : Token Name",
	})

	defineAst(outputDir, "Stmt", "", []string{
		"Empty              :",
		"ExprStmt           : Expr Expression",
		"PrimaryAssign      : Token Name, Expr Initializer",
		"FileHeader         : Token FileHeaderToken",
		"ArgBlock           : Token ArgsKeyword, []ArgStmt ArgStmts",
		"RadBlock           : Token RadKeyword, *Expr Url, []RadStmt RadStmts",
		"JsonPathAssign     : Token Identifier, []JsonPathElement Elements",
	})

	defineAst(outputDir, "ArgStmt", "", []string{
		"ArgDeclaration     : Token Identifier, *Token Rename, *Token Flag, RslType ArgType, " +
			"bool IsOptional, *LiteralOrArray Default, ArgCommentToken Comment",
	})

	defineAst(outputDir, "RadStmt", "", []string{
		"Fields     : []Token Identifiers",
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
