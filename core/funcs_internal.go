package core

import (
	"fmt"
	"os"
	"os/exec"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/check"
	radfmt "github.com/amterp/rad/rts/radfmt"
	"github.com/amterp/rad/rts/rl"

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
					f.i.emitErrorf(rl.ErrFileWrite, idArg.node, "Failed to delete stash: %s", err.Error())
				}
				return VOID_SENTINEL
			},
		},
		{
			Name: INTERNAL_FUNC_RUN_CHECK,
			Execute: func(f FuncInvocation) RadValue {
				scriptPath := f.GetStr("_script").Plain()
				if !com.IsRegularFile(scriptPath) {
					f.i.emitErrorf(rl.ErrFileRead, f.callNode, "Cannot check '%s': not a regular file", scriptPath)
				}
				result := com.LoadFile(scriptPath)
				if result.Error != nil {
					// todo don't think we can point at the node -- it's an internal function. Generally true for embedded commands, actually
					f.i.emitErrorf(rl.ErrFileRead, f.callNode, "Failed to load script for checking: %s", result.Error.Error())
				}

				contents := NormalizeLineEndings(result.Content)
				checker, err := check.NewChecker()
				if err != nil {
					f.i.emitErrorf(rl.ErrGenericRuntime, f.callNode, "Failed to create checker: %s", err.Error())
				}

				checker.SetStrict(f.GetBool("_strict"))

				checker.UpdateSrc(contents)

				checkR, err := checker.Check()
				if err != nil {
					f.i.emitErrorf(rl.ErrGenericRuntime, f.callNode, "Failed to run checker: %s", err.Error())
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
					if diag.Suggestion != nil && *diag.Suggestion != "" {
						diagMap.SetPrimitiveStr("suggestion", *diag.Suggestion)
					}
					diagnostics.Append(newRadValueMap(diagMap))
				}
				radMap.SetPrimitiveList("diagnostics", diagnostics)

				return newRadValues(f.i, f.callNode, radMap)
			},
		},
		{
			Name: INTERNAL_FUNC_FMT,
			Execute: func(f FuncInvocation) RadValue {
				// Format normalizes line endings itself, so pass the raw source.
				formatted, changed, ok := radfmt.Format(f.GetStr("_src").Plain())

				radMap := NewRadMap()
				radMap.SetPrimitiveStr("formatted", formatted)
				radMap.SetPrimitiveBool("changed", changed)
				radMap.SetPrimitiveBool("ok", ok)
				return newRadValues(f.i, f.callNode, radMap)
			},
		},
		FuncInternalCheckFromLogs,
		{
			Name: INTERNAL_FUNC_EXPLAIN,
			Execute: func(f FuncInvocation) RadValue {
				codeArg := f.args[0]
				code := codeArg.value.RequireStr(f.i, codeArg.node).Plain()

				doc := GetErrorDoc(code)
				if doc == "" {
					return RAD_NULL_VAL
				}

				rendered := RenderMarkdownForTerminal(doc)
				return newRadValues(f.i, f.callNode, rendered)
			},
		},
		{
			Name: INTERNAL_FUNC_EXPLAIN_LIST,
			Execute: func(f FuncInvocation) RadValue {
				codes := ListErrorCodes()
				list := NewRadList()
				for _, code := range codes {
					list.Append(newRadValueStr(code))
				}
				return newRadValues(f.i, f.callNode, list)
			},
		},
		{
			Name: INTERNAL_FUNC_DOCS_TOC,
			Execute: func(f FuncInvocation) RadValue {
				return newRadValues(f.i, f.callNode, NewRadString(BuildDocsTOC()))
			},
		},
		{
			Name: INTERNAL_FUNC_DOCS_GET,
			Execute: func(f FuncInvocation) RadValue {
				slug := f.GetStr("_slug").Plain()
				content, ok := GetDocTopic(slug)
				if !ok {
					return RAD_NULL_VAL
				}
				return newRadValues(f.i, f.callNode, NewRadString(content))
			},
		},
		{
			Name: INTERNAL_FUNC_DOCS_FULL,
			Execute: func(f FuncInvocation) RadValue {
				return newRadValues(f.i, f.callNode, NewRadString(BuildDocsFull()))
			},
		},
		{
			Name: INTERNAL_FUNC_DOCS_SLUGS,
			Execute: func(f FuncInvocation) RadValue {
				list := NewRadList()
				for _, slug := range GetDocSlugs() {
					list.Append(newRadValueStr(slug))
				}
				return newRadValues(f.i, f.callNode, list)
			},
		},
		{
			// Single rendering gate for `rad docs` output: raw markdown
			// when piped (agent capture), pretty-rendered on a TTY
			// (human at the terminal), with explicit overrides. Keeping
			// it here means the data funcs above stay pure (raw md).
			Name: INTERNAL_FUNC_RENDER,
			Execute: func(f FuncInvocation) RadValue {
				md := f.GetStr("_md").Plain()
				mode := f.GetStr("_mode").Plain()
				render := com.IsTty
				switch mode {
				case "raw":
					render = false
				case "render":
					render = true
				}
				if render {
					return newRadValues(f.i, f.callNode, RenderMarkdownForTerminal(md))
				}
				return newRadValues(f.i, f.callNode, NewRadString(md))
			},
		},
	}

	for _, f := range functions {
		f.Signature = rts.GetSignature(f.Name)
		FunctionsByName[f.Name] = f
	}
}
