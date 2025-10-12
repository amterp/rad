package core

import (
	"fmt"
	"os"
	"os/exec"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts/check"

	"github.com/amterp/rad/rts"
)

func AddInternalFuncs() {
	functions := []BuiltInFunc{
		{
			Name: INTERNAL_FUNC_GET_STASH_ID,
			Execute: func(f FuncInvocation) RadValue {
				argNode := f.args[0]
				cmd := argNode.value.RequireStr(f.i, argNode.node).Plain()
				path, err := exec.LookPath(cmd)
				if err != nil {
					// todo return msg
					return newRadValues(f.i, f.callNode, RAD_NULL)
				}

				radParser, err := rts.NewRadParser()
				if err != nil {
					// todo return msg
					return newRadValues(f.i, f.callNode, RAD_NULL)
				}

				src, err := readSource(path)
				if err != nil {
					// todo return msg
					return newRadValues(f.i, f.callNode, RAD_NULL)
				}
				tree := radParser.Parse(src)
				_ = tree

				fh, ok := tree.FindFileHeader()
				if !ok {
					// todo return msg
					return newRadValues(f.i, f.callNode, RAD_NULL)
				}

				stashId, ok := fh.MetadataEntries[MACRO_STASH_ID]
				if ok {
					return newRadValues(f.i, f.callNode, stashId)
				}

				return newRadValues(f.i, f.callNode, RAD_NULL)
			},
		},
		{
			Name: INTERNAL_FUNC_DELETE_STASH,
			Execute: func(f FuncInvocation) RadValue {
				idArg := f.args[0]
				id := idArg.value.RequireStr(f.i, idArg.node).Plain()
				path := RadHomeInst.GetStashForId(id)
				RP.Printf("Deleting %s\n", path)
				err := os.RemoveAll(path)
				if err != nil {
					f.i.errorf(idArg.node, "Failed to delete stash: %s", err.Error())
				}
				return VOID_SENTINEL
			},
		},
		{
			Name: INTERNAL_FUNC_RUN_CHECK,
			Execute: func(f FuncInvocation) RadValue {
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

				radMap := NewRadMap()
				diagnostics := NewRadList()
				for _, diag := range checkR.Diagnostics {
					diagMap := NewRadMap()
					diagMap.SetPrimitiveStr("src", diag.RangedSrc)
					diagMap.SetPrimitiveStr("line_src", diag.LineSrc)

					diagMap.SetPrimitiveInt("start_line", diag.Range.Start.Line)
					diagMap.SetPrimitiveInt("start_char", diag.Range.Start.Character)
					diagMap.SetPrimitiveStr(
						"pos",
						fmt.Sprintf("L%d:%d", diag.Range.Start.Line+1, diag.Range.Start.Character+1),
					)

					diagMap.SetPrimitiveStr("severity", diag.Severity.String())
					diagMap.SetPrimitiveStr("msg", diag.Message)
					if diag.Code != nil {
						diagMap.SetPrimitiveStr("code", diag.Code.String())
					}
					diagnostics.Append(newRadValueMap(diagMap))
				}
				radMap.SetPrimitiveList("diagnostics", diagnostics)

				return newRadValues(f.i, f.callNode, radMap)
			},
		},
		FuncInternalCheckFromLogs,
	}

	for _, f := range functions {
		f.Signature = rts.GetSignature(f.Name)
		FunctionsByName[f.Name] = f
	}
}
