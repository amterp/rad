package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	ts "github.com/tree-sitter/go-tree-sitter"
)

// todo In Go you can apparently define custom json serde with e.g. MarshalJSON and UnmarshalJSON,
//   see if can leverage it

type PickResource struct {
	Opts []PickResourceOpt
}

type PickResourceOpt struct {
	Keys   []string
	Values []RslValue
}

type PickResourceSerde struct {
	Options []PickResourceOptionSerde `json:"options"`
}

type PickResourceOptionSerde struct {
	Keys   []string      `json:"keys"`
	Values []interface{} `json:"values"`
}

func LoadPickResource(i *Interpreter, callNode *ts.Node, jsonPath string, numExpectedReturnValues int) PickResource {
	finalPath := resolveFinalPath(jsonPath)
	file, err := os.Open(finalPath)
	if err != nil {
		i.errorf(callNode, "Error opening file: %s", err)
	}
	defer file.Close()

	resource := PickResourceSerde{}
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&resource); err != nil {
		i.errorf(callNode, "Error decoding JSON into pick resource: %s", err)
	}

	var opts []PickResourceOpt
	for _, option := range resource.Options {
		if len(option.Keys) == 0 {
			i.errorf(callNode, "pick resource options must have at least one key")
		}

		if len(option.Values) == 0 {
			i.errorf(callNode, "pick resource options must have at least one value")
		}

		if numExpectedReturnValues != NO_NUM_RETURN_VALUES_CONSTRAINT && len(option.Values) != numExpectedReturnValues {
			i.errorf(callNode, "Expected %d return values from resource option: %q", numExpectedReturnValues, option.Values)
		}

		opts = append(opts, PickResourceOpt{
			// todo we should probably do some type checking e.g. only array of primitives
			Keys:   option.Keys,
			Values: ConvertValuesToNativeTypes(i, callNode, option.Values),
		})
	}
	return PickResource{
		Opts: opts,
	}
}

func resolveFinalPath(pathFromRslScript string) string {
	if filepath.IsAbs(pathFromRslScript) {
		return pathFromRslScript
	}

	finalPath := filepath.Clean(filepath.Join(ScriptDir, pathFromRslScript))
	RP.RadDebugf(fmt.Sprintf("Joined %q and %q to get %q", ScriptDir, pathFromRslScript, finalPath))
	return finalPath
}
