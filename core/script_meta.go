package core

import (
	"github.com/samber/lo"
)

type ScriptData struct {
	ScriptName         string
	Args               []*ScriptArg
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

func extractArgs(instructions []Stmt) []*ScriptArg {
	args := make(map[string]*ScriptArg)
	var orderedArgs []*ScriptArg

	argBlockIfFound, ok := lo.Find(instructions, func(stmt Stmt) bool {
		_, ok := stmt.(*ArgBlock)
		return ok
	})

	if !ok {
		return []*ScriptArg{}
	}

	literalInterpreter := NewLiteralInterpreter(nil) // todo should probably not be nil, for erroring?

	argBlock := argBlockIfFound.(*ArgBlock)
	for _, argStmt := range argBlock.Stmts {
		argDecl, ok := argStmt.(*ArgDeclaration)
		if ok {
			arg := FromArgDecl(literalInterpreter, argDecl)
			args[arg.Name] = arg
			orderedArgs = append(orderedArgs, arg)
		}
	}

	for _, argStmt := range argBlock.Stmts {
		enumConstraint, ok := argStmt.(*ArgEnum)
		if ok {
			scriptArg, ok := args[enumConstraint.Identifier.GetLexeme()]
			if !ok {
				RP.ErrorExit("Enum constraint applied to undeclared arg: " + enumConstraint.Identifier.GetLexeme())
			}
			literal := enumConstraint.Values.Accept(literalInterpreter)
			switch coerced := literal.(type) {
			case []interface{}:
				strArr, ok := AsStringArray(coerced)
				if !ok {
					RP.RadErrorExit("Bug! Parser should not have allowed a non-string enum declaration for arg: " + enumConstraint.Identifier.GetLexeme())
				}
				scriptArg.EnumConstraint = &strArr
			}
		}
	}

	return orderedArgs
}
