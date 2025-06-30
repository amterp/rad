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
	Values []RadValue
}

type PickResourceSerde struct {
	Options []PickResourceOptionSerde `json:"options"`
}

type PickResourceOptionSerde struct {
	Keys   []string      `json:"keys"`
	Values []interface{} `json:"values"`
}

func LoadPickResource(i *Interpreter, callNode *ts.Node, jsonPath string) (PickResource, *RadError) {
	finalPath := resolveFinalPath(jsonPath)
	file, err := os.Open(finalPath)
	if err != nil {
		return PickResource{}, NewErrorStrf("Error opening file: %s", err)
	}
	defer file.Close()

	resource := PickResourceSerde{}
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&resource); err != nil {
		return PickResource{}, NewErrorStrf("Error decoding JSON into pick resource: %s", err)
	}

	var opts []PickResourceOpt
	for _, option := range resource.Options {
		if len(option.Keys) == 0 {
			return PickResource{}, NewErrorStrf("pick resource options must have at least one key")
		}

		if len(option.Values) == 0 {
			return PickResource{}, NewErrorStrf("pick resource options must have at least one value")
		}

		opts = append(opts, PickResourceOpt{
			// todo we should probably do some type checking e.g. only array of primitives
			Keys:   option.Keys,
			Values: ConvertValuesToNativeTypes(i, callNode, option.Values),
		})
	}
	return PickResource{
		Opts: opts,
	}, nil
}

func resolveFinalPath(pathFromRadScript string) string {
	if filepath.IsAbs(pathFromRadScript) {
		return pathFromRadScript
	}

	finalPath := filepath.Clean(filepath.Join(ScriptDir, pathFromRadScript))
	RP.RadDebugf(fmt.Sprintf("Joined %q and %q to get %q", ScriptDir, pathFromRadScript, finalPath))
	return finalPath
}
