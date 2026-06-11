package core

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/amterp/rad/rts/rl"
)

var FuncToJson = BuiltInFunc{
	Name: FUNC_TO_JSON,
	Execute: func(f FuncInvocation) RadValue {
		indent := f.GetInt("indent")
		if indent < 0 {
			return f.ReturnErrf(rl.ErrNumInvalidRange, "Indent must be non-negative, got %d", indent)
		}

		jsonStruct := RadToJsonType(f.GetArg("_val"))

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		// json.Marshal escapes <, >, & for HTML embedding by default - noise
		// for a scripting language's JSON output (pprint also disables it).
		enc.SetEscapeHTML(false)
		if indent > 0 {
			enc.SetIndent("", strings.Repeat(" ", int(indent)))
		}
		if err := enc.Encode(jsonStruct); err != nil {
			return f.ReturnErrf(rl.ErrInternalBug, "Failed to serialize to JSON: %v", err)
		}

		return f.Return(strings.TrimSuffix(buf.String(), "\n"))
	},
}
