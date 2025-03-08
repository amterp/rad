package core

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	com "rad/core/common"
	"sort"
	"strconv"
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
	FUNC_KEYS               = "keys"
	FUNC_VALUES             = "values"
	FUNC_TRUNCATE           = "truncate"
	FUNC_SPLIT              = "split"
	FUNC_RANGE              = "range"
	FUNC_UNIQUE             = "unique"
	FUNC_CONFIRM            = "confirm"
	FUNC_INPUT              = "input"
	FUNC_PARSE_JSON         = "parse_json"
	FUNC_PARSE_INT          = "parse_int"
	FUNC_PARSE_FLOAT        = "parse_float"
	FUNC_HTTP_GET           = "http_get"
	FUNC_HTTP_POST          = "http_post"
	FUNC_HTTP_PUT           = "http_put"
	FUNC_HTTP_PATCH         = "http_patch"
	FUNC_HTTP_DELETE        = "http_delete"
	FUNC_HTTP_HEAD          = "http_head"
	FUNC_HTTP_OPTIONS       = "http_options"
	FUNC_HTTP_TRACE         = "http_trace"
	FUNC_HTTP_CONNECT       = "http_connect"
	FUNC_ABS                = "abs"
	FUNC_GET_PATH           = "get_path"
	FUNC_FIND_PATHS         = "find_paths"
	FUNC_COUNT              = "count"
	FUNC_ZIP                = "zip"
	FUNC_STR                = "str"
	FUNC_SUM                = "sum"
	FUNC_TRIM               = "trim"
	FUNC_TRIM_PREFIX        = "trim_prefix"
	FUNC_TRIM_SUFFIX        = "trim_suffix"
	FUNC_READ_FILE          = "read_file"
	FUNC_ROUND              = "round"
	FUNC_CEIL               = "ceil"
	FUNC_FLOOR              = "floor"
	FUNC_MIN                = "min"
	FUNC_MAX                = "max"
	FUNC_CLAMP              = "clamp"
	FUNC_REVERSE            = "reverse"

	namedArgReverse  = "reverse"
	namedArgTitle    = "title"
	namedArgPrompt   = "prompt"
	namedArgHeaders  = "headers"
	namedArgBody     = "body"
	namedArgHint     = "hint"
	namedArgDefault  = "default"
	namedArgEnd      = "end"
	namedArgSep      = "sep"
	namedArgFill     = "fill"
	namedArgStrict   = "strict"
	namedArgMode     = "mode"
	namedArgDepth    = "depth"
	namedArgRelative = "relative"

	constContent   = "content"
	constSizeBytes = "size_bytes"
	constText      = "text"
	constBytes     = "bytes"
	constCode      = "code"
	constMsg       = "msg"
	constTarget    = "target"
	constCwd       = "cwd"
	constAbsolute  = "absolute"
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

// todo add 'usage' to each function? self-documenting errors when incorrectly using
type Func struct {
	Name             string
	ReturnValues     []int
	RequiredArgCount int
	ArgTypes         [][]RslTypeEnum          // by index, what types are allowed for that index. empty == any
	NamedArgs        map[string][]RslTypeEnum // name -> allowed types. empty == any
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
				case RslString:
					// todo maintain attributes
					str := f.i.evaluate(arg.node, 1)[0].RequireStr(f.i, f.callNode).Plain()
					runes := []rune(str)
					sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
					return newRslValues(f.i, f.callNode, string(runes))
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
				str = Truncate(str, maxLen)

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

				response, err := InputConfirm("", prompt)
				if err != nil {
					// todo I think this errors if user aborts
					f.i.errorf(f.callNode, fmt.Sprintf("Error reading input: %v", err))
				}

				return newRslValues(f.i, f.callNode, response)
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

				if err == nil {
					if f.numExpectedOutputs == 1 {
						return newRslValues(f.i, f.callNode, parsed)
					} else {
						return newRslValues(f.i, f.callNode, parsed, NewRslMap())
					}
				} else {
					errMsg := fmt.Sprintf("%s() failed to parse %q", FUNC_PARSE_INT, str)
					if f.numExpectedOutputs == 1 {
						// todo when errors require codes, redo
						f.i.errorf(f.callNode, errMsg)
						panic(UNREACHABLE)
					} else {
						return newRslValues(f.i, f.callNode, 0, ErrorRslMap(com.ErrParseIntFailed, errMsg))
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

				if err == nil {
					if f.numExpectedOutputs == 1 {
						return newRslValues(f.i, f.callNode, parsed)
					} else {
						return newRslValues(f.i, f.callNode, parsed, NewRslMap())
					}
				} else {
					errMsg := fmt.Sprintf("%s() failed to parse %q", FUNC_PARSE_FLOAT, str)
					if f.numExpectedOutputs == 1 {
						// todo when errors require codes, redo
						f.i.errorf(f.callNode, errMsg)
						panic(UNREACHABLE)
					} else {
						return newRslValues(f.i, f.callNode, 0, ErrorRslMap(com.ErrParseFloatFailed, errMsg))
					}
				}
			},
		},
		{
			Name:             FUNC_ABS,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslFloatT, RslIntT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]

				switch coerced := arg.value.Val.(type) {
				case int64:
					return newRslValues(f.i, f.callNode, AbsInt(coerced))
				case float64:
					return newRslValues(f.i, f.callNode, AbsFloat(coerced))
				default:
					bugIncorrectTypes(FUNC_ABS)
					panic(UNREACHABLE)
				}
			},
		},
		{
			Name:             FUNC_INPUT,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 0,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs: map[string][]RslTypeEnum{
				namedArgHint:    {RslStringT},
				namedArgDefault: {RslStringT},
			},
			Execute: func(f FuncInvocationArgs) []RslValue {
				prompt := "> "
				if promptArg := tryGetArg(0, f.args); promptArg != nil {
					prompt = promptArg.value.RequireStr(f.i, promptArg.node).Plain()
				}

				hint := ""
				if hintArg, exists := f.namedArgs[namedArgHint]; exists {
					hint = hintArg.value.RequireStr(f.i, hintArg.valueNode).Plain()
				}

				default_ := ""
				if defaultArg, exists := f.namedArgs[namedArgDefault]; exists {
					default_ = defaultArg.value.RequireStr(f.i, defaultArg.valueNode).Plain()
				}

				response, err := InputText(prompt, hint, default_)
				if err != nil {
					f.i.errorf(f.callNode, fmt.Sprintf("Error reading input: %v", err))
				}
				return newRslValues(f.i, f.callNode, response)
			},
		},
		{
			Name:             FUNC_GET_PATH,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				pathArg := f.args[0]
				path := pathArg.value.RequireStr(f.i, pathArg.node).Plain()

				rslMap := NewRslMap()

				stat, err1 := os.Stat(path) // todo should be abstracted away for testing
				absPath, err2 := filepath.Abs(path)
				if err1 == nil && err2 == nil {
					rslMap.SetPrimitiveStr("full_path", absPath)
					rslMap.SetPrimitiveStr("base_name", stat.Name())
					rslMap.SetPrimitiveStr("permissions", stat.Mode().Perm().String())
					fileType := lo.Ternary(stat.IsDir(), "dir", "file")
					rslMap.SetPrimitiveStr("type", fileType)
					if fileType == "file" {
						rslMap.SetPrimitiveInt64("size_bytes", stat.Size())
					}
				}

				return newRslValues(f.i, f.callNode, rslMap)
			},
		},
		{
			Name:             FUNC_FIND_PATHS,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs: map[string][]RslTypeEnum{
				// todo: filtering by name, file type
				//  potentially allow `include_root`
				namedArgDepth:    {RslIntT},
				namedArgRelative: {RslStringT},
			},
			Execute: func(f FuncInvocationArgs) []RslValue {
				pathArg := f.args[0]
				pathStr := pathArg.value.RequireStr(f.i, pathArg.node).Plain()

				depth := int64(-1) // -1 is unlimited
				if depthArg, exists := f.namedArgs[namedArgDepth]; exists {
					depth = depthArg.value.RequireInt(f.i, depthArg.valueNode)
				}

				relativeMode := constTarget
				if relativeArg, exists := f.namedArgs[namedArgRelative]; exists {
					relativeMode = relativeArg.value.RequireStr(f.i, relativeArg.valueNode).Plain()
				}

				absTarget, err := filepath.Abs(pathStr) // todo should be abstracted away for testing
				if err != nil {
					f.i.errorf(f.callNode, "Error resolving absolute path for target: %v", err)
				}

				list := NewRslList()
				err = filepath.WalkDir(absTarget, func(currPath string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}

					pathRelativeToTarget, err := filepath.Rel(absTarget, currPath)
					if err != nil {
						return err
					}

					if pathRelativeToTarget == "." {
						// don't include root
						return nil
					}

					if depth >= 0 && int64(strings.Count(pathRelativeToTarget, string(os.PathSeparator))) >= depth {
						if d.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}

					var formattedPath string
					switch relativeMode {
					case constTarget:
						formattedPath = pathRelativeToTarget
					case constCwd:
						cwd, err := os.Getwd() // todo should be abstracted away for testing
						if err != nil {
							return err
						}
						relToCwd, err := filepath.Rel(cwd, currPath)
						if err != nil {
							return err
						}
						formattedPath = relToCwd
					case constAbsolute:
						absPath, err := filepath.Abs(currPath)
						if err != nil {
							return err
						}
						formattedPath = absPath
					default:
						f.i.errorf(f.callNode, "Invalid target mode %q. Allowed: %v",
							relativeMode, []string{constTarget, constCwd, constAbsolute})
					}
					list.Append(newRslValueStr(formattedPath))
					return nil
				})

				if err != nil {
					f.i.errorf(f.callNode, "Error walking directory: %v", err)
				}

				return newRslValues(f.i, f.callNode, list)
			},
		},
		{
			Name:             FUNC_COUNT,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 2,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				strArg := f.args[0]
				substrArg := f.args[1]

				str := strArg.value.RequireStr(f.i, strArg.node).Plain()
				substr := substrArg.value.RequireStr(f.i, substrArg.node).Plain()

				count := strings.Count(str, substr)
				return newRslValues(f.i, f.callNode, count)
			},
		},
		{
			Name:             FUNC_ZIP,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 0,
			// TODO RAD-167 make truly unlimited
			ArgTypes: [][]RslTypeEnum{{RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}, {RslListT}},
			NamedArgs: map[string][]RslTypeEnum{
				namedArgFill:   {},
				namedArgStrict: {RslBoolT},
			},
			Execute: func(f FuncInvocationArgs) []RslValue {
				strictArg, strictExists := f.namedArgs[namedArgStrict]
				strict := false
				if strictExists {
					strict = strictArg.value.RequireBool(f.i, strictArg.valueNode)
				}

				fillArg, fillExists := f.namedArgs[namedArgFill]
				var fill *RslValue
				if fillExists {
					fill = &fillArg.value
				}

				if strictExists && fillExists {
					f.i.errorf(f.callNode, "Cannot specify both 'strict' and 'fill' named arguments")
				}

				if len(f.args) == 0 {
					return newRslValues(f.i, f.callNode, NewRslList())
				}

				length := int64(-1)
				for _, argList := range f.args {
					list := argList.value.RequireList(f.i, argList.node)
					if length == -1 {
						length = list.Len()
					} else if length != list.Len() {
						if strict {
							f.i.errorf(f.callNode, "Strict mode enabled: all lists must have the same length, but got %d and %d", length, list.Len())
						}
						if fill == nil {
							length = com.Int64Min(length, list.Len())
						} else {
							length = com.Int64Max(length, list.Len())
						}
					}
				}

				out := NewRslList()

				for idx := int64(0); idx < length; idx++ {
					listAtIdx := NewRslList()
					out.Append(newRslValueList(listAtIdx))
					for _, argList := range f.args {
						argList := argList.value.RequireList(f.i, argList.node)

						if idx < argList.Len() {
							listAtIdx.Append(argList.IndexAt(f.i, f.callNode, idx))
						} else {
							// logically: this should only happen if fill is provided
							listAtIdx.Append(*fill)
						}
					}
				}

				return newRslValues(f.i, f.callNode, out)
			},
		},
		{
			Name:             FUNC_STR,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				asStr := ToPrintableQuoteStr(arg.value, false)
				return newRslValues(f.i, f.callNode, asStr)
			},
		},
		{
			Name:             FUNC_SUM,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslListT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				list := arg.value.RequireList(f.i, arg.node)

				sum := 0.0
				for idx, item := range list.Values {
					num, ok := item.TryGetFloatAllowingInt()
					if !ok {
						f.i.errorf(arg.node, "%s() requires a list of numbers, got %q at index %d", FUNC_SUM, TypeAsString(item), idx)
					}
					sum += num
				}

				return newRslValues(f.i, f.callNode, sum)
			},
		},
		{
			Name:             FUNC_TRIM,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				textArg := f.args[0]

				chars := " \t\n"
				if len(f.args) > 1 {
					charsArg := f.args[1]
					chars = charsArg.value.RequireStr(f.i, charsArg.node).Plain()
				}

				rslString := textArg.value.RequireStr(f.i, textArg.node)
				rslString = rslString.Trim(chars)
				return newRslValues(f.i, f.callNode, rslString)
			},
		},
		{
			Name:             FUNC_TRIM_PREFIX,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				textArg := f.args[0]
				prefix := " \t\n"
				if len(f.args) > 1 {
					charsArg := f.args[1]
					prefix = charsArg.value.RequireStr(f.i, charsArg.node).Plain()
				}
				rslString := textArg.value.RequireStr(f.i, textArg.node)
				rslString = rslString.TrimPrefix(prefix)
				return newRslValues(f.i, f.callNode, rslString)
			},
		},
		{
			Name:             FUNC_TRIM_SUFFIX,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}, {RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				textArg := f.args[0]
				suffix := " \t\n"
				if len(f.args) > 1 {
					charsArg := f.args[1]
					suffix = charsArg.value.RequireStr(f.i, charsArg.node).Plain()
				}
				rslString := textArg.value.RequireStr(f.i, textArg.node)
				rslString = rslString.TrimSuffix(suffix)
				return newRslValues(f.i, f.callNode, rslString)
			},
		},
		{
			// todo potential additional named args
			//   - encoding="utf-8", # Or null for raw bytes
			//   - start             # Byte offset start
			//   - length            # Number of bytes to read
			//   - head              # First N bytes
			//   - tail              # Last N bytes
			Name:             FUNC_READ_FILE,
			ReturnValues:     UP_TO_TWO_RETURN_VALS,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs: map[string][]RslTypeEnum{
				namedArgMode: {RslStringT},
			},
			Execute: func(f FuncInvocationArgs) []RslValue {
				path := f.args[0].value.RequireStr(f.i, f.args[0].node).Plain()

				mode := constText
				if modeArg, exists := f.namedArgs[namedArgMode]; exists {
					mode = modeArg.value.RequireStr(f.i, modeArg.valueNode).Plain()
				}

				resultMap := NewRslMap()
				errMap := NewRslMap()

				data, err := os.ReadFile(path)
				if err == nil {
					resultMap.SetPrimitiveInt64(constSizeBytes, int64(len(data)))

					switch strings.ToLower(mode) {
					case constText:
						resultMap.SetPrimitiveStr(constContent, string(data))
					case constBytes:
						byteList := NewRslList()
						for _, b := range data {
							byteList.Append(newRslValueInt64(int64(b)))
						}
						resultMap.SetPrimitiveList(constContent, byteList)
					default:
						f.i.errorf(f.callNode, "Invalid mode %q in read_file; expected %q or %q", mode, constText, constBytes)
					}
				} else if os.IsNotExist(err) {
					errMap = ErrorRslMap(com.ErrFileNoExist, err.Error())
				} else if os.IsPermission(err) {
					errMap = ErrorRslMap(com.ErrFileNoPermission, err.Error())
				} else {
					errMap = ErrorRslMap(com.ErrFileRead, err.Error())
				}

				if f.numExpectedOutputs == 1 {
					if errMap.Len() > 0 {
						f.i.errorf(f.callNode, errMap.AsErrMsg(f.i, f.callNode))
					}
					return newRslValues(f.i, f.callNode, resultMap)
				}
				return newRslValues(f.i, f.callNode, resultMap, errMap)
			},
		},
		{
			Name:             FUNC_ROUND,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslFloatT, RslIntT}, {RslIntT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				var precision int64 = 0
				if len(f.args) > 1 {
					precisionArg := f.args[1]
					precision = precisionArg.value.RequireInt(f.i, precisionArg.node)
					if precision < 0 {
						f.i.errorf(f.args[1].node, "Precision must be non-negative, got %d", precision)
					}
				}

				val := arg.value.RequireFloatAllowingInt(f.i, arg.node)
				factor := math.Pow10(int(precision))
				rounded := math.Round(val*factor) / factor
				return newRslValues(f.i, f.callNode, rounded)
			},
		},
		{
			Name:             FUNC_CEIL,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslFloatT, RslIntT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				val := arg.value.RequireFloatAllowingInt(f.i, arg.node)
				return newRslValues(f.i, f.callNode, math.Ceil(val))
			},
		},
		{
			Name:             FUNC_FLOOR,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslFloatT, RslIntT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				val := arg.value.RequireFloatAllowingInt(f.i, arg.node)
				return newRslValues(f.i, f.callNode, math.Floor(val))
			},
		},
		{
			Name:             FUNC_MIN,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslListT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				list := arg.value.RequireList(f.i, arg.node)
				if list.Len() == 0 {
					f.i.errorf(f.callNode, "Cannot find minimum of empty list")
				}

				minVal := math.MaxFloat64
				for idx, item := range list.Values {
					val, ok := item.TryGetFloatAllowingInt()
					if !ok {
						f.i.errorf(arg.node, "%s() requires a list of numbers, got %q at index %d", FUNC_MIN, TypeAsString(item), idx)
					}
					minVal = math.Min(minVal, val)
				}
				return newRslValues(f.i, f.callNode, minVal)
			},
		},
		{
			Name:             FUNC_MAX,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslListT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				list := arg.value.RequireList(f.i, arg.node)
				if list.Len() == 0 {
					f.i.errorf(f.callNode, "Cannot find maximum of empty list")
				}

				maxVal := -math.MaxFloat64
				for idx, item := range list.Values {
					val, ok := item.TryGetFloatAllowingInt()
					if !ok {
						f.i.errorf(arg.node, "%s() requires a list of numbers, got %q at index %d", FUNC_MAX, TypeAsString(item), idx)
					}
					maxVal = math.Max(maxVal, val)
				}
				return newRslValues(f.i, f.callNode, maxVal)
			},
		},
		{
			Name:             FUNC_CLAMP,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 3,
			ArgTypes:         [][]RslTypeEnum{{RslFloatT, RslIntT}, {RslFloatT, RslIntT}, {RslFloatT, RslIntT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				// input is a number and a min and max
				val := f.args[0].value.RequireFloatAllowingInt(f.i, f.args[0].node)
				minVal := f.args[1].value.RequireFloatAllowingInt(f.i, f.args[1].node)
				maxVal := f.args[2].value.RequireFloatAllowingInt(f.i, f.args[2].node)

				if minVal > maxVal {
					f.i.errorf(f.callNode, "min must be less than max, got %f and %f", minVal, maxVal)
				}
				return newRslValues(f.i, f.callNode, math.Min(math.Max(val, minVal), maxVal))
			},
		},
		{
			Name:             FUNC_REVERSE,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				arg := f.args[0]
				rslString := arg.value.RequireStr(f.i, arg.node)
				return newRslValues(f.i, f.callNode, rslString.Reverse())
			},
		},
	}

	functions = append(functions, createColorFunctions()...)
	functions = append(functions, createHttpFunctions()...)

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
			Name:             color,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{}},
			NamedArgs:        NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) []RslValue {
				clr := ColorFromString(f.i, f.callNode, color)
				arg := f.args[0]
				switch coerced := arg.value.Val.(type) {
				case RslString:
					return newRslValues(f.i, arg.node, coerced.Color(clr))
				default:
					s := NewRslString(ToPrintable(arg.value))
					s.SetSegmentsColor(clr)
					return newRslValues(f.i, f.callNode, s)
				}
			},
		}
	}
	return funcs
}

func createHttpFunctions() []Func {
	httpFuncs := []string{
		FUNC_HTTP_GET,
		FUNC_HTTP_POST,
		FUNC_HTTP_PUT,
		FUNC_HTTP_PATCH,
		FUNC_HTTP_DELETE,
		FUNC_HTTP_HEAD,
		FUNC_HTTP_OPTIONS,
		FUNC_HTTP_TRACE,
		FUNC_HTTP_CONNECT,
	}

	funcs := make([]Func, len(httpFuncs))
	for idx, httpFunc := range httpFuncs {
		// todo handle exceptions?
		//   - auth?
		//   - query params help?
		//   - generic http for other/all methods?
		funcs[idx] = Func{
			Name:             httpFunc,
			ReturnValues:     ONE_RETURN_VAL,
			RequiredArgCount: 1,
			ArgTypes:         [][]RslTypeEnum{{RslStringT}},
			NamedArgs: map[string][]RslTypeEnum{
				namedArgHeaders: {RslMapT}, // string->string or string->list[string]
				namedArgBody:    {RslMapT, RslStringT},
			},
			Execute: func(f FuncInvocationArgs) []RslValue {
				urlArg := f.args[0]

				method := httpMethodFromFuncName(httpFunc)
				url := urlArg.value.RequireStr(f.i, urlArg.node).Plain()

				headers := make(map[string][]string)
				if headersArg, exists := f.namedArgs[namedArgHeaders]; exists {
					headerMap := headersArg.value.RequireMap(f.i, headersArg.valueNode)
					keys := headerMap.Keys()
					for _, key := range keys {
						value, _ := headerMap.Get(key)
						keyStr := key.RequireStr(f.i, headersArg.valueNode).Plain()
						switch coercedV := value.Val.(type) {
						case RslString:
							headers[keyStr] = []string{coercedV.Plain()}
						case *RslList:
							headers[keyStr] = coercedV.AsActualStringList(f.i, headersArg.valueNode)
						}
					}
				}

				var body *string
				if bodyArg, exists := f.namedArgs[namedArgBody]; exists {
					bodyStr := JsonToString(RslToJsonType(bodyArg.value))
					body = &bodyStr
				}

				reqDef := NewRequestDef(method, url, headers, body)
				response := RReq.Request(reqDef)
				rslMap := response.ToRslMap(f.i, f.callNode)
				return newRslValues(f.i, f.callNode, rslMap)
			},
		}
	}
	return funcs
}

func httpMethodFromFuncName(httpFunc string) string {
	switch httpFunc {
	case FUNC_HTTP_GET:
		return "GET"
	case FUNC_HTTP_POST:
		return "POST"
	case FUNC_HTTP_PUT:
		return "PUT"
	case FUNC_HTTP_PATCH:
		return "PATCH"
	case FUNC_HTTP_DELETE:
		return "DELETE"
	case FUNC_HTTP_HEAD:
		return "HEAD"
	case FUNC_HTTP_OPTIONS:
		return "OPTIONS"
	case FUNC_HTTP_TRACE:
		return "TRACE"
	case FUNC_HTTP_CONNECT:
		return "CONNECT"
	default:
		panic(fmt.Sprintf("Bug! Unknown HTTP function: %q", httpFunc))
	}
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
