package core

import (
	"fmt"
	"strings"

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

	namedArgReverse = "reverse"
	namedArgTitle   = "title"
	namedArgPrompt  = "prompt"
)

var (
	NO_POS_ARGS   = [][]RslTypeEnum{}
	NO_NAMED_ARGS = map[string][]RslTypeEnum{}
)

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
	Execute func(*Interpreter, *ts.Node, []positionalArg, map[string]namedArg) []RslValue
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
		//FuncPickFromResource,
		{
			Name:             FUNC_LEN,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT, RslListT, RslMapT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
				arg := args[0]
				switch v := arg.value.Val.(type) {
				case RslString:
					return newRslValues(i, arg.node, v.Len())
				case *RslList:
					return newRslValues(i, arg.node, v.Len())
				case *RslMap:
					return newRslValues(i, arg.node, v.Len())
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
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, namedArgs map[string]namedArg) []RslValue {
				reverseArg, exists := namedArgs[namedArgReverse]
				reverse := false
				if exists {
					reverse = reverseArg.value.RequireBool(i, reverseArg.valueNode)
				}

				arg := args[0]
				switch coerced := arg.value.Val.(type) {
				case *RslList:
					sortedValues := sortList(i, arg.node, coerced, lo.Ternary(reverse, Desc, Asc))
					list := NewRslList()
					for _, v := range sortedValues {
						list.Append(v)
					}
					return newRslValues(i, arg.node, list)
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
			Execute: func(i *Interpreter, callNode *ts.Node, _ []positionalArg, _ map[string]namedArg) []RslValue {
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

				return newRslValues(i, callNode, m)
			},
		},
		{
			Name:             FUNC_TYPE_OF,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
				return newRslValues(i, callNode, NewRslString(TypeAsString(args[0].value)))
			},
		},
		{
			Name:             FUNC_JOIN,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 2,
			ArgTypes:         [][]RslTypeEnum{{RslListT}, {RslStringT}, {RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
				listArg := args[0]
				sepArg := args[1]
				prefixArg := tryGetArg(2, args)
				suffixArg := tryGetArg(3, args)

				list := listArg.value.RequireList(i, listArg.node)
				sep := sepArg.value.RequireStr(i, sepArg.node).String()
				prefix := ""
				if prefixArg != nil {
					prefix = prefixArg.value.RequireStr(i, prefixArg.node).String()
				}
				suffix := ""
				if suffixArg != nil {
					suffix = suffixArg.value.RequireStr(i, suffixArg.node).String()
				}

				return newRslValues(i, callNode, list.Join(sep, prefix, suffix))
			},
		},
		{
			Name:             FUNC_UPPER,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
				arg := args[0]
				return newRslValues(i, arg.node, arg.value.RequireStr(i, arg.node).Upper())
			},
		},
		{
			Name:             FUNC_LOWER,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
				arg := args[0]
				return newRslValues(i, arg.node, arg.value.RequireStr(i, arg.node).Lower())
			},
		},
		{
			Name:             FUNC_STARTS_WITH,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 2,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
				subjectArg := args[0]
				prefixArg := args[1]
				subjectStr := subjectArg.value.RequireStr(i, subjectArg.node)
				prefixStr := prefixArg.value.RequireStr(i, prefixArg.node)
				return newRslValues(i, callNode, strings.HasPrefix(subjectStr.Plain(), prefixStr.Plain()))
			},
		},
		{
			Name:             FUNC_ENDS_WITH,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 2,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
				subjectArg := args[0]
				prefixArg := args[1]
				subjectStr := subjectArg.value.RequireStr(i, subjectArg.node)
				prefixStr := prefixArg.value.RequireStr(i, prefixArg.node)
				return newRslValues(i, callNode, strings.HasSuffix(subjectStr.Plain(), prefixStr.Plain()))
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
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
				clr := ColorFromString(i, callNode, color)
				arg := args[0]
				switch coerced := arg.value.Val.(type) {
				case RslString:
					return newRslValues(i, arg.node, coerced.Color(clr))
				default:
					s := NewRslString(ToPrintable(arg))
					s.SetSegmentsColor(clr)
					return newRslValues(i, callNode, s)
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
