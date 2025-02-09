package core

import (
	"fmt"

	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

const (
	FUNC_PRINT   = "print"
	FUNC_PPRINT  = "pprint"
	FUNC_DEBUG   = "debug"
	FUNC_LEN     = "len"
	FUNC_SORT    = "sort"
	FUNC_NOW     = "now"
	FUNC_TYPE_OF = "type_of"
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
				"reverse": {RslBoolT},
			},
			Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, namedArgs map[string]namedArg) []RslValue {
				reverseArg, exists := namedArgs["reverse"]
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
	}

	functions = append(functions, createColorFunctions()...)

	FunctionsByName = make(map[string]Func)
	for _, f := range functions {
		validateFunction(f)
		FunctionsByName[f.Name] = f
	}
}

func validateFunction(f Func) {
	if f.RequiredArgCount > len(f.ArgTypes) {
		panic(fmt.Sprintf("Bug! Function %q has more required args than arg types", f.Name))
	}
}

func createColorFunctions() []Func {
	colorStrs := COLOR_STRINGS
	funcs := make([]Func, len(colorStrs))
	for _, color := range colorStrs {
		funcs = append(funcs, Func{
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
		})
	}
	return funcs
}

func bugIncorrectTypes(funcName string) string {
	return fmt.Sprintf("Bug! Switch cases should line up with %q definition", funcName)
}
