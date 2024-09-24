package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type PickResource struct {
	Opts       []PickResourceOpt
	ReturnType RslTypeEnum
}

type PickResourceOpt struct {
	Match  []string
	Return []interface{}
}

type PickResourceSerde struct {
	Options []PickResourceOptionSerde `json:"options"`
}

type PickResourceOptionSerde struct {
	Match  []string      `json:"match"`
	Return []interface{} `json:"return"`
}

func LoadPickResource(i *MainInterpreter, function Token, jsonPath string, numExpectedReturnValues int) PickResource {
	finalPath := resolveFinalPath(jsonPath)
	file, err := os.Open(finalPath)
	if err != nil {
		i.error(function, fmt.Sprintf("Error opening file: %s", err))
	}
	defer file.Close()

	resource := PickResourceSerde{}
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&resource); err != nil {
		i.error(function, fmt.Sprintf("Error decoding JSON into pick resource: %s", err))
	}

	var opts []PickResourceOpt
	for _, option := range resource.Options {
		if len(option.Match) == 0 {
			i.error(function, "pick resource options must have at least one match value")
		}

		if len(option.Return) == 0 {
			i.error(function, "pick resource options must have at least one return value")
		}

		if numExpectedReturnValues != NO_NUM_RETURN_VALUES_CONSTRAINT && len(option.Return) != numExpectedReturnValues {
			i.error(function, fmt.Sprintf("Expected %d return values from resource option: %q", numExpectedReturnValues, option.Return))
		}

		opts = append(opts, PickResourceOpt{
			// todo we should probably do some type checking e.g. only array of primitives
			Match:  option.Match,
			Return: option.Return,
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
	RP.RadDebug(fmt.Sprintf("Joined %q and %q to get %q", ScriptDir, pathFromRslScript, finalPath))
	return finalPath
}
