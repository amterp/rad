package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

const (
	FUNC_PRINT              = "print"
	FUNC_PPRINT             = "pprint"
	FUNC_DEBUG              = "debug"
	FUNC_EXIT               = "exit"
	FUNC_SLEEP              = "sleep"
	FUNC_SEED_RANDOM        = "seed_random"
	FUNC_RAND               = "rand"
	FUNC_RAND_INT           = "rand_int"
	FUNC_REPLACE            = "replace"
	FUNC_LEN                = "len"
	FUNC_SORT               = "sort"
	FUNC_NOW                = "now"
	FUNC_TYPE_OF            = "type_of"
	FUNC_JOIN               = "join"
	FUNC_UPPER              = "upper"
	FUNC_LOWER              = "lower"
	FUNC_STARTS_WITH        = "starts_with"
	FUNC_ENDS_WITH          = "ends_with"
	FUNC_PICK               = "pick"
	FUNC_PICK_KV            = "pick_kv"
	FUNC_PICK_FROM_RESOURCE = "pick_from_resource"
	FUNC_KEYS               = "keys"
	FUNC_VALUES             = "values"
	FUNC_TRUNCATE           = "truncate"
	FUNC_SPLIT              = "split"
	FUNC_RANGE              = "range"
	FUNC_UNIQUE             = "unique"
	FUNC_CONFIRM            = "confirm"
	FUNC_PARSE_JSON         = "parse_json"
	FUNC_PARSE_INT          = "parse_int"
	FUNC_PARSE_FLOAT        = "parse_float"
	FUNC_HTTP_REQ           = "http_req"
	FUNC_ABS                = "abs"

	namedArgReverse = "reverse"
	namedArgTitle   = "title"
	namedArgPrompt  = "prompt"
)

var (
	NO_POS_ARGS   = [][]RslTypeEnum{}
	NO_NAMED_ARGS = map[string][]RslTypeEnum{}
)

type FuncInvocationArgs struct {
	i                  *Interpreter
	callNode           *ts.Node
	args               []positionalArg
	namedArgs          map[string]namedArg
	numExpectedOutputs int
}

func NewFuncInvocationArgs(i *Interpreter, callNode *ts.Node, args []positionalArg, namedArgs map[string]namedArg, numExpectedOutputs int) FuncInvocationArgs {
	return FuncInvocationArgs{
		i:                  i,
		callNode:           callNode,
		args:               args,
		namedArgs:          namedArgs,
		numExpectedOutputs: numExpectedOutputs,
	}
}

type Func struct {
	Name             string
	ReturnValues     []int
	RequiredArgCount int
	ArgTypes         [][]RslTypeEnum          // by index, what types are allowed for that index. empty == any
	NamedArgs        map[string][]RslTypeEnum // name -> allowed types
	// interpreter, callNode, positional args, named args
	// Guarantees when Execute invoked:
	// - given at least as many args as required (RequiredArgCount)
	// - not given more args than types have been defined for (ArgTypes)
	// - only valid named args are given (if given) (valid name, valid type) (NamedArgs)
	Execute func(FuncInvocationArgs) []RslValue
}

var FunctionsByName map[string]Func

func init() {
	functions := []Func{
		FuncPrint,
		FuncPPrint,
		FuncDebug,
		FuncExit,
		FuncSleep,
		FuncSeedRandom,
		FuncRand,
		FuncRandInt,
		FuncReplace,
		FuncPick,
		FuncPickKv,
		FuncPickFromResource,
		FuncSplit,
		FuncRange,
		{
			Name:             FUNC_LEN,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT, RslListT, RslMapT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				switch v := arg.value.Val.(type) {
				case RslString:
					return newRslValues(f.i, arg.node, v.Len())
				case *RslList:
					return newRslValues(f.i, arg.node, v.Len())
				case *RslMap:
					return newRslValues(f.i, arg.node, v.Len())
				default:
					panic(bugIncorrectTypes(FUNC_LEN))
				}
			},
		},
		{
			Name:             FUNC_SORT,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT, RslListT}},
			NamedArgs: map[string][]RslTypeEnum{
				namedArgReverse: {RslBoolT},
			},
			Execute: func(f FuncInvocationArgs) []RslValue {
				reverseArg, exists := f.namedArgs[namedArgReverse]
				reverse := false
				if exists {
					reverse = reverseArg.value.RequireBool(f.i, reverseArg.valueNode)
				}

				arg := f.args[0]
				switch coerced := arg.value.Val.(type) {
				case *RslList:
					sortedValues := sortList(f.i, arg.node, coerced, lo.Ternary(reverse, Desc, Asc))
					list := NewRslList()
					for _, v := range sortedValues {
						list.Append(v)
					}
					return newRslValues(f.i, arg.node, list)
				default:
					panic(bugIncorrectTypes(FUNC_SORT))
				}
			},
		},
		{
			Name:             FUNC_NOW,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 0,
			ArgTypes:         NO_POS_ARGS,
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				m := NewRslMap()
				m.SetPrimitiveStr("date", RClock.Now().Format("2006-01-02"))
				m.SetPrimitiveInt("year", RClock.Now().Year())
				m.SetPrimitiveInt("month", int(RClock.Now().Month()))
				m.SetPrimitiveInt("day", RClock.Now().Day())
				m.SetPrimitiveInt("hour", RClock.Now().Hour())
				m.SetPrimitiveInt("minute", RClock.Now().Minute())
				m.SetPrimitiveInt("second", RClock.Now().Second())

				epochM := NewRslMap()
				epochM.SetPrimitiveInt64("seconds", RClock.Now().Unix())
				epochM.SetPrimitiveInt64("millis", RClock.Now().UnixMilli())
				epochM.SetPrimitiveInt64("nanos", RClock.Now().UnixNano())

				m.SetPrimitiveMap("epoch", epochM)

				return newRslValues(f.i, f.callNode, m)
			},
		},
		{
			Name:             FUNC_TYPE_OF,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				return newRslValues(f.i, f.callNode, NewRslString(TypeAsString(f.args[0].value)))
			},
		},
		{
			Name:             FUNC_JOIN,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 2,
			ArgTypes:         [][]RslTypeEnum{{RslListT}, {RslStringT}, {RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				listArg := f.args[0]
				sepArg := f.args[1]
				prefixArg := tryGetArg(2, f.args)
				suffixArg := tryGetArg(3, f.args)

				list := listArg.value.RequireList(f.i, listArg.node)
				sep := sepArg.value.RequireStr(f.i, sepArg.node).String()
				prefix := ""
				if prefixArg != nil {
					prefix = prefixArg.value.RequireStr(f.i, prefixArg.node).String()
				}
				suffix := ""
				if suffixArg != nil {
					suffix = suffixArg.value.RequireStr(f.i, suffixArg.node).String()
				}

				return newRslValues(f.i, f.callNode, list.Join(sep, prefix, suffix))
			},
		},
		{
			Name:             FUNC_UPPER,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				return newRslValues(f.i, arg.node, arg.value.RequireStr(f.i, arg.node).Upper())
			},
		},
		{
			Name:             FUNC_LOWER,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				return newRslValues(f.i, arg.node, arg.value.RequireStr(f.i, arg.node).Lower())
			},
		},
		{
			Name:             FUNC_STARTS_WITH,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 2,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				subjectArg := f.args[0]
				prefixArg := f.args[1]
				subjectStr := subjectArg.value.RequireStr(f.i, subjectArg.node)
				prefixStr := prefixArg.value.RequireStr(f.i, prefixArg.node)
				return newRslValues(f.i, f.callNode, strings.HasPrefix(subjectStr.Plain(), prefixStr.Plain()))
			},
		},
		{
			Name:             FUNC_ENDS_WITH,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 2,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				subjectArg := f.args[0]
				prefixArg := f.args[1]
				subjectStr := subjectArg.value.RequireStr(f.i, subjectArg.node)
				prefixStr := prefixArg.value.RequireStr(f.i, prefixArg.node)
				return newRslValues(f.i, f.callNode, strings.HasSuffix(subjectStr.Plain(), prefixStr.Plain()))
			},
		},
		{
			Name:             FUNC_KEYS,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslMapT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				return newRslValues(f.i, arg.node, arg.value.RequireMap(f.i, arg.node).Keys())
			},
		},
		{
			Name:             FUNC_VALUES,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslMapT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				return newRslValues(f.i, arg.node, arg.value.RequireMap(f.i, arg.node).Values())
			},
		},
		{
			Name:             FUNC_TRUNCATE,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 2,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslIntT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				strArg := f.args[0]
				maxLenArg := f.args[1]
				maxLen := maxLenArg.value.RequireInt(f.i, maxLenArg.node)

				if maxLen < 0 {
					f.i.errorf(maxLenArg.node, "%s() takes a non-negative int, got %d", FUNC_TRUNCATE, maxLen)
				}

				rslStr := strArg.value.RequireStr(f.i, strArg.node)
				strLen := rslStr.Len()

				if maxLen >= strLen {
					return newRslValues(f.i, f.callNode, rslStr)
				}

				str := rslStr.Plain() // todo should maintain attributes

				if terminalSupportsUtf8 {
					str = str[:maxLen-1]
					str += "…"
				} else {
					str = str[:maxLen-3]
					str += "..."
				}
				return newRslValues(f.i, f.callNode, str)
			},
		},
		{
			Name:             FUNC_UNIQUE,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslListT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]

				output := NewRslList()

				seen := make(map[string]struct{})
				list := arg.value.RequireList(f.i, arg.node)
				for _, item := range list.Values {
					key := ToPrintable(item) // todo not a solid approach
					if _, exists := seen[key]; !exists {
						seen[key] = struct{}{}
						output.Append(item)
					}
				}

				return newRslValues(f.i, f.callNode, output)
			},
		},
		{
			Name:             FUNC_CONFIRM,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 0,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := tryGetArg(0, f.args)

				prompt := "Confirm? [y/n] > "

				if arg != nil {
					prompt = arg.value.RequireStr(f.i, arg.node).Plain()
				}

				var response string
				err := huh.NewInput().
					Prompt(prompt).
					Value(&response).
					Run()
				if err != nil {
					// todo I think this errors if user aborts
					f.i.errorf(f.callNode, fmt.Sprintf("Error reading input: %v", err))
				}

				return newRslValues(f.i, f.callNode, strings.HasPrefix(strings.ToLower(response), "y"))
			},
		},
		{
			Name:             FUNC_PARSE_JSON,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]

				out, err := TryConvertJsonToNativeTypes(f.i, f.callNode, arg.value.RequireStr(f.i, arg.node).Plain())
				if err != nil {
					f.i.errorf(f.callNode, fmt.Sprintf("Error parsing JSON: %v", err))
				}
				return newRslValues(f.i, f.callNode, out)
			},
		},
		{
			Name:             FUNC_PARSE_INT,
			ReturnValues:     UP_TO_TWO_RETURN_VALS,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]

				str := arg.value.RequireStr(f.i, arg.node).Plain()
				parsed, err := strconv.ParseInt(str, 10, 64)

				if err != nil {
					errMsg := fmt.Sprintf("%s() failed to parse %q", PARSE_INT, str)
					if f.numExpectedOutputs == 1 {
						f.i.errorf(f.callNode, errMsg) // todo when errors require codes, redo
						panic(UNREACHABLE)
					} else {
						return newRslValues(f.i, f.callNode, 0, ErrorRslMap(PARSE_INT_FAILED, errMsg))
					}
				} else {
					if f.numExpectedOutputs == 1 {
						return newRslValues(f.i, f.callNode, parsed)
					} else {
						return newRslValues(f.i, f.callNode, parsed, NoErrorRslMap())
					}
				}
			},
		},
		{
			Name:             FUNC_PARSE_FLOAT,
			ReturnValues:     UP_TO_TWO_RETURN_VALS,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]

				str := arg.value.RequireStr(f.i, arg.node).Plain()
				parsed, err := strconv.ParseFloat(str, 64)

				if err != nil {
					errMsg := fmt.Sprintf("%s() failed to parse %q", PARSE_FLOAT, str)
					if f.numExpectedOutputs == 1 {
						f.i.errorf(f.callNode, errMsg) // todo when errors require codes, redo
						panic(UNREACHABLE)
					} else {
						return newRslValues(f.i, f.callNode, 0, ErrorRslMap(PARSE_FLOAT_FAILED, errMsg))
					}
				} else {
					if f.numExpectedOutputs == 1 {
						return newRslValues(f.i, f.callNode, parsed)
					} else {
						return newRslValues(f.i, f.callNode, parsed, NoErrorRslMap())
					}
				}
			},
		},
	}

	functions = append(functions, createColorFunctions()...)

	FunctionsByName = make(map[string]Func)
	for _, f := range functions {
		validateFunction(f, FunctionsByName)
		FunctionsByName[f.Name] = f
	}
}

func validateFunction(f Func, functionsSoFar map[string]Func) {
	if f.RequiredArgCount > len(f.ArgTypes) {
		panic(fmt.Sprintf("Bug! Function %q has more required args than arg types", f.Name))
	}

	if _, exists := functionsSoFar[f.Name]; exists {
		panic(fmt.Sprintf("Bug! Function %q already exists", f.Name))
	}
}

func createColorFunctions() []Func {
	colorStrs := lo.Values(colorEnumToStrings)
	funcs := make([]Func, len(colorStrs))
	for idx, color := range colorStrs {
		funcs[idx] = Func{
			Name:         color,
			ReturnValues: ONE_RETURN_VAL,
			ArgTypes:     [][]RslTypeEnum{{}},
			NamedArgs:    NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				clr := ColorFromString(f.i, f.callNode, color)
				arg := f.args[0]
				switch coerced := arg.value.Val.(type) {
				case RslString:
					return newRslValues(f.i, arg.node, coerced.Color(clr))
				default:
					s := NewRslString(ToPrintable(arg))
					s.SetSegmentsColor(clr)
					return newRslValues(f.i, f.callNode, s)
				}
			},
		}
	}
	return funcs
}

func tryGetArg(idx int, args []positionalArg) *positionalArg {
	if idx >= len(args) {
		return nil
	}
	return &args[idx]
}

func bugIncorrectTypes(funcName string) string {
	return fmt.Sprintf("Bug! Switch cases should line up with %q definition", funcName)
}
