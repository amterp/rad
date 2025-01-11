package core

import (
	"fmt"

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

	// read out arguments
	for _, argStmt := range argBlock.Stmts {
		argDecl, ok := argStmt.(*ArgDeclaration)
		if ok {
			arg := FromArgDecl(literalInterpreter, argDecl)
			args[arg.Name] = arg
			orderedArgs = append(orderedArgs, arg)
		}
	}

	// now check for constraints
	for _, argStmt := range argBlock.Stmts {
		switch coerced := argStmt.(type) {
		case *ArgDeclaration:
			// already handled above
		case *ArgEnum:
			scriptArg, ok := args[coerced.Identifier.GetLexeme()]
			if !ok {
				RP.ErrorExit("Enum constraint applied to undeclared arg: " + coerced.Identifier.GetLexeme())
			}
			literal := coerced.Values.Accept(literalInterpreter)
			validValues, ok := literal.([]interface{})
			if !ok {
				RP.RadErrorExit(fmt.Sprintf("Bug! Parser should not have allowed a non-array enum declaration for arg (got %T): %s", literal, coerced.Identifier.GetLexeme()))
			}
			strArr, ok := AsStringArray(validValues)
			if !ok {
				RP.RadErrorExit("Bug! Parser should not have allowed a non-string enum declaration for arg: " + coerced.Identifier.GetLexeme())
			}
			scriptArg.EnumConstraint = &strArr
		case *ArgRegex:
			scriptArg, ok := args[coerced.Identifier.GetLexeme()]
			if !ok {
				RP.ErrorExit("Regex constraint applied to undeclared arg: " + coerced.Identifier.GetLexeme())
			}
			scriptArg.RegexConstraint = coerced.Regex
		default:
			RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled arg stmt type: %T", coerced))
		}
	}

	return orderedArgs
}
