package core

import (
	"github.com/samber/lo"
)

type ScriptData struct {
	ScriptName         string
	Args               []ScriptArg
	OneLineDescription *string // todo not really using this atm, throw away? revisit this syntax
	BlockDescription   *string
	Instructions       []Stmt
}

func ExtractMetadata(instructions []Stmt) *ScriptData {
	args := extractArgs(instructions)
	oneLineDescription, blockDescription := extractDescriptions(instructions)
	return &ScriptData{
		ScriptName:         ScriptName,
		Args:               args,
		OneLineDescription: oneLineDescription,
		BlockDescription:   blockDescription,
		Instructions:       instructions,
	}
}

func extractDescriptions(instructions []Stmt) (*string, *string) {
	fileHeader, ok := lo.Find(instructions, func(stmt Stmt) bool {
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

func extractArgs(instructions []Stmt) []ScriptArg {
	var args []ScriptArg
	argBlockIfFound, ok := lo.Find(instructions, func(stmt Stmt) bool {
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
