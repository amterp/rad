package core

import "fmt"

// todo
//   - implement steps?
//   - somehow improve implementation to be a generator, rather than eagerly created list? chugs at e.g. 100_000
func runRange(i *MainInterpreter, function Token, args []interface{}) []interface{} {
	if len(args) < 1 || len(args) > 3 {
		i.error(function, RANGE+fmt.Sprintf("() takes 1 to 3 arguments, got %d", len(args)))
	}

	useFloats := false
	for _, arg := range args {
		switch arg.(type) {
		case float64:
			useFloats = true
		case int64:
		default:
			i.error(function, RANGE+fmt.Sprintf("() takes int or float arguments, got %s", TypeAsString(arg)))
		}
	}

	if useFloats {
		return runFloatRange(i, function, args)
	} else {
		return runIntRange(i, function, args)
	}
}

func runFloatRange(i *MainInterpreter, function Token, args []interface{}) []interface{} {
	var start, end, step float64
	switch len(args) {
	case 1:
		start = 0
		end = args[0].(float64)
		step = 1
	case 2:
		switch args[0].(type) {
		case float64:
			start = args[0].(float64)
		case int64:
			start = float64(args[0].(int64))
		}
		switch args[1].(type) {
		case float64:
			end = args[1].(float64)
		case int64:
			end = float64(args[1].(int64))
		}
		step = 1
	case 3:
		switch args[0].(type) {
		case float64:
			start = args[0].(float64)
		case int64:
			start = float64(args[0].(int64))
		}
		switch args[1].(type) {
		case float64:
			end = args[1].(float64)
		case int64:
			end = float64(args[1].(int64))
		}
		switch args[2].(type) {
		case float64:
			step = args[2].(float64)
		case int64:
			step = float64(args[2].(int64))
		}
	}

	if step == 0 {
		i.error(function, RANGE+fmt.Sprintf("() step argument cannot be zero"))
	}

	if start > end && step > 0 {
		i.error(function, RANGE+fmt.Sprintf("() start %f cannot be greater than end %f with positive step %f", start, end, step))
	}

	if start < end && step < 0 {
		i.error(function, RANGE+fmt.Sprintf("() start %f cannot be less than end %f with negative step %f", start, end, step))
	}

	var result []interface{}

	if step < 0 {
		for i := start; i > end; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i < end; i += step {
			result = append(result, i)
		}
	}

	return result
}

func runIntRange(i *MainInterpreter, function Token, args []interface{}) []interface{} {
	var start, end, step int64
	switch len(args) {
	case 1:
		start = 0
		end = args[0].(int64)
		step = 1
	case 2:
		start = args[0].(int64)
		end = args[1].(int64)
		step = 1
	case 3:
		start = args[0].(int64)
		end = args[1].(int64)
		step = args[2].(int64)
	}

	if step == 0 {
		i.error(function, RANGE+fmt.Sprintf("() step argument cannot be zero"))
	}

	if start > end && step > 0 {
		i.error(function, RANGE+fmt.Sprintf("() start %d cannot be greater than end %d with positive step %d", start, end, step))
	}

	if start < end && step < 0 {
		i.error(function, RANGE+fmt.Sprintf("() start %d cannot be less than end %d with negative step %d", start, end, step))
	}

	var result []interface{}

	if step < 0 {
		for i := start; i > end; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i < end; i += step {
			result = append(result, i)
		}
	}

	return result
}
