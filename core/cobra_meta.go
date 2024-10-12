package core

import (
	"fmt"
	"github.com/samber/lo"
	"strings"
)

type ScriptMetadata struct {
	Args               []ScriptArg
	OneLineDescription *string
	BlockDescription   *string
}

func GenerateUseString(scriptName string, args []ScriptArg) string {
	useString := scriptName
	for _, arg := range args {
		if arg.IsOptional {
			useString += fmt.Sprintf(" [%s]", arg.ApiName)
		} else if arg.Type == ArgBoolT { // todo add shorthand flag in here if defined
			useString += fmt.Sprintf(" [--%s]", arg.ApiName)
		} else {
			useString += fmt.Sprintf(" <%s>", arg.ApiName)
		}
	}
	return strings.Trim(useString, " ")
}

func ShortDescription(metadata ScriptMetadata) string {
	description := metadata.OneLineDescription
	if description != nil {
		return *description
	} else {
		return ""
	}
}

func LongDescription(metadata ScriptMetadata) string {
	description := metadata.BlockDescription
	if description != nil {
		return *description
	} else {
		return ""
	}
}

func ExtractMetadata(statements []Stmt) ScriptMetadata {
	args := extractArgs(statements)
	oneLineDescription, blockDescription := extractDescriptions(statements)
	return ScriptMetadata{
		Args:               args,
		OneLineDescription: oneLineDescription,
		BlockDescription:   blockDescription,
	}
}

func extractDescriptions(statements []Stmt) (*string, *string) {
	fileHeader, ok := lo.Find(statements, func(stmt Stmt) bool {
		_, ok := stmt.(*FileHeader)
		return ok
	})

	if !ok {
		return nil, nil
	}

	fh := fileHeader.(*FileHeader)
	oneLiner := &fh.FhToken.OneLiner
	block := fh.FhToken.Rest
	return oneLiner, block
}

func extractArgs(statements []Stmt) []ScriptArg {
	var args []ScriptArg
	argBlockIfFound, ok := lo.Find(statements, func(stmt Stmt) bool {
		_, ok := stmt.(*ArgBlock)
		return ok
	})

	if !ok {
		return args
	}

	argBlock := argBlockIfFound.(*ArgBlock)
	for _, argStmt := range argBlock.Stmts {
		argDecl, ok := argStmt.(*ArgDeclaration)
		if ok {
			literalInterpreter := NewLiteralInterpreter(nil) // todo should probably not be nil, for erroring?
			arg := FromArgDecl(literalInterpreter, argDecl)
			args = append(args, *arg)
		}
	}

	return args
}
