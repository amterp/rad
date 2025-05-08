package core

import (
	"os"
	"os/exec"

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

				for key, value := range fh.MetadataEntries {
					if key == "stash_id" {
						switch value.(type) {
						case string, int, float64, bool:
							return newRslValues(f.i, f.callNode, value)
						default:
							// todo return msg
							return newRslValues(f.i, f.callNode, RSL_NULL)
						}
					}
				}

				return newRslValues(f.i, f.callNode, RSL_NULL)
			},
		},
		{
			Name:            INTERNAL_FUNC_DELETE_STASH,
			ReturnValues:    ZERO_RETURN_VALS,
			MinPosArgCount:  0,
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
	}

	for _, f := range functions {
		FunctionsByName[f.Name] = f
	}
}
