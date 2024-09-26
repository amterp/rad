package core

import (
	"fmt"
)

func runPickFromResource(
	i *MainInterpreter,
	function Token,
	args []interface{},
	numExpectedReturnValues int,
) interface{} {
	numArgs := len(args)
	if numArgs < 1 {
		i.error(function, fmt.Sprintf("%s() takes at least one argument", PICK_FROM_RESOURCE))
	}

	if numArgs > 2 {
		i.error(function, fmt.Sprintf("%s() takes at most two arguments, got %v", PICK_FROM_RESOURCE, numArgs))
	}

	var stringFilter string
	switch numArgs {
	case 1:
		stringFilter = ""
	case 2:
		filter := args[1]
		switch filter.(type) {
		case string, int64, float64, bool:
			stringFilter = ToPrintable(filter)
		default:
			i.error(function, fmt.Sprintf("%s() does not allow arrays as filters", PICK_FROM_RESOURCE))
		}
	}

	jsonResourcePath, ok := args[0].(string)
	if !ok {
		i.error(function, fmt.Sprintf("%s() takes a string as the first argument", PICK_FROM_RESOURCE))
	}
	resource := LoadPickResource(i, function, jsonResourcePath, numExpectedReturnValues)
	return pickFromResource(i, function, stringFilter, resource)
}

func pickFromResource(i *MainInterpreter, function Token, filter string, resource PickResource) interface{} {
	var matchedOptions []PickResourceOpt
	for _, opt := range resource.Opts {
		for _, match := range opt.Match {
			if match == filter {
				matchedOptions = append(matchedOptions, opt)
				break
			}
		}
	}

	if len(matchedOptions) == 0 {
		// todo do we want users to be able to recover?
		i.error(function, fmt.Sprintf("Filtered %d options to 0 with filter: %q", len(resource.Opts), filter))
	}

	if len(matchedOptions) > 1 {
		// todo here we should launch into a huh.Select
		i.error(function, fmt.Sprintf("Filtered %d options to more than 1 with filter: %q", len(resource.Opts), filter))
	}

	returnValues := matchedOptions[0].Return
	if len(returnValues) == 1 {
		return returnValues[0]
	}
	return returnValues
}
