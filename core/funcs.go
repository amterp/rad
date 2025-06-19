package core

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	com "rad/core/common"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/amterp/rad/rts/rl"

	fid "github.com/amterp/flexid"

	"github.com/google/uuid"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/raderr"

	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

const (
	FUNC_PRINT              = "print"
	FUNC_PRINT_ERR          = "print_err"
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
	FUNC_PARSE_EPOCH        = "parse_epoch"
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
	FUNC_ERROR              = "error"
	FUNC_GET_PATH           = "get_path"
	FUNC_FIND_PATHS         = "find_paths"
	FUNC_DELETE_PATH        = "delete_path"
	FUNC_COUNT              = "count"
	FUNC_ZIP                = "zip"
	FUNC_STR                = "str"
	FUNC_INT                = "int"
	FUNC_FLOAT              = "float"
	FUNC_SUM                = "sum"
	FUNC_TRIM               = "trim"
	FUNC_TRIM_PREFIX        = "trim_prefix"
	FUNC_TRIM_SUFFIX        = "trim_suffix"
	FUNC_READ_FILE          = "read_file"
	FUNC_WRITE_FILE         = "write_file"
	FUNC_ROUND              = "round"
	FUNC_CEIL               = "ceil"
	FUNC_FLOOR              = "floor"
	FUNC_MIN                = "min"
	FUNC_MAX                = "max"
	FUNC_CLAMP              = "clamp"
	FUNC_REVERSE            = "reverse"
	FUNC_IS_DEFINED         = "is_defined" // todo might be poorly named. should focus on vars. Or maybe just embrace works for anything, name it 'exists'?
	FUNC_HYPERLINK          = "hyperlink"
	FUNC_UUID_V4            = "uuid_v4"
	FUNC_UUID_V7            = "uuid_v7"
	FUNC_GEN_FID            = "gen_fid"
	FUNC_GET_DEFAULT        = "get_default"
	FUNC_GET_RAD_HOME       = "get_rad_home"
	FUNC_GET_STASH_DIR      = "get_stash_dir" // todo 'path' vs. 'dir' inconsistent naming
	FUNC_LOAD_STATE         = "load_state"
	FUNC_SAVE_STATE         = "save_state"
	FUNC_LOAD_STASH_FILE    = "load_stash_file"
	FUNC_WRITE_STASH_FILE   = "write_stash_file"
	FUNC_GET_ENV            = "get_env"
	FUNC_HASH               = "hash"
	FUNC_ENCODE_BASE64      = "encode_base64"
	FUNC_DECODE_BASE64      = "decode_base64"
	FUNC_ENCODE_BASE16      = "encode_base16"
	FUNC_DECODE_BASE16      = "decode_base16"
	FUNC_MAP                = "map"
	FUNC_FILTER             = "filter"
	FUNC_LOAD               = "load"
	FUNC_COLOR_RGB          = "color_rgb"
	FUNC_COLORIZE           = "colorize"
	FUNC_GET_ARGS           = "get_args"

	INTERNAL_FUNC_GET_STASH_ID = "_rad_get_stash_id"
	INTERNAL_FUNC_DELETE_STASH = "_rad_delete_stash"
	INTERNAL_FUNC_RUN_CHECK    = "_rad_run_check"

	namedArgReverse        = "reverse"
	namedArgTitle          = "title"
	namedArgPrompt         = "prompt"
	namedArgHeaders        = "headers"
	namedArgBody           = "body"
	namedArgHint           = "hint"
	namedArgDefault        = "default"
	namedArgSecret         = "secret"
	namedArgEnd            = "end"
	namedArgSep            = "sep"
	namedArgFill           = "fill"
	namedArgStrict         = "strict"
	namedArgMode           = "mode"
	namedArgDepth          = "depth"
	namedArgRelative       = "relative"
	namedArgAppend         = "append"
	namedArgTickSizeMs     = "tick_size_ms"
	namedArgNumRandomChars = "num_random_chars"
	namedArgAlphabet       = "alphabet"
	namedArgUrlSafe        = "url_safe"
	namedArgPadding        = "padding"
	namedArgReload         = "reload"
	namedArgOverride       = "override"
	namedArgUnit           = "unit"
	namedArgTz             = "tz"

	constContent      = "content" // todo rename to 'contents'? feels more natural
	constCreated      = "created"
	constSizeBytes    = "size_bytes"
	constBytesWritten = "bytes_written"
	constText         = "text"
	constBytes        = "bytes"
	constCode         = "code"
	constMsg          = "msg"
	constTarget       = "target"
	constCwd          = "cwd"
	constAbsolute     = "absolute"
	constPath         = "path"
	constAlgo         = "algo"
	constSha1         = "sha1"
	constSha256       = "sha256"
	constSha512       = "sha512"
	constMd5          = "md5"
	constExists       = "exists"
	constFullPath     = "full_path"
	constBaseName     = "base_name"
	constPermissions  = "permissions"
	constType         = "type"
	constDir          = "dir"
	constFile         = "file"
	constDefault      = "default"
	constAuto         = "auto"
	constSeconds      = "seconds"
	constMilliseconds = "milliseconds"
	constMicroseconds = "microseconds"
	constNanoseconds  = "nanoseconds"
)

var (
	NO_POS_ARGS         = NewEnumerableArgSchema([][]rl.RadType{})
	NO_NAMED_ARGS       = map[string][]rl.RadType{}
	NO_NAMED_ARGS_INPUT = map[string]namedArg{}
)

type FuncInvocationArgs struct {
	i            *Interpreter
	callNode     *ts.Node
	funcName     string
	args         []PosArg
	namedArgs    map[string]namedArg
	panicIfError bool
}

func NewFuncInvocationArgs(
	i *Interpreter,
	callNode *ts.Node,
	funcName string,
	args []PosArg,
	namedArgs map[string]namedArg,
	isBuiltIn bool,
) FuncInvocationArgs {
	return FuncInvocationArgs{
		i:            i,
		callNode:     callNode,
		funcName:     funcName,
		args:         args,
		namedArgs:    namedArgs,
		panicIfError: !(funcName == FUNC_ERROR && isBuiltIn),
	}
}

type PositionalArgSchema interface {
	validate(f FuncInvocationArgs, fn *BuiltInFunc)
}

type EnumerablePositionalArgSchema struct {
	argTypes [][]rl.RadType
}

func NewEnumerableArgSchema(argTypes [][]rl.RadType) PositionalArgSchema {
	return EnumerablePositionalArgSchema{argTypes: argTypes}
}

func (s EnumerablePositionalArgSchema) validate(f FuncInvocationArgs, builtInFunc *BuiltInFunc) {
	maxAcceptableArgs := len(s.argTypes)
	if len(f.args) > maxAcceptableArgs {
		f.i.errorf(f.callNode, "%s() requires at most %s, but got %d",
			builtInFunc.Name, com.Pluralize(maxAcceptableArgs, "argument"), len(f.args))
	}

	for idx, acceptableTypes := range s.argTypes {
		if len(acceptableTypes) == 0 {
			// there are no type constraints
			continue
		}

		if idx >= len(f.args) {
			// rest of the args are optional and not supplied
			break
		}

		arg := f.args[idx]
		if !lo.Contains(acceptableTypes, arg.value.Type()) {
			acceptable := english.OxfordWordSeries(
				lo.Map(acceptableTypes, func(t rl.RadType, _ int) string { return t.AsString() }), "or")
			f.i.errorf(arg.node, "Got %q as the %s argument of %s(), but must be: %s",
				arg.value.Type().AsString(), humanize.Ordinal(idx+1), builtInFunc.Name, acceptable)
		}
	}
}

type VarPositionalArgSchema struct {
	// acceptable types for any arg
	acceptableTypes []rl.RadType
}

func NewVarArgSchema(acceptableTypes []rl.RadType) PositionalArgSchema {
	return VarPositionalArgSchema{acceptableTypes: acceptableTypes}
}

func (s VarPositionalArgSchema) validate(f FuncInvocationArgs, builtInFunc *BuiltInFunc) {
	if len(s.acceptableTypes) == 0 {
		// there are no type constraints
		return
	}

	for idx, arg := range f.args {
		if !lo.Contains(s.acceptableTypes, arg.value.Type()) {
			acceptable := english.OxfordWordSeries(
				lo.Map(s.acceptableTypes, func(t rl.RadType, _ int) string { return t.AsString() }), "or")
			f.i.errorf(arg.node, "Got %q as the %s argument of %s(), but must be: %s",
				arg.value.Type().AsString(), humanize.Ordinal(idx+1), builtInFunc.Name, acceptable)
		}
	}
}

// todo add 'usage' to each function? self-documenting errors when incorrectly using
type BuiltInFunc struct {
	Name            string
	ReturnValues    []int
	MinPosArgCount  int
	PosArgValidator PositionalArgSchema     // by index, what types are allowed for that index. empty == any
	NamedArgs       map[string][]rl.RadType // name -> allowed types. empty == any
	// interpreter, callNode, positional args, named args
	// Guarantees when Execute invoked:
	// - given at least as many args as required (MinPosArgCount)
	// - not given more args than types have been defined for (PosArgTypes)
	// - only valid named args are given (if given) (valid name, valid type) (NamedArgs)
	Execute func(FuncInvocationArgs) RadValue
}

var FunctionsByName map[string]BuiltInFunc

func init() {
	functions := []BuiltInFunc{
		FuncPrint,
		FuncPPrint,
		FuncDebug,
		FuncPrintErr,
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
		FuncColorize,
		{
			Name:            FUNC_LEN,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT, rl.RadListT, rl.RadMapT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				switch v := arg.value.Val.(type) {
				case RadString:
					return newRadValues(f.i, arg.node, v.Len())
				case *RadList:
					return newRadValues(f.i, arg.node, v.Len())
				case *RadMap:
					return newRadValues(f.i, arg.node, v.Len())
				default:
					panic(bugIncorrectTypes(FUNC_LEN))
				}
			},
		},
		{
			Name:            FUNC_SORT,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT, rl.RadListT}}),
			NamedArgs: map[string][]rl.RadType{
				namedArgReverse: {rl.RadBoolT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				reverseArg, exists := f.namedArgs[namedArgReverse]
				reverse := false
				if exists {
					reverse = reverseArg.value.RequireBool(f.i, reverseArg.valueNode)
				}

				arg := f.args[0]
				switch coerced := arg.value.Val.(type) {
				case RadString:
					// todo maintain attributes
					str := f.i.eval(arg.node).Val.RequireStr(f.i, f.callNode).Plain()
					runes := []rune(str)
					sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
					return newRadValues(f.i, f.callNode, string(runes))
				case *RadList:
					sortedValues := sortList(f.i, arg.node, coerced, lo.Ternary(reverse, Desc, Asc))
					list := NewRadList()
					for _, v := range sortedValues {
						list.Append(v)
					}
					return newRadValues(f.i, arg.node, list)
				default:
					panic(bugIncorrectTypes(FUNC_SORT))
				}
			},
		},
		{
			Name:            FUNC_NOW,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NO_POS_ARGS,
			NamedArgs: map[string][]rl.RadType{
				namedArgTz: {rl.RadStrT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				tz := constDefault
				if tzArg, exists := f.namedArgs[namedArgTz]; exists {
					tz = tzArg.value.RequireStr(f.i, tzArg.valueNode).Plain()
				}
				var location *time.Location
				if tz == constDefault {
					location = time.Local
				} else {
					var err error
					location, err = time.LoadLocation(tz)
					if err != nil {
						errMsg := fmt.Sprintf("Invalid time zone '%s'", tz)
						return newRadValues(f.i, f.callNode, NewErrorStr(errMsg).SetCode(raderr.ErrInvalidTimeZone))
					}
				}

				nowMap := NewTimeMap(RClock.Now().In(location))
				return newRadValues(f.i, f.callNode, nowMap)
			},
		},
		{
			Name:            FUNC_PARSE_EPOCH,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadIntT}}),
			NamedArgs: map[string][]rl.RadType{
				namedArgTz:   {rl.RadStrT},
				namedArgUnit: {rl.RadStrT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				epochArg := f.args[0]
				epoch := epochArg.value.RequireInt(f.i, epochArg.node)

				tz := constDefault
				if tzArg, exists := f.namedArgs[namedArgTz]; exists {
					tz = tzArg.value.RequireStr(f.i, tzArg.valueNode).Plain()
				}

				unit := constAuto
				if unitArg, exists := f.namedArgs[namedArgUnit]; exists {
					unit = strings.ToLower(unitArg.value.RequireStr(f.i, unitArg.valueNode).Plain())
				}

				isNegative := epoch < 0
				absEpoch := epoch
				if isNegative {
					absEpoch = -absEpoch
				}

				digitCount := len(strconv.FormatInt(absEpoch, 10))
				var second int64
				var nanoSecond int64

				if unit == constAuto {
					switch digitCount {
					case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10:
						second = absEpoch
						nanoSecond = 0
					case 13:
						second = absEpoch / 1_000
						nanoSecond = (absEpoch % 1_000) * 1_000_000
					case 16:
						second = absEpoch / 1_000_000
						nanoSecond = (absEpoch % 1_000_000) * 1_000
					case 19:
						second = absEpoch / 1_000_000_000
						nanoSecond = absEpoch % 1_000_000_000
					default:
						errMsg := fmt.Sprintf(
							"Ambiguous epoch length (%d digits). Use '%s' to disambiguate.",
							digitCount,
							namedArgUnit,
						)
						return newRadValues(
							f.i,
							f.callNode,
							NewErrorStr(errMsg).SetCode(raderr.ErrAmbiguousEpoch).SetNode(epochArg.node),
						)
					}
				} else {
					switch unit {
					case constSeconds:
						second = absEpoch
						nanoSecond = 0
					case constMilliseconds:
						second = absEpoch / 1_000
						nanoSecond = (absEpoch % 1_000) * 1_000_000
					case constMicroseconds:
						second = absEpoch / 1_000_000
						nanoSecond = (absEpoch % 1_000_000) * 1_000
					case constNanoseconds:
						second = absEpoch / 1_000_000_000
						nanoSecond = absEpoch % 1_000_000_000
					default:
						errMsg := fmt.Sprintf("%s(): invalid units %q; expected one of %s, %s, %s, %s, %s",
							FUNC_PARSE_EPOCH, unit, constAuto, constSeconds, constMilliseconds, constMicroseconds, constNanoseconds)
						return newRadValues(f.i, f.callNode, NewErrorStr(errMsg).SetCode(raderr.ErrInvalidTimeUnit).SetNode(epochArg.node))
					}
				}

				if isNegative {
					second = -second
					nanoSecond = -nanoSecond
				}

				var location *time.Location
				if tz == constDefault {
					location = time.Local
				} else {
					var err error
					location, err = time.LoadLocation(tz)
					if err != nil {
						errMsg := fmt.Sprintf("%s(): invalid time zone %q", FUNC_PARSE_EPOCH, tz)
						return newRadValues(f.i, f.callNode, NewErrorStr(errMsg).SetCode(raderr.ErrInvalidTimeZone).SetNode(epochArg.node))
					}
				}

				goTime := time.Unix(second, nanoSecond).In(location)
				timeMap := NewTimeMap(goTime)
				return newRadValues(f.i, f.callNode, timeMap)
			},
		},
		{
			Name:            FUNC_TYPE_OF,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				return newRadValues(f.i, f.callNode, NewRadString(TypeAsString(f.args[0].value)))
			},
		},
		{
			Name:           FUNC_JOIN,
			ReturnValues:   ONE_RETURN_VAL,
			MinPosArgCount: 2, // todo: should "" just be the default joiner?
			PosArgValidator: NewEnumerableArgSchema(
				[][]rl.RadType{{rl.RadListT}, {rl.RadStrT}, {rl.RadStrT}, {rl.RadStrT}},
			),
			NamedArgs: NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
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

				return newRadValues(f.i, f.callNode, list.Join(sep, prefix, suffix))
			},
		},
		{
			Name:            FUNC_UPPER,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				return newRadValues(f.i, arg.node, arg.value.RequireStr(f.i, arg.node).Upper())
			},
		},
		{
			Name:            FUNC_LOWER,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				return newRadValues(f.i, arg.node, arg.value.RequireStr(f.i, arg.node).Lower())
			},
		},
		{
			Name:            FUNC_STARTS_WITH,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				subjectArg := f.args[0]
				prefixArg := f.args[1]
				subjectStr := subjectArg.value.RequireStr(f.i, subjectArg.node)
				prefixStr := prefixArg.value.RequireStr(f.i, prefixArg.node)
				return newRadValues(f.i, f.callNode, strings.HasPrefix(subjectStr.Plain(), prefixStr.Plain()))
			},
		},
		{
			Name:            FUNC_ENDS_WITH,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				subjectArg := f.args[0]
				prefixArg := f.args[1]
				subjectStr := subjectArg.value.RequireStr(f.i, subjectArg.node)
				prefixStr := prefixArg.value.RequireStr(f.i, prefixArg.node)
				return newRadValues(f.i, f.callNode, strings.HasSuffix(subjectStr.Plain(), prefixStr.Plain()))
			},
		},
		{
			Name:            FUNC_KEYS,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadMapT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				return newRadValues(f.i, arg.node, arg.value.RequireMap(f.i, arg.node).Keys())
			},
		},
		{
			Name:            FUNC_VALUES,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadMapT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				return newRadValues(f.i, arg.node, arg.value.RequireMap(f.i, arg.node).Values())
			},
		},
		{
			Name:            FUNC_TRUNCATE,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadIntT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				strArg := f.args[0]
				maxLenArg := f.args[1]
				maxLen := maxLenArg.value.RequireInt(f.i, maxLenArg.node)

				if maxLen < 0 {
					f.i.errorf(maxLenArg.node, "%s() takes a non-negative int, got %d", FUNC_TRUNCATE, maxLen)
				}

				radStr := strArg.value.RequireStr(f.i, strArg.node)
				strLen := radStr.Len()

				if maxLen >= strLen {
					return newRadValues(f.i, f.callNode, radStr)
				}

				str := radStr.Plain() // todo should maintain attributes
				str = com.Truncate(str, maxLen)

				return newRadValues(f.i, f.callNode, str)
			},
		},
		{
			Name:            FUNC_UNIQUE,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadListT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]

				output := NewRadList()

				seen := make(map[string]struct{})
				list := arg.value.RequireList(f.i, arg.node)
				for _, item := range list.Values {
					key := ToPrintable(item) // todo not a solid approach
					if _, exists := seen[key]; !exists {
						seen[key] = struct{}{}
						output.Append(item)
					}
				}

				return newRadValues(f.i, f.callNode, output)
			},
		},
		{
			Name:            FUNC_CONFIRM,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
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

				return newRadValues(f.i, f.callNode, response)
			},
		},
		{
			Name:            FUNC_PARSE_JSON,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]

				out, err := TryConvertJsonToNativeTypes(f.i, f.callNode, arg.value.RequireStr(f.i, arg.node).Plain())
				if err != nil {
					f.i.errorf(f.callNode, fmt.Sprintf("Error parsing JSON: %v", err))
				}
				return newRadValues(f.i, f.callNode, out)
			},
		},
		{
			Name:            FUNC_PARSE_INT,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]

				str := arg.value.RequireStr(f.i, arg.node).Plain()
				parsed, err := rts.ParseInt(str)

				if err == nil {
					return newRadValues(f.i, f.callNode, parsed)
				} else {
					errMsg := fmt.Sprintf("%s() failed to parse %q", FUNC_PARSE_INT, str)
					return newRadValues(f.i, f.callNode, NewErrorStr(errMsg).SetCode(raderr.ErrParseIntFailed))
				}
			},
		},
		{
			Name:            FUNC_PARSE_FLOAT,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]

				str := arg.value.RequireStr(f.i, arg.node).Plain()
				parsed, err := rts.ParseFloat(str)

				if err == nil {
					return newRadValues(f.i, f.callNode, parsed)
				} else {
					errMsg := fmt.Sprintf("%s() failed to parse %q", FUNC_PARSE_FLOAT, str)
					return newRadValues(f.i, f.callNode, NewErrorStr(errMsg).SetCode(raderr.ErrParseFloatFailed))
				}
			},
		},
		{
			Name:            FUNC_ABS,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadFloatT, rl.RadIntT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]

				switch coerced := arg.value.Val.(type) {
				case int64:
					return newRadValues(f.i, f.callNode, AbsInt(coerced))
				case float64:
					return newRadValues(f.i, f.callNode, AbsFloat(coerced))
				default:
					bugIncorrectTypes(FUNC_ABS)
					panic(UNREACHABLE)
				}
			},
		},
		{
			Name:            FUNC_ERROR,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				err := f.args[0].value.RequireStr(f.i, f.args[0].node)
				return newRadValues(f.i, f.callNode, NewError(err))
			},
		},
		{
			Name:            FUNC_INPUT,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs: map[string][]rl.RadType{
				namedArgHint:    {rl.RadStrT},
				namedArgDefault: {rl.RadStrT},
				namedArgSecret:  {rl.RadBoolT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
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

				secret := false
				if secretArg, exists := f.namedArgs[namedArgSecret]; exists {
					secret = secretArg.value.RequireBool(f.i, secretArg.valueNode)
				}

				response, err := InputText(prompt, hint, default_, secret)
				if err != nil {
					f.i.errorf(f.callNode, fmt.Sprintf("Error reading input: %v", err))
				}
				return newRadValues(f.i, f.callNode, response)
			},
		},
		{
			Name:            FUNC_GET_PATH,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				pathArg := f.args[0]
				path := pathArg.value.RequireStr(f.i, pathArg.node).Plain()

				radMap := NewRadMap()
				radMap.SetPrimitiveBool(constExists, false)

				// todo os. calls should be abstracted away for testing
				absPath := com.ToAbsolutePath(path)
				RP.RadDebugf("Abs path: '%s'", absPath)

				radMap.SetPrimitiveStr(constFullPath, absPath)

				stat, err1 := os.Stat(path)
				if err1 == nil {
					radMap.SetPrimitiveStr(constBaseName, stat.Name())
					radMap.SetPrimitiveStr(constPermissions, stat.Mode().Perm().String())
					fileType := lo.Ternary(stat.IsDir(), constDir, constFile)
					radMap.SetPrimitiveStr(constType, fileType)
					if fileType == constFile {
						radMap.SetPrimitiveInt64(constSizeBytes, stat.Size())
					}
					radMap.SetPrimitiveBool(constExists, true)
				}

				return newRadValues(f.i, f.callNode, radMap)
			},
		},
		{
			Name:            FUNC_GET_ENV,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				envVarArg := f.args[0]
				envVar := envVarArg.value.RequireStr(f.i, envVarArg.node).Plain()
				envValue := os.Getenv(envVar)
				return newRadValues(f.i, f.callNode, newRadValueStr(envValue))
			},
		},
		{
			Name:            FUNC_FIND_PATHS,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs: map[string][]rl.RadType{
				// todo: filtering by name, file type
				//  potentially allow `include_root`
				namedArgDepth:    {rl.RadIntT},
				namedArgRelative: {rl.RadStrT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
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

				list := NewRadList()
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
					list.Append(newRadValueStr(formattedPath))
					return nil
				})

				if err != nil {
					f.i.errorf(f.callNode, "Error walking directory: %v", err)
				}

				return newRadValues(f.i, f.callNode, list)
			},
		},
		{
			Name:            FUNC_DELETE_PATH,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				path := f.args[0].value.RequireStr(f.i, f.args[0].node).Plain()

				deleted := false

				if _, err := os.Stat(path); err == nil {
					// The path exists, so attempt to delete it.
					err = os.RemoveAll(path)
					deleted = err == nil
				}

				return newRadValues(f.i, f.callNode, deleted)
			},
		},
		{
			Name:            FUNC_COUNT,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				strArg := f.args[0]
				substrArg := f.args[1]

				str := strArg.value.RequireStr(f.i, strArg.node).Plain()
				substr := substrArg.value.RequireStr(f.i, substrArg.node).Plain()

				count := strings.Count(str, substr)
				return newRadValues(f.i, f.callNode, count)
			},
		},
		{
			Name:            FUNC_ZIP,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NewVarArgSchema([]rl.RadType{rl.RadListT}),
			NamedArgs: map[string][]rl.RadType{
				namedArgFill:   {},
				namedArgStrict: {rl.RadBoolT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				strictArg, strictExists := f.namedArgs[namedArgStrict]
				strict := false
				if strictExists {
					strict = strictArg.value.RequireBool(f.i, strictArg.valueNode)
				}

				fillArg, fillExists := f.namedArgs[namedArgFill]
				var fill *RadValue
				if fillExists {
					fill = &fillArg.value
				}

				if strictExists && fillExists {
					f.i.errorf(f.callNode, "Cannot specify both 'strict' and 'fill' named arguments")
				}

				if len(f.args) == 0 {
					return newRadValues(f.i, f.callNode, NewRadList())
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

				out := NewRadList()

				for idx := int64(0); idx < length; idx++ {
					listAtIdx := NewRadList()
					out.Append(newRadValueList(listAtIdx))
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

				return newRadValues(f.i, f.callNode, out)
			},
		},
		{
			Name:            FUNC_STR,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				asStr := ToPrintableQuoteStr(arg.value, false)
				return newRadValues(f.i, f.callNode, asStr)
			},
		},
		{
			Name:            FUNC_INT,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]

				output := int64(0)
				NewTypeVisitor(f.i, arg.node).
					ForInt(func(v RadValue, i int64) {
						output = i
					}).
					ForFloat(func(v RadValue, f float64) {
						output = int64(f)
					}).
					ForBool(func(v RadValue, b bool) {
						if b {
							output = 1
						} else {
							output = 0
						}
					}).
					ForString(func(v RadValue, str RadString) {
						f.i.errorf(
							arg.node,
							"Cannot cast string to int. Did you mean to use '%s' to parse the given string?",
							FUNC_PARSE_INT,
						)
					}).
					ForDefault(func(v RadValue) {
						f.i.errorf(arg.node, "Cannot cast %q to int", v.Type().AsString())
					}).Visit(arg.value)
				return newRadValues(f.i, f.callNode, output)
			},
		},
		{
			Name:            FUNC_FLOAT,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]

				output := 0.0
				NewTypeVisitor(f.i, arg.node).
					ForInt(func(v RadValue, i int64) {
						output = float64(i)
					}).
					ForFloat(func(v RadValue, f float64) {
						output = f
					}).
					ForBool(func(v RadValue, b bool) {
						if b {
							output = 1.0
						} else {
							output = 0.0
						}
					}).
					ForString(func(v RadValue, str RadString) {
						f.i.errorf(
							arg.node,
							"Cannot cast string to float. Did you mean to use '%s' to parse the given string?",
							FUNC_PARSE_FLOAT,
						)
					}).
					ForDefault(func(v RadValue) {
						f.i.errorf(arg.node, "Cannot cast %q to float", v.Type().AsString())
					}).
					Visit(arg.value)
				return newRadValues(f.i, f.callNode, output)
			},
		},
		{
			Name:            FUNC_SUM,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadListT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				list := arg.value.RequireList(f.i, arg.node)

				sum := 0.0
				for idx, item := range list.Values {
					num, ok := item.TryGetFloatAllowingInt()
					if !ok {
						f.i.errorf(
							arg.node,
							"%s() requires a list of numbers, got %q at index %d",
							FUNC_SUM,
							TypeAsString(item),
							idx,
						)
					}
					sum += num
				}

				return newRadValues(f.i, f.callNode, sum)
			},
		},
		{
			Name:            FUNC_TRIM,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				return runTrim(f, func(str RadString, chars string) RadString {
					return str.Trim(chars)
				})
			},
		},
		{
			Name:            FUNC_TRIM_PREFIX,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				return runTrim(f, func(str RadString, chars string) RadString {
					return str.TrimPrefix(chars)
				})
			},
		},
		{
			Name:            FUNC_TRIM_SUFFIX,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				return runTrim(f, func(str RadString, chars string) RadString {
					return str.TrimSuffix(chars)
				})
			},
		},
		{
			// todo potential additional named args
			//   - encoding="utf-8", # Or null for raw bytes
			//   - start             # Byte offset start
			//   - length            # Number of bytes to read
			//   - head              # First N bytes
			//   - tail              # Last N bytes
			Name:            FUNC_READ_FILE,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs: map[string][]rl.RadType{
				namedArgMode: {rl.RadStrT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				path := f.args[0].value.RequireStr(f.i, f.args[0].node).Plain()

				mode := constText
				if modeArg, exists := f.namedArgs[namedArgMode]; exists {
					mode = modeArg.value.RequireStr(f.i, modeArg.valueNode).Plain()
				}

				data, err := os.ReadFile(path)
				if err == nil {
					resultMap := NewRadMap()
					resultMap.SetPrimitiveInt64(constSizeBytes, int64(len(data)))

					switch strings.ToLower(mode) {
					case constText:
						resultMap.SetPrimitiveStr(constContent, string(data))
					case constBytes:
						byteList := NewRadList()
						for _, b := range data {
							byteList.Append(newRadValueInt64(int64(b)))
						}
						resultMap.SetPrimitiveList(constContent, byteList)
					default:
						f.i.errorf(
							f.callNode,
							"Invalid mode %q in read_file; expected %q or %q",
							mode,
							constText,
							constBytes,
						)
					}
					return newRadValues(f.i, f.callNode, resultMap)
				} else if os.IsNotExist(err) {
					return newRadValues(f.i, f.callNode, NewErrorStr(err.Error()).SetCode(raderr.ErrFileNoExist))
				} else if os.IsPermission(err) {
					return newRadValues(f.i, f.callNode, NewErrorStr(err.Error()).SetCode(raderr.ErrFileNoPermission))
				} else {
					return newRadValues(f.i, f.callNode, NewErrorStr(err.Error()).SetCode(raderr.ErrFileRead))
				}
			},
		},
		{
			Name:            FUNC_WRITE_FILE,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs: map[string][]rl.RadType{
				namedArgAppend: {rl.RadBoolT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				path := f.args[0].value.RequireStr(f.i, f.args[0].node).Plain()
				content := f.args[1].value.RequireStr(f.i, f.args[1].node).String()

				appendFlag := false
				if appendArg, exists := f.namedArgs[namedArgAppend]; exists {
					appendFlag = appendArg.value.RequireBool(f.i, appendArg.valueNode)
				}

				data := []byte(content)
				var err error
				var bytesWritten int

				if appendFlag {
					// Open the file in append mode (create if it doesn't exist).
					file, fileErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if fileErr != nil {
						err = fileErr
					} else {
						defer file.Close()
						bytesWritten, err = file.Write(data)
					}
				} else {
					// Overwrite the file (or create it if it doesn't exist).
					err = os.WriteFile(path, data, 0644)
					if err == nil {
						bytesWritten = len(data)
					}
				}

				if err == nil {
					resultMap := NewRadMap()
					resultMap.SetPrimitiveInt64(constBytesWritten, int64(bytesWritten))
					resultMap.SetPrimitiveStr(constPath, path)
					return newRadValues(f.i, f.callNode, resultMap)
				} else if os.IsNotExist(err) {
					return newRadValues(f.i, f.callNode, NewErrorStr(err.Error()).SetCode(raderr.ErrFileNoExist))
				} else if os.IsPermission(err) {
					return newRadValues(f.i, f.callNode, NewErrorStr(err.Error()).SetCode(raderr.ErrFileNoPermission))
				} else {
					return newRadValues(f.i, f.callNode, NewErrorStr(err.Error()).SetCode(raderr.ErrFileWrite))
				}
			},
		},
		{
			Name:            FUNC_ROUND,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadFloatT, rl.RadIntT}, {rl.RadIntT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				var precision int64 = 0
				if len(f.args) > 1 {
					precisionArg := f.args[1]
					precision = precisionArg.value.RequireInt(f.i, precisionArg.node)
					if precision < 0 {
						f.i.errorf(precisionArg.node, "Precision must be non-negative, got %d", precision)
					}
				}

				val := arg.value.RequireFloatAllowingInt(f.i, arg.node)
				if precision == 0 {
					return newRadValues(f.i, f.callNode, int64(math.Round(val)))
				}

				factor := math.Pow10(int(precision))
				rounded := math.Round(val*factor) / factor
				return newRadValues(f.i, f.callNode, rounded)
			},
		},
		{
			Name:            FUNC_CEIL,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadFloatT, rl.RadIntT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				val := arg.value.RequireFloatAllowingInt(f.i, arg.node)
				return newRadValues(f.i, f.callNode, math.Ceil(val))
			},
		},
		{
			Name:            FUNC_FLOOR,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadFloatT, rl.RadIntT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				val := arg.value.RequireFloatAllowingInt(f.i, arg.node)
				return newRadValues(f.i, f.callNode, math.Floor(val))
			},
		},
		{
			Name:            FUNC_MIN,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadListT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				list := arg.value.RequireList(f.i, arg.node)
				if list.Len() == 0 {
					f.i.errorf(f.callNode, "Cannot find minimum of empty list")
				}

				minVal := math.MaxFloat64
				for idx, item := range list.Values {
					val, ok := item.TryGetFloatAllowingInt()
					if !ok {
						f.i.errorf(
							arg.node,
							"%s() requires a list of numbers, got %q at index %d",
							FUNC_MIN,
							TypeAsString(item),
							idx,
						)
					}
					minVal = math.Min(minVal, val)
				}
				return newRadValues(f.i, f.callNode, minVal)
			},
		},
		{
			Name:            FUNC_MAX,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadListT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				list := arg.value.RequireList(f.i, arg.node)
				if list.Len() == 0 {
					f.i.errorf(f.callNode, "Cannot find maximum of empty list")
				}

				maxVal := -math.MaxFloat64
				for idx, item := range list.Values {
					val, ok := item.TryGetFloatAllowingInt()
					if !ok {
						f.i.errorf(
							arg.node,
							"%s() requires a list of numbers, got %q at index %d",
							FUNC_MAX,
							TypeAsString(item),
							idx,
						)
					}
					maxVal = math.Max(maxVal, val)
				}
				return newRadValues(f.i, f.callNode, maxVal)
			},
		},
		{
			Name:           FUNC_CLAMP,
			ReturnValues:   ONE_RETURN_VAL,
			MinPosArgCount: 3,
			PosArgValidator: NewEnumerableArgSchema(
				[][]rl.RadType{{rl.RadFloatT, rl.RadIntT}, {rl.RadFloatT, rl.RadIntT}, {rl.RadFloatT, rl.RadIntT}},
			),
			NamedArgs: NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				valArg := f.args[0]
				minArg := f.args[1]
				maxArg := f.args[2]

				val := valArg.value.RequireFloatAllowingInt(f.i, valArg.node)
				minVal := minArg.value.RequireFloatAllowingInt(f.i, minArg.node)
				maxVal := maxArg.value.RequireFloatAllowingInt(f.i, maxArg.node)

				if minVal > maxVal {
					f.i.errorf(f.callNode, "min must be <= max, got %f and %f", minVal, maxVal)
				}
				return newRadValues(f.i, f.callNode, math.Min(math.Max(val, minVal), maxVal))
			},
		},
		{
			Name:            FUNC_REVERSE,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				radString := arg.value.RequireStr(f.i, arg.node)
				return newRadValues(f.i, f.callNode, radString.Reverse())
			},
		},
		{
			Name:            FUNC_IS_DEFINED,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				arg := f.args[0]
				str := arg.value.RequireStr(f.i, arg.node).Plain()
				val, ok := f.i.env.GetVar(str)
				if !ok {
					return newRadValues(f.i, f.callNode, false)
				}
				return newRadValues(f.i, f.callNode, val.Type() != rl.RadNullT)
			},
		},
		{
			Name:            FUNC_HYPERLINK,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				text := f.args[0]
				linkArg := f.args[1]
				link := linkArg.value.RequireStr(f.i, linkArg.node)
				switch coerced := text.value.Val.(type) {
				case RadString:
					return newRadValues(f.i, text.node, coerced.Hyperlink(link))
				default:
					s := NewRadString(ToPrintable(text.value))
					s.SetSegmentsHyperlink(link)
					return newRadValues(f.i, f.callNode, s)
				}
			},
		},
		{
			Name:            FUNC_UUID_V4,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NO_POS_ARGS,
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				id, _ := uuid.NewRandom()
				return newRadValues(f.i, f.callNode, id.String())
			},
		},
		{
			Name:            FUNC_UUID_V7,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NO_POS_ARGS,
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				id, _ := uuid.NewV7()
				return newRadValues(f.i, f.callNode, id.String())
			},
		},
		{
			Name:            FUNC_GEN_FID,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NO_POS_ARGS,
			NamedArgs: map[string][]rl.RadType{
				namedArgAlphabet:       {rl.RadStrT},
				namedArgTickSizeMs:     {rl.RadIntT},
				namedArgNumRandomChars: {rl.RadIntT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				// defaults
				config := fid.NewConfig().
					WithTickSize(fid.Decisecond). // todo maybe milli, but reduce num random chars to 4?
					WithNumRandomChars(5).
					WithAlphabet(fid.Base62Alphabet)

				if alphabetArg, exists := f.namedArgs[namedArgAlphabet]; exists {
					alphabet := alphabetArg.value.RequireStr(f.i, alphabetArg.valueNode).Plain()
					config = config.WithAlphabet(alphabet)
				}

				if tickSizeArg, exists := f.namedArgs[namedArgTickSizeMs]; exists {
					tickSize := tickSizeArg.value.RequireInt(f.i, tickSizeArg.valueNode)
					config = config.WithTickSize(time.Duration(tickSize) * time.Millisecond)
				}

				if numRandomCharsArg, exists := f.namedArgs[namedArgNumRandomChars]; exists {
					numRandomChars := numRandomCharsArg.value.RequireInt(f.i, numRandomCharsArg.valueNode)
					if numRandomChars < 0 {
						f.i.errorf(
							numRandomCharsArg.valueNode,
							"Number of random chars must be non-negative, got %d",
							numRandomChars,
						)
					}
					config = config.WithNumRandomChars(int(numRandomChars))
				}

				generator, err := fid.NewGenerator(config)
				if err != nil {
					f.i.errorf(f.callNode, "Error creating FID generator: %v", err)
				}

				id, err := generator.Generate()
				if err != nil {
					f.i.errorf(f.callNode, "Error generating FID: %v", err)
				}

				return newRadValues(f.i, f.callNode, id)
			},
		},
		{
			Name:            FUNC_GET_DEFAULT,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  3,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadMapT}, {}, {}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				mapArg := f.args[0]
				keyArg := f.args[1]
				defaultArg := f.args[2]

				mapValue := mapArg.value.RequireMap(f.i, mapArg.node)
				value, ok := mapValue.Get(keyArg.value)
				if !ok {
					value = defaultArg.value
				}

				return newRadValues(f.i, f.callNode, value)
			},
		},
		{
			Name:            FUNC_GET_RAD_HOME,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NO_POS_ARGS,
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				radHome := RadHomeInst.HomeDir
				return newRadValues(f.i, f.callNode, radHome)
			},
		},
		{
			Name:            FUNC_GET_STASH_DIR,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				stashPath := RadHomeInst.GetStash()
				if stashPath == nil {
					errMissingScriptId(f.i, f.callNode)
				}

				subPathArg := tryGetArg(0, f.args)
				if subPathArg != nil {
					subPath := subPathArg.value.RequireStr(f.i, subPathArg.node).Plain()
					path := filepath.Join(*stashPath, subPath)
					stashPath = &path
				}

				return newRadValues(f.i, f.callNode, *stashPath)
			},
		},
		{
			Name:            FUNC_LOAD_STATE,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  0,
			PosArgValidator: NO_POS_ARGS,
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				state, _ := RadHomeInst.LoadState(f.i, f.callNode)
				state.RequireMap(f.i, f.callNode)
				return newRadValues(f.i, f.callNode, state)
			},
		},
		{
			Name:            FUNC_SAVE_STATE,
			ReturnValues:    ZERO_RETURN_VALS,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadMapT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				mapArg := f.args[0]
				mapArg.value.RequireMap(f.i, mapArg.node)

				RadHomeInst.SaveState(f.i, f.callNode, mapArg.value)

				return newRadValues(f.i, f.callNode)
			},
		},
		{
			Name:            FUNC_LOAD_STASH_FILE,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				pathArg := f.args[0]
				defaultArg := f.args[1]

				pathFromStash := pathArg.value.RequireStr(f.i, pathArg.node).Plain()
				path := RadHomeInst.GetStashSub(pathFromStash, f.i, f.callNode)

				output := NewRadMap()
				output.SetPrimitiveStr(constPath, path) // todo 'full_path' to be consistent with get_path?

				if !com.FileExists(path) {
					defaultStr := defaultArg.value.RequireStr(f.i, defaultArg.node).Plain()
					err := com.CreateFilePathAndWriteString(path, defaultStr)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to create file %q: %v", path, err)
						return newRadValues(f.i, f.callNode, NewErrorStr(errMsg))
					}

					output.Set(newRadValueStr(constContent), newRadValueStr(defaultStr))
					output.SetPrimitiveBool(constCreated, true)
					return newRadValues(f.i, f.callNode, output) // signal not existed
				}

				loadResult := com.LoadFile(path)
				if loadResult.Error != nil {
					errMsg := fmt.Sprintf("Error loading file %q: %v", path, loadResult.Error)
					return newRadValues(f.i, f.callNode, NewErrorStr(errMsg))
				}

				output.SetPrimitiveStr(constContent, loadResult.Content)
				output.SetPrimitiveBool(constCreated, false)
				return newRadValues(f.i, f.callNode, output) // signal existed
			},
		},
		{
			Name:            FUNC_WRITE_STASH_FILE,
			ReturnValues:    UP_TO_TWO_RETURN_VALS,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				pathArg := f.args[0]
				contentArg := f.args[1]

				pathFromStash := pathArg.value.RequireStr(f.i, pathArg.node).Plain()
				path := RadHomeInst.GetStashSub(pathFromStash, f.i, f.callNode)

				err := com.CreateFilePathAndWriteString(path, contentArg.value.RequireStr(f.i, contentArg.node).Plain())
				if err != nil {
					errMsg := fmt.Sprintf("Error writing stash file %q: %v", path, err)
					return newRadValues(f.i, f.callNode, NewErrorStr(errMsg).SetCode(raderr.ErrFileWrite))
				}

				return newRadValues(f.i, f.callNode, path) // todo seems weird to return full path?
			},
		},
		{
			Name:            FUNC_HASH,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs: map[string][]rl.RadType{
				constAlgo: {rl.RadStrT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				contentArg := f.args[0]
				content := contentArg.value.RequireStr(f.i, contentArg.node).Plain()

				algo := constSha1
				if algoArg, exists := f.namedArgs[constAlgo]; exists {
					algo = algoArg.value.RequireStr(f.i, algoArg.valueNode).Plain()
				}

				var digest string
				switch algo {
				case constSha1:
					sum := sha1.Sum([]byte(content))
					digest = hex.EncodeToString(sum[:])
				case constSha256:
					sum := sha256.Sum256([]byte(content))
					digest = hex.EncodeToString(sum[:])
				case constSha512:
					sum := sha512.Sum512([]byte(content))
					digest = hex.EncodeToString(sum[:])
				case constMd5:
					sum := md5.Sum([]byte(content))
					digest = hex.EncodeToString(sum[:])
				default:
					algoArg := f.namedArgs[constAlgo]
					errMsg := fmt.Sprintf("Unsupported hash algorithm %q; supported: %s, %s, %s, %s",
						algo, constSha1, constSha256, constSha512, constMd5)
					return newRadValues(f.i, algoArg.valueNode, NewErrorStr(errMsg))
				}
				return newRadValues(f.i, f.callNode, newRadValueStr(digest))
			},
		},
		{
			Name:           FUNC_ENCODE_BASE64,
			ReturnValues:   ONE_RETURN_VAL,
			MinPosArgCount: 1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{
				{rl.RadStrT},
			}),
			NamedArgs: map[string][]rl.RadType{
				namedArgUrlSafe: {rl.RadBoolT},
				namedArgPadding: {rl.RadBoolT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				contentArg := f.args[0]

				input := contentArg.value.RequireStr(f.i, contentArg.node).Plain()

				urlSafe := false
				if arg, exists := f.namedArgs[namedArgUrlSafe]; exists {
					urlSafe = arg.value.RequireBool(f.i, arg.valueNode)
				}

				padding := true
				if arg, exists := f.namedArgs[namedArgPadding]; exists {
					padding = arg.value.RequireBool(f.i, arg.valueNode)
				}

				encoder := base64.StdEncoding
				if urlSafe {
					encoder = base64.URLEncoding
				}
				if !padding {
					encoder = encoder.WithPadding(base64.NoPadding)
				}

				encoded := encoder.EncodeToString([]byte(input))
				return newRadValues(f.i, f.callNode, newRadValueStr(encoded))
			},
		},
		{
			Name:           FUNC_DECODE_BASE64,
			ReturnValues:   ONE_RETURN_VAL,
			MinPosArgCount: 1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{
				{rl.RadStrT},
			}),
			NamedArgs: map[string][]rl.RadType{
				namedArgUrlSafe: {rl.RadBoolT},
				namedArgPadding: {rl.RadBoolT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				input := f.args[0].value.RequireStr(f.i, f.args[0].node).Plain()

				urlSafe := false
				if arg, exists := f.namedArgs[namedArgUrlSafe]; exists {
					urlSafe = arg.value.RequireBool(f.i, arg.valueNode)
				}

				padding := true
				if arg, exists := f.namedArgs[namedArgPadding]; exists {
					padding = arg.value.RequireBool(f.i, arg.valueNode)
				}

				encoder := base64.StdEncoding
				if urlSafe {
					encoder = base64.URLEncoding
				}
				if !padding {
					encoder = encoder.WithPadding(base64.NoPadding)
				}

				decodedBytes, err := encoder.DecodeString(input)
				if err != nil {
					f.i.errorf(f.callNode, "Error decoding base64: %v", err)
				}
				decoded := string(decodedBytes)
				return newRadValues(f.i, f.callNode, newRadValueStr(decoded))
			},
		},
		{
			Name:            FUNC_ENCODE_BASE16,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				contentArg := f.args[0]
				input := contentArg.value.RequireStr(f.i, contentArg.node).Plain()
				encoded := hex.EncodeToString([]byte(input))
				return newRadValues(f.i, f.callNode, newRadValueStr(encoded))
			},
		},
		{
			Name:            FUNC_DECODE_BASE16,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				contentArg := f.args[0]
				input := contentArg.value.RequireStr(f.i, contentArg.node).Plain()
				decodedBytes, err := hex.DecodeString(input)
				if err != nil {
					f.i.errorf(f.callNode, "Error decoding base16: %v", err)
				}
				decoded := string(decodedBytes)
				return newRadValues(f.i, f.callNode, newRadValueStr(decoded))
			},
		},
		{
			Name:            FUNC_MAP,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadListT, rl.RadMapT}, {rl.RadFnT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				collectionArg := f.args[0]
				fnArg := f.args[1]

				fnNode := fnArg.node
				fn := fnArg.value.RequireFn(f.i, fnNode)
				fnName := rl.GetSrc(fnNode, f.i.sd.Src)

				var outputValue RadValue
				NewTypeVisitor(f.i, collectionArg.node).ForList(func(v RadValue, l *RadList) {
					outputList := NewRadList()
					for _, val := range l.Values {
						invocation := NewFuncInvocationArgs(
							f.i,
							f.callNode,
							fnName,
							NewPosArgs(NewPosArg(fnNode, val)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						outputList.Append(out)
					}
					outputValue = newRadValue(f.i, f.callNode, outputList)
				}).ForMap(func(v RadValue, m *RadMap) {
					outputList := NewRadList()
					m.Range(func(key, value RadValue) bool {
						invocation := NewFuncInvocationArgs(
							f.i,
							f.callNode,
							fnName,
							NewPosArgs(NewPosArg(fnNode, key), NewPosArg(fnNode, value)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						outputList.Append(out)
						return true // signal to keep going
					})
					outputValue = newRadValue(f.i, f.callNode, outputList)
				}).Visit(collectionArg.value)

				return newRadValues(f.i, f.callNode, outputValue)
			},
		},
		{
			Name:            FUNC_FILTER,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  2,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadListT, rl.RadMapT}, {rl.RadFnT}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				collectionArg := f.args[0]
				fnArg := f.args[1]

				fnNode := fnArg.node
				fn := fnArg.value.RequireFn(f.i, fnNode)
				fnName := rl.GetSrc(fnNode, f.i.sd.Src)

				var outputValue RadValue
				NewTypeVisitor(f.i, collectionArg.node).ForList(func(_ RadValue, l *RadList) {
					outputList := NewRadList()
					for _, val := range l.Values {
						invocation := NewFuncInvocationArgs(
							f.i,
							f.callNode,
							fnName,
							NewPosArgs(NewPosArg(fnNode, val)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						if out.RequireBool(f.i, fnNode) {
							// keep item
							outputList.Append(val)
						}
					}
					outputValue = newRadValue(f.i, f.callNode, outputList)
				}).ForMap(func(_ RadValue, m *RadMap) {
					outputMap := NewRadMap()
					m.Range(func(key, value RadValue) bool {
						invocation := NewFuncInvocationArgs(
							f.i,
							f.callNode,
							fnName,
							NewPosArgs(NewPosArg(fnNode, key), NewPosArg(fnNode, value)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						if out.RequireBool(f.i, fnNode) {
							// keep entry
							outputMap.Set(key, value)
						}
						return true // continue iteration
					})
					outputValue = newRadValue(f.i, f.callNode, outputMap)
				}).Visit(collectionArg.value)

				return newRadValues(f.i, f.callNode, outputValue)
			},
		},
		{
			Name:           FUNC_LOAD,
			ReturnValues:   ONE_RETURN_VAL,
			MinPosArgCount: 3,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{
				{rl.RadMapT}, // the map
				{},           // the key
				{rl.RadFnT},  // zero‐arg loader function
			}),
			NamedArgs: map[string][]rl.RadType{
				namedArgReload:   {rl.RadBoolT},
				namedArgOverride: {}, // any type
			},
			Execute: func(f FuncInvocationArgs) RadValue {
				mapArg := f.args[0]
				keyArg := f.args[1]
				loaderFnArg := f.args[2]

				m := mapArg.value.RequireMap(f.i, mapArg.node)
				key := keyArg.value.RequireStr(f.i, keyArg.node).Plain()

				reload := false
				if a, ok := f.namedArgs[namedArgReload]; ok {
					reload = a.value.RequireBool(f.i, a.valueNode)
				}
				overrideProvided := false
				var overrideVal RadValue
				if override, ok := f.namedArgs[namedArgOverride]; ok {
					overrideProvided = override.value.TruthyFalsy() // todo we need a null type, this is not correct
					overrideVal = override.value
				}

				if overrideProvided && reload {
					f.i.errorf(f.callNode,
						"Cannot provide values for both %q and %q", namedArgReload, namedArgOverride)
				}

				// prioritize override
				if overrideProvided {
					m.Set(newRadValueStr(key), overrideVal)
					return newRadValues(f.i, f.callNode, overrideVal)
				}

				// helper to invoke the loader fn
				runLoader := func() RadValue {
					fnNode := loaderFnArg.node
					fn := loaderFnArg.value.RequireFn(f.i, fnNode)
					fnName := rl.GetSrc(fnNode, f.i.sd.Src)
					inv := NewFuncInvocationArgs(f.i, fnNode, fnName, NewPosArgs(), NO_NAMED_ARGS_INPUT, fn.IsBuiltIn())
					out := fn.Execute(inv)
					return out
				}

				// if reload, ignore existing value if present and just load
				if reload {
					v := runLoader()
					m.Set(newRadValueStr(key), v)
					return newRadValues(f.i, f.callNode, v)
				}

				if existing, ok := m.Get(newRadValueStr(key)); ok {
					return newRadValues(f.i, f.callNode, existing)
				}

				// doesn't exist, so load
				v := runLoader()
				m.Set(newRadValueStr(key), v)
				return newRadValues(f.i, f.callNode, v)
			},
		},
		{
			Name:           FUNC_COLOR_RGB,
			ReturnValues:   ONE_RETURN_VAL,
			MinPosArgCount: 4,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{
				{},
				{rl.RadIntT},
				{rl.RadIntT},
				{rl.RadIntT},
			}),
			NamedArgs: NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				textArg := f.args[0]
				redArg := f.args[1]
				greenArg := f.args[2]
				blueArg := f.args[3]

				extractRgb := func(arg PosArg) int64 {
					node := arg.node
					val := arg.value.RequireInt(f.i, node)
					if val < 0 || val > 255 {
						f.i.errorf(node, "RGB values must be [0, 255]; got %d", val)
					}
					return val
				}
				red := extractRgb(redArg)
				green := extractRgb(greenArg)
				blue := extractRgb(blueArg)

				switch coerced := textArg.value.Val.(type) {
				case RadString:
					str := coerced.DeepCopy()
					str.SetRgb64(red, green, blue)
					return newRadValues(f.i, textArg.node, str)
				default:
					s := NewRadString(ToPrintable(textArg.value))
					s.SetRgb64(red, green, blue)
					return newRadValues(f.i, f.callNode, s)
				}
			},
		},
		{
			Name:            FUNC_GET_ARGS,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  0,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				// When a rad script is invoked, os.Args will look like:
				// [ "rad", "./script.rl", "arg1", "arg2" ]
				// Users will not expect or want the initial "rad", so we cut that out.
				args := os.Args[1:]
				return newRadValues(f.i, f.callNode, args)
			},
		},
	}

	functions = append(functions, createTextAttrFunctions()...)
	functions = append(functions, createHttpFunctions()...)

	FunctionsByName = make(map[string]BuiltInFunc)
	for _, f := range functions {
		validateFunction(f, FunctionsByName)
		FunctionsByName[f.Name] = f
	}
}

func validateFunction(f BuiltInFunc, functionsSoFar map[string]BuiltInFunc) {
	validator := f.PosArgValidator
	switch coerced := validator.(type) {
	case *EnumerablePositionalArgSchema:
		if f.MinPosArgCount > len(coerced.argTypes) {
			panic(fmt.Sprintf("Bug! Function %q has more required args than arg types", f.Name))
		}
	}

	if _, exists := functionsSoFar[f.Name]; exists {
		panic(fmt.Sprintf("Bug! Function %q already exists", f.Name))
	}
}

func createTextAttrFunctions() []BuiltInFunc {
	attrStrs := lo.Values(attrEnumToStrings)
	funcs := make([]BuiltInFunc, len(attrStrs))
	for idx, attrStr := range attrStrs {
		funcs[idx] = BuiltInFunc{
			Name:            attrStr,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{}}),
			NamedArgs:       NO_NAMED_ARGS,
			Execute: func(f FuncInvocationArgs) RadValue {
				attr := AttrFromString(f.i, f.callNode, attrStr)
				arg := f.args[0]
				switch coerced := arg.value.Val.(type) {
				case RadString:
					return newRadValues(f.i, arg.node, coerced.CopyWithAttr(attr))
				default:
					s := NewRadString(ToPrintable(arg.value))
					s.SetAttr(attr)
					return newRadValues(f.i, f.callNode, s)
				}
			},
		}
	}
	return funcs
}

func createHttpFunctions() []BuiltInFunc {
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

	funcs := make([]BuiltInFunc, len(httpFuncs))
	for idx, httpFunc := range httpFuncs {
		// todo handle exceptions?
		//   - auth?
		//   - query params help?
		//   - generic http for other/all methods?
		funcs[idx] = BuiltInFunc{
			Name:            httpFunc,
			ReturnValues:    ONE_RETURN_VAL,
			MinPosArgCount:  1,
			PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}}),
			NamedArgs: map[string][]rl.RadType{
				namedArgHeaders: {rl.RadMapT}, // string->string or string->list[string]
				namedArgBody:    {rl.RadStrT, rl.RadMapT, rl.RadListT},
			},
			Execute: func(f FuncInvocationArgs) RadValue {
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
						case RadString:
							headers[keyStr] = []string{coercedV.Plain()}
						case *RadList:
							headers[keyStr] = coercedV.AsActualStringList(f.i, headersArg.valueNode)
						}
					}
				}

				var body *string
				if bodyArg, exists := f.namedArgs[namedArgBody]; exists {
					bodyStr := JsonToString(RadToJsonType(bodyArg.value))
					body = &bodyStr
				}

				reqDef := NewRequestDef(method, url, headers, body)
				response := RReq.Request(reqDef)
				radMap := response.ToRadMap(f.i, f.callNode)
				return newRadValues(f.i, f.callNode, radMap)
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

func runTrim(f FuncInvocationArgs, trimFunc func(str RadString, chars string) RadString) RadValue {
	textArg := f.args[0]

	chars := " \t\n"
	if len(f.args) > 1 {
		charsArg := f.args[1]
		chars = charsArg.value.RequireStr(f.i, charsArg.node).Plain()
	}

	radString := textArg.value.RequireStr(f.i, textArg.node)
	radString = trimFunc(radString, chars)
	return newRadValues(f.i, f.callNode, radString)
}

func tryGetArg(idx int, args []PosArg) *PosArg {
	if idx >= len(args) {
		return nil
	}
	return &args[idx]
}

func bugIncorrectTypes(funcName string) string {
	return fmt.Sprintf("Bug! Switch cases should line up with %q definition", funcName)
}

func errMissingScriptId(i *Interpreter, node *ts.Node) {
	i.errorf(node, "Script ID is not set. Set the '%s' macro in the file header.", MACRO_STASH_ID)
}
