package core

import (
	"fmt"
	"os"
	"os/exec"
	com "rad/core/common"

	"github.com/amterp/rts/check"

	"github.com/amterp/rts"
)

func AddInternalFuncs() {
	functions := []BuiltInFunc{
		{
			Name:            INTERNAL_FUNC_GET_STASH_ID,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]RslTypeEnum{{RslStringT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				argNode := f.args[0]
				cmd := argNode.value.RequireStr(f.i, argNode.node).Plain()
				path, err := exec.LookPath(cmd)
				if err != nil {
					// todo return msg
					return newRslValues(f.i, f.callNode, RSL_NULL)
				}

				rslParser, err := rts.NewRslParser()
				if err != nil {
					// todo return msg
					return newRslValues(f.i, f.callNode, RSL_NULL)
				}

				src, err := readSource(path)
				if err != nil {
					// todo return msg
					return newRslValues(f.i, f.callNode, RSL_NULL)
				}
				tree := rslParser.Parse(src)
				_ = tree

				fh, ok := tree.FindFileHeader()
				if !ok {
					// todo return msg
					return newRslValues(f.i, f.callNode, RSL_NULL)
				}

				stashId, ok := fh.MetadataEntries[MACRO_STASH_ID]
				if ok {
					return newRslValues(f.i, f.callNode, stashId)
				}

				return newRslValues(f.i, f.callNode, RSL_NULL)
			},
		},
		{
			Name:            INTERNAL_FUNC_DELETE_STASH,
			ReturnValues:    ZERO_RETURN_VALS,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]RslTypeEnum{{RslStringT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				idArg := f.args[0]
				id := idArg.value.RequireStr(f.i, idArg.node).Plain()
				path := RadHomeInst.GetStashForId(id)
				RP.Printf("Deleting %s\n", path)
				err := os.RemoveAll(path)
				if err != nil {
					f.i.errorf(idArg.node, "Failed to delete stash: %s", err.Error())
				}
				return EMPTY
			},
		},
		{
			Name:            INTERNAL_FUNC_RUN_CHECK,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]RslTypeEnum{{RslStringT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				scriptArg := f.args[0]
				scriptPath := scriptArg.value.RequireStr(f.i, scriptArg.node).Plain()
				result := com.LoadFile(scriptPath)
				if result.Error != nil {
					// todo don't think we can point at the node -- it's an internal function. Generally true for embedded commands, actually
					f.i.errorf(scriptArg.node, "Failed to load script for checking: %s", result.Error.Error())
				}

				contents := result.Content
				checker, err := check.NewChecker()
				if err != nil {
					f.i.errorf(scriptArg.node, "Failed to create checker: %s", err.Error())
				}

				checker.UpdateSrc(contents)
				checkR, err := checker.CheckDefault()
				if err != nil {
					f.i.errorf(scriptArg.node, "Failed to run checker: %s", err.Error())
				}

				rslMap := NewRslMap()
				diagnostics := NewRslList()
				for _, diag := range checkR.Diagnostics {
					diagMap := NewRslMap()
					diagMap.SetPrimitiveStr("src", diag.RangedSrc)
					diagMap.SetPrimitiveStr("line_src", diag.LineSrc)

					diagMap.SetPrimitiveInt("start_line", diag.Range.Start.Line)
					diagMap.SetPrimitiveInt("start_char", diag.Range.Start.Character)
					diagMap.SetPrimitiveStr("pos", fmt.Sprintf("L%d:%d", diag.Range.Start.Line+1, diag.Range.Start.Character+1))

					diagMap.SetPrimitiveStr("severity", diag.Severity.String())
					diagMap.SetPrimitiveStr("msg", diag.Message)
					if diag.Code != nil {
						diagMap.SetPrimitiveStr("code", diag.Code.String())
					}
					diagnostics.Append(newRslValueMap(diagMap))
				}
				rslMap.SetPrimitiveList("diagnostics", diagnostics)

				return newRslValues(f.i, f.callNode, rslMap)
			},
		},
	}

	for _, f := range functions {
		FunctionsByName[f.Name] = f
	}
}
