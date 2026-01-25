package core

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts/rl"

	fid "github.com/amterp/flexid"

	"github.com/google/uuid"

	"github.com/amterp/rad/rts"
	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// Note: when adding functions, update the docs! docs-web/docs/reference/functions.md
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
	FUNC_MULTIPICK          = "multipick"
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
	FUNC_POW                = "pow"
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
	FUNC_READ_STDIN         = "read_stdin"
	FUNC_HAS_STDIN          = "has_stdin"
	FUNC_ROUND              = "round"
	FUNC_CEIL               = "ceil"
	FUNC_FLOOR              = "floor"
	FUNC_MIN                = "min"
	FUNC_MAX                = "max"
	FUNC_MATCHES            = "matches"
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
	FUNC_FLAT_MAP           = "flat_map"
	FUNC_LOAD               = "load"
	FUNC_COLOR_RGB          = "color_rgb"
	FUNC_COLORIZE           = "colorize"
	FUNC_GET_ARGS           = "get_args"

	INTERNAL_FUNC_GET_STASH_ID    = "_rad_get_stash_id"
	INTERNAL_FUNC_DELETE_STASH    = "_rad_delete_stash"
	INTERNAL_FUNC_RUN_CHECK       = "_rad_run_check"
	INTERNAL_FUNC_CHECK_FROM_LOGS = "_rad_check_from_logs"

	namedArgPreferExact    = "prefer_exact"
	namedArgReverse        = "reverse"
	namedArgTitle          = "title"
	namedArgPrompt         = "prompt"
	namedArgHeaders        = "headers"
	namedArgBody           = "body"
	namedArgJson           = "json"
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

	constContent        = "content" // todo rename to 'contents'? feels more natural
	constCreated        = "created"
	constSizeBytes      = "size_bytes"
	constBytesWritten   = "bytes_written"
	constText           = "text"
	constBytes          = "bytes"
	constCode           = "code"
	constMsg            = "msg"
	constTarget         = "target"
	constCwd            = "cwd"
	constAbsolute       = "absolute"
	constPath           = "path"
	constAlgo           = "algo"
	constSha1           = "sha1"
	constSha256         = "sha256"
	constSha512         = "sha512"
	constMd5            = "md5"
	constExists         = "exists"
	constFullPath       = "full_path"
	constBaseName       = "base_name"
	constPermissions    = "permissions"
	constType           = "type"
	constModifiedMillis = "modified_millis"
	constAccessedMillis = "accessed_millis"
	constDir            = "dir"
	constFile           = "file"
	constDefault        = "default"
	constAuto           = "auto"
	constSeconds        = "seconds"
	constMilliseconds   = "milliseconds"
	constMicroseconds   = "microseconds"
	constNanoseconds    = "nanoseconds"
)

var (
	NO_NAMED_ARGS_INPUT = map[string]namedArg{}
)

type FuncInvocation struct {
	i            *Interpreter
	callNode     *ts.Node
	args         []PosArg
	namedArgs    map[string]namedArg
	panicIfError bool
}

func NewFnInvocation(
	i *Interpreter,
	callNode *ts.Node,
	funcName string,
	args []PosArg,
	namedArgs map[string]namedArg,
	isBuiltIn bool,
) FuncInvocation {
	return FuncInvocation{
		i:            i,
		callNode:     callNode,
		args:         args,
		namedArgs:    namedArgs,
		panicIfError: !(funcName == FUNC_ERROR && isBuiltIn),
	}
}

func (f FuncInvocation) GetArg(name string) RadValue {
	val, ok := f.i.env.GetVar(name)
	if !ok {
		panic(fmt.Sprintf("Bug!! In built function requested undefined arg '%s', is your signature correct?",
			name))
	}
	return val
}

func (f FuncInvocation) GetBool(name string) bool {
	return f.GetArg(name).RequireBool(f.i, f.callNode)
}

func (f FuncInvocation) GetInt(name string) int64 {
	return f.GetArg(name).RequireInt(f.i, f.callNode)
}

func (f FuncInvocation) GetIntAllowingBool(name string) int64 {
	return f.GetArg(name).RequireIntAllowingBool(f.i, f.callNode)
}

func (f FuncInvocation) GetFloat(name string) float64 {
	return f.GetArg(name).RequireFloatAllowingInt(f.i, f.callNode)
}

func (f FuncInvocation) GetStr(name string) RadString {
	return f.GetArg(name).RequireStr(f.i, f.callNode)
}

func (f FuncInvocation) GetList(name string) *RadList {
	return f.GetArg(name).RequireList(f.i, f.callNode)
}

func (f FuncInvocation) GetMap(name string) *RadMap {
	return f.GetArg(name).RequireMap(f.i, f.callNode)
}

func (f FuncInvocation) GetFn(name string) RadFn {
	return f.GetArg(name).RequireFn(f.i, f.callNode)
}

func (f FuncInvocation) Return(vals ...interface{}) RadValue {
	return newRadValues(f.i, f.callNode, vals...)
}

func (f FuncInvocation) ReturnErrf(code rl.Error, msg string, args ...interface{}) RadValue {
	return f.Return(NewErrorStrf(msg, args...).SetCode(code).SetNode(f.callNode))
}

// todo add 'usage' to each function? self-documenting errors when incorrectly using
type BuiltInFunc struct {
	Name      string
	Signature *rts.FnSignature
	Execute   func(FuncInvocation) RadValue
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
		FuncMatches,
		FuncPick,
		FuncPickKv,
		FuncPickFromResource,
		FuncMultipick,
		FuncSplit,
		FuncRange,
		FuncColorize,
		{
			Name: FUNC_LEN,
			Execute: func(f FuncInvocation) RadValue {
				coll := f.GetArg("_val")
				switch v := coll.Val.(type) {
				case RadString:
					return f.Return(v.Len())
				case *RadList:
					return f.Return(v.Len())
				case *RadMap:
					return f.Return(v.Len())
				default:
					bugIncorrectTypes(FUNC_LEN)
					panic(UNREACHABLE)
				}
			},
		},
		{
			Name: FUNC_SORT,
			Execute: func(f FuncInvocation) RadValue {
				reverse := f.GetBool("reverse")
				primary := f.GetArg("_primary")
				others := f.GetList("_others")

				switch coerced := primary.Val.(type) {
				case RadString:
					if others.Len() > 0 {
						return f.ReturnErrf(rl.ErrGenericRuntime, "Can only parallel sort lists.")
					}
					runes := []rune(coerced.Plain())
					if reverse {
						sort.Slice(runes, func(i, j int) bool { return runes[i] > runes[j] })
					} else {
						sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
					}
					return f.Return(string(runes))

				case *RadList:
					n := coerced.Len()
					dir := Asc
					if reverse {
						dir = Desc
					}

					if others.Len() == 0 {
						sortedVals, _ := sortListParallel(f.i, f.callNode, coerced, dir)
						out := NewRadList()
						for _, v := range sortedVals {
							out.Append(v)
						}
						return f.Return(out)
					}

					// validate and collect the parallel lists
					otherLists := make([]*RadList, others.Len())
					for i, otherVal := range others.Values {
						lst, ok := otherVal.Val.(*RadList)
						if !ok || lst.Len() != n {
							return f.ReturnErrf(rl.ErrGenericRuntime,
								"Input lists were not the same length: %d vs %d", n, lst.Len())
						}
						otherLists[i] = lst
					}

					// sort primary and get the index permutation
					_, idxs := sortListParallel(f.i, f.callNode, coerced, dir)

					// rebuild the primary in sorted order
					sortedPrimary := NewRadList()
					for _, orig := range idxs {
						sortedPrimary.Append(coerced.Values[orig])
					}

					// rebuild each parallel list in lock-step
					sortedOthers := make([]RadValue, len(otherLists))
					for k, lst := range otherLists {
						newLst := NewRadList()
						for _, orig := range idxs {
							newLst.Append(lst.Values[orig])
						}
						sortedOthers[k] = newRadValueList(newLst)
					}

					// if there are parallels, return [primary, …others] as one RadList tuple
					tuple := NewRadList()
					tuple.Append(newRadValueList(sortedPrimary))
					for _, v := range sortedOthers {
						tuple.Append(v)
					}
					return f.Return(tuple)

				default:
					bugIncorrectTypes(FUNC_SORT)
					panic(UNREACHABLE)
				}
			},
		},
		{
			Name: FUNC_NOW,
			Execute: func(f FuncInvocation) RadValue {
				tz := f.GetStr("tz").Plain()
				var location *time.Location
				if tz == "local" {
					location = RClock.Local()
				} else {
					var err error
					location, err = time.LoadLocation(tz)
					if err != nil {
						errMsg := fmt.Sprintf("Invalid time zone '%s'", tz)
						return f.Return(NewErrorStrf(errMsg).SetCode(rl.ErrInvalidTimeZone))
					}
				}

				nowMap := NewTimeMap(RClock.Now().In(location))
				return f.Return(nowMap)
			},
		},
		{
			Name: FUNC_PARSE_EPOCH,
			Execute: func(f FuncInvocation) RadValue {
				epochArg := f.GetArg("_epoch")
				tz := f.GetStr("tz").Plain()
				unit := f.GetStr("unit").Plain()

				var isFloat bool
				var epochInt int64
				var fracPart float64
				var isNegative bool

				switch epochArg.Type() {
				case rl.RadIntT:
					epochInt = epochArg.RequireInt(f.i, f.callNode)
					isNegative = epochInt < 0
				case rl.RadFloatT:
					isFloat = true
					epochFloat := epochArg.RequireFloatAllowingInt(f.i, f.callNode)
					isNegative = epochFloat < 0
					epochInt = int64(epochFloat)
					fracPart = epochFloat - float64(epochInt)
				default:
					bugIncorrectTypes(FUNC_PARSE_EPOCH)
				}

				absEpoch := epochInt
				absFracPart := fracPart
				if isNegative {
					absEpoch = -absEpoch
					absFracPart = -absFracPart
				}

				digitCount := len(strconv.FormatInt(absEpoch, 10))
				var second int64
				var nanoSecond int64
				var fracMultiplier = 1e9 // default for seconds

				if unit == constAuto {
					switch digitCount {
					case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10:
						second = absEpoch
						nanoSecond = 0
						fracMultiplier = 1e9
					case 13:
						second = absEpoch / 1_000
						nanoSecond = (absEpoch % 1_000) * 1_000_000
						fracMultiplier = 1e6
					case 16:
						second = absEpoch / 1_000_000
						nanoSecond = (absEpoch % 1_000_000) * 1_000
						fracMultiplier = 1e3
					case 19:
						second = absEpoch / 1_000_000_000
						nanoSecond = absEpoch % 1_000_000_000
						fracMultiplier = 1
					default:
						errMsg := fmt.Sprintf(
							"Ambiguous epoch length (%d digits). Use '%s' to disambiguate.",
							digitCount,
							namedArgUnit,
						)
						return f.Return(NewErrorStrf(errMsg).SetCode(rl.ErrAmbiguousEpoch).SetNode(f.callNode))
					}
				} else {
					switch unit {
					case constSeconds:
						second = absEpoch
						nanoSecond = 0
						fracMultiplier = 1e9
					case constMilliseconds:
						second = absEpoch / 1_000
						nanoSecond = (absEpoch % 1_000) * 1_000_000
						fracMultiplier = 1e6
					case constMicroseconds:
						second = absEpoch / 1_000_000
						nanoSecond = (absEpoch % 1_000_000) * 1_000
						fracMultiplier = 1e3
					case constNanoseconds:
						second = absEpoch / 1_000_000_000
						nanoSecond = absEpoch % 1_000_000_000
						fracMultiplier = 1
					default:
						return f.ReturnErrf(rl.ErrInvalidTimeUnit,
							"invalid units %q; expected one of %s, %s, %s, %s, %s",
							unit, constAuto, constSeconds, constMilliseconds, constMicroseconds, constNanoseconds)
					}
				}

				if isFloat {
					nanoSecond += int64(math.Round(absFracPart * fracMultiplier))
				}

				if isNegative {
					second = -second
					nanoSecond = -nanoSecond
				}

				var location *time.Location
				if tz == "local" {
					location = RClock.Local()
				} else {
					var err error
					location, err = time.LoadLocation(tz)
					if err != nil {
						return f.ReturnErrf(rl.ErrInvalidTimeZone, "invalid time zone %q", tz)
					}
				}

				goTime := time.Unix(second, nanoSecond).In(location)
				timeMap := NewTimeMap(goTime)
				return f.Return(timeMap)
			},
		},
		{
			Name: FUNC_TYPE_OF,
			Execute: func(f FuncInvocation) RadValue {
				return f.Return(NewRadString(TypeAsString(f.GetArg("_var"))))
			},
		},
		{
			Name: FUNC_JOIN,
			Execute: func(f FuncInvocation) RadValue {
				list := f.GetList("_list")
				sep := f.GetStr("sep").Plain()
				prefix := f.GetStr("prefix").Plain()
				suffix := f.GetStr("suffix").Plain()
				return f.Return(list.Join(sep, prefix, suffix))
			},
		},
		{
			Name: FUNC_UPPER,
			Execute: func(f FuncInvocation) RadValue {
				return f.Return(f.GetStr("_val").Upper())
			},
		},
		{
			Name: FUNC_LOWER,
			Execute: func(f FuncInvocation) RadValue {
				return f.Return(f.GetStr("_val").Lower())
			},
		},
		{
			Name: FUNC_STARTS_WITH,
			Execute: func(f FuncInvocation) RadValue {
				val := f.GetStr("_val")
				start := f.GetStr("_start")
				return f.Return(strings.HasPrefix(val.Plain(), start.Plain()))
			},
		},
		{
			Name: FUNC_ENDS_WITH,
			Execute: func(f FuncInvocation) RadValue {
				val := f.GetStr("_val")
				end := f.GetStr("_end")
				return f.Return(strings.HasSuffix(val.Plain(), end.Plain()))
			},
		},
		{
			Name: FUNC_KEYS,
			Execute: func(f FuncInvocation) RadValue {
				return f.Return(f.GetMap("_map").Keys())
			},
		},
		{
			Name: FUNC_VALUES,
			Execute: func(f FuncInvocation) RadValue {
				return f.Return(f.GetMap("_map").Values())
			},
		},
		{
			Name: FUNC_TRUNCATE,
			Execute: func(f FuncInvocation) RadValue {
				str := f.GetStr("_str")
				maxLen := f.GetInt("_len")
				if maxLen < 0 {
					return f.ReturnErrf(rl.ErrNumInvalidRange, "Requires a non-negative int, got %d", maxLen)
				}

				strLen := str.Len()
				if maxLen >= strLen {
					return f.Return(str)
				}

				newStr := str.Plain() // todo should maintain attributes
				newStr = com.Truncate(newStr, maxLen)

				return f.Return(newStr)
			},
		},
		{
			Name: FUNC_UNIQUE,
			Execute: func(f FuncInvocation) RadValue {
				output := NewRadList()

				seen := make(map[string]struct{})
				list := f.GetList("_list")
				for _, item := range list.Values {
					key := ToPrintable(item) // todo not a solid approach
					if _, exists := seen[key]; !exists {
						seen[key] = struct{}{}
						output.Append(item)
					}
				}

				return f.Return(output)
			},
		},
		{
			Name: FUNC_CONFIRM,
			Execute: func(f FuncInvocation) RadValue {
				prompt := f.GetStr("prompt").Plain()

				response, err := InputConfirm("", prompt)
				if err != nil {
					// todo I think this errors if user aborts
					return f.ReturnErrf(rl.ErrUserInput, "Error reading input: %v", err)
				}

				return f.Return(response)
			},
		},
		{
			Name: FUNC_PARSE_JSON,
			Execute: func(f FuncInvocation) RadValue {
				out, err := TryConvertJsonToNativeTypes(f.i, f.callNode, f.GetStr("_str").Plain())
				if err != nil {
					return f.ReturnErrf(rl.ErrParseJson, "Error parsing JSON: %v", err)
				}
				return f.Return(out)
			},
		},
		{
			Name: FUNC_PARSE_INT,
			Execute: func(f FuncInvocation) RadValue {
				str := f.GetStr("_str").Plain()
				parsed, err := rts.ParseInt(str)

				if err == nil {
					return f.Return(parsed)
				} else {
					errMsg := fmt.Sprintf("%s() failed to parse %q", FUNC_PARSE_INT, str)
					return f.Return(NewErrorStrf(errMsg).SetCode(rl.ErrParseIntFailed))
				}
			},
		},
		{
			Name: FUNC_PARSE_FLOAT,
			Execute: func(f FuncInvocation) RadValue {
				str := f.GetStr("_str").Plain()
				parsed, err := rts.ParseFloat(str)

				if err == nil {
					return f.Return(parsed)
				} else {
					errMsg := fmt.Sprintf("%s() failed to parse %q", FUNC_PARSE_FLOAT, str)
					return f.Return(NewErrorStrf(errMsg).SetCode(rl.ErrParseFloatFailed))
				}
			},
		},
		{
			Name: FUNC_ABS,
			Execute: func(f FuncInvocation) RadValue {
				switch coerced := f.GetArg("_num").Val.(type) {
				case int64:
					return f.Return(AbsInt(coerced))
				case float64:
					return f.Return(AbsFloat(coerced))
				default:
					bugIncorrectTypes(FUNC_ABS)
					panic(UNREACHABLE)
				}
			},
		},
		{
			Name: FUNC_POW,
			Execute: func(f FuncInvocation) RadValue {
				base := f.GetFloat("_base")
				exponent := f.GetFloat("_exponent")
				result := math.Pow(base, exponent)
				return f.Return(result)
			},
		},
		{
			Name: FUNC_ERROR,
			Execute: func(f FuncInvocation) RadValue {
				err := f.GetStr("_msg")
				return f.Return(NewError(err))
			},
		},
		{
			Name: FUNC_INPUT,
			Execute: func(f FuncInvocation) RadValue {
				prompt := f.GetStr("prompt").Plain()
				hint := f.GetStr("hint").Plain()
				default_ := f.GetStr("default").Plain()
				secret := f.GetBool("secret")

				response, err := InputText(prompt, hint, default_, secret)
				if err != nil {
					return f.ReturnErrf(rl.ErrUserInput, "Error reading input: %v", err)
				}
				return f.Return(response)
			},
		},
		{
			Name: FUNC_GET_PATH,
			Execute: func(f FuncInvocation) RadValue {
				path := f.GetStr("_path").Plain()

				radMap := NewRadMap()
				radMap.SetPrimitiveBool(constExists, false)

				// todo os. calls should be abstracted away for testing
				absPath := com.ToAbsolutePath(path)
				RP.RadDebugf("Abs path: '%s'", absPath)

				radMap.SetPrimitiveStr(constFullPath, NormalizePath(absPath))

				stat, err1 := os.Stat(path)
				if err1 == nil {
					radMap.SetPrimitiveStr(constBaseName, stat.Name())
					radMap.SetPrimitiveStr(constPermissions, stat.Mode().Perm().String())
					fileType := lo.Ternary(stat.IsDir(), constDir, constFile)
					radMap.SetPrimitiveStr(constType, fileType)
					if fileType == constFile {
						radMap.SetPrimitiveInt64(constSizeBytes, stat.Size())
					}
					radMap.SetPrimitiveInt64(constModifiedMillis, stat.ModTime().UnixMilli())
					if atimeMillis, ok := getAccessTimeMillis(stat); ok {
						radMap.SetPrimitiveInt64(constAccessedMillis, atimeMillis)
					}
					radMap.SetPrimitiveBool(constExists, true)
				}

				return f.Return(radMap)
			},
		},
		{
			Name: FUNC_GET_ENV,
			Execute: func(f FuncInvocation) RadValue {
				envVar := f.GetStr("_var").Plain()
				envValue := os.Getenv(envVar)
				return f.Return(envValue)
			},
		},
		{
			Name: FUNC_FIND_PATHS,
			// todo: filtering by name, file type
			//  potentially allow `include_root`
			Execute: func(f FuncInvocation) RadValue {
				path := f.GetStr("_path").Plain()
				depth := f.GetInt("depth")

				relativeMode := f.GetStr("relative").Plain()

				switch relativeMode {
				case constTarget, constCwd, constAbsolute:
					// no-op, valid values
				default:
					return f.ReturnErrf(rl.ErrBugTypeCheck, "Invalid target mode %q. Allowed: %v",
						relativeMode, []string{constTarget, constCwd, constAbsolute})
				}

				absTarget, err := filepath.Abs(path) // todo should be abstracted away for testing
				if err != nil {
					return f.ReturnErrf(rl.ErrGenericRuntime, "Error resolving absolute path for target: %v", err)
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
						panic(fmt.Sprintf("Bug! Invalid target mode %q, should've been caught earlier.", relativeMode))
					}
					list.Append(newRadValueStr(NormalizePath(formattedPath)))
					return nil
				})

				if err != nil {
					return f.ReturnErrf(rl.ErrFileWalk, "Error walking directory: %v", err)
				}

				return f.Return(list)
			},
		},
		{
			// todo should offer args like find_paths
			Name: FUNC_DELETE_PATH,
			Execute: func(f FuncInvocation) RadValue {
				path := f.GetStr("_path").Plain()
				deleted := false

				if _, err := os.Stat(path); err == nil {
					// The path exists, so attempt to delete it.
					err = os.RemoveAll(path)
					deleted = err == nil
				}

				return f.Return(deleted)
			},
		},
		{
			// todo should support counting matches in a list as well
			Name: FUNC_COUNT,
			Execute: func(f FuncInvocation) RadValue {
				outer := f.GetStr("_str").Plain()
				inner := f.GetStr("_substr").Plain()

				count := strings.Count(outer, inner)
				return f.Return(count)
			},
		},
		{
			Name: FUNC_ZIP,
			Execute: func(f FuncInvocation) RadValue {
				strict := f.GetBool("strict")
				fill := f.GetArg("fill")

				if strict && !fill.IsNull() {
					return f.ReturnErrf(rl.ErrMutualExclArgs, "Cannot enable 'strict' with 'fill' specified")
				}

				lists := f.GetList("_lists")
				if lists.LenInt() == 0 {
					return f.Return(NewRadList())
				}

				length := int64(-1)
				for _, subList := range lists.Values {
					list := subList.RequireList(f.i, f.callNode)
					if length == -1 {
						length = list.Len()
					} else if length != list.Len() {
						if strict {
							return f.ReturnErrf(rl.ErrZipStrict,
								"Strict mode enabled: all lists must have the same length, but got %d and %d", length, list.Len())
						}
						if fill.IsNull() {
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
					for _, subListVal := range lists.Values {
						subList := subListVal.RequireList(f.i, f.callNode)

						if idx < subList.Len() {
							listAtIdx.Append(subList.IndexAt(f.i, f.callNode, idx))
						} else {
							// logically: this should only happen if fill is provided
							listAtIdx.Append(fill)
						}
					}
				}

				return f.Return(out)
			},
		},
		{
			Name: FUNC_STR,
			Execute: func(f FuncInvocation) RadValue {
				asStr := ToPrintableQuoteStr(f.GetArg("_var"), false)
				return f.Return(asStr)
			},
		},
		{
			Name: FUNC_INT,
			Execute: func(f FuncInvocation) RadValue {
				arg := f.GetArg("_var")

				switch coerced := arg.Val.(type) {
				case int64:
					return f.Return(coerced)
				case float64:
					return f.Return(int64(coerced))
				case bool:
					if coerced {
						return f.Return(int64(1))
					} else {
						return f.Return(int64(0))
					}
				case RadString:
					return f.ReturnErrf(
						rl.ErrCast,
						"Cannot cast string to int. Did you mean to use '%s' to parse the given string?",
						FUNC_PARSE_INT)
				default:
					return f.ReturnErrf(rl.ErrCast, "Cannot cast %q to int", arg.Type().AsString())
				}
			},
		},
		{
			Name: FUNC_FLOAT,
			Execute: func(f FuncInvocation) RadValue {
				arg := f.GetArg("_var")

				switch coerced := arg.Val.(type) {
				case int64:
					return f.Return(float64(coerced))
				case float64:
					return f.Return(coerced)
				case bool:
					if coerced {
						return f.Return(float64(1))
					} else {
						return f.Return(float64(0))
					}
				case RadString:
					return f.ReturnErrf(
						rl.ErrCast,
						"Cannot cast string to float. Did you mean to use '%s' to parse the given string?",
						FUNC_PARSE_FLOAT)
				default:
					return f.ReturnErrf(rl.ErrCast, "Cannot cast %q to float", arg.Type().AsString())
				}
			},
		},
		{
			Name: FUNC_SUM,
			Execute: func(f FuncInvocation) RadValue {
				list := f.GetList("_nums")

				sum := 0.0
				for idx, item := range list.Values {
					num, ok := item.TryGetFloatAllowingInt()
					if !ok {
						return f.ReturnErrf(
							rl.ErrBugTypeCheck,
							"Requires a list of numbers, got %q at index %d",
							TypeAsString(item),
							idx)
					}
					sum += num
				}

				return f.Return(sum)
			},
		},
		{
			Name: FUNC_TRIM,
			Execute: func(f FuncInvocation) RadValue {
				return runTrim(f, func(str RadString, chars string) RadString {
					return str.Trim(chars)
				})
			},
		},
		{
			Name: FUNC_TRIM_PREFIX,
			Execute: func(f FuncInvocation) RadValue {
				return runTrim(f, func(str RadString, chars string) RadString {
					return str.TrimPrefix(chars)
				})
			},
		},
		{
			Name: FUNC_TRIM_SUFFIX,
			Execute: func(f FuncInvocation) RadValue {
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
			Name: FUNC_READ_FILE,
			Execute: func(f FuncInvocation) RadValue {
				path := f.GetStr("_path").Plain()
				mode := f.GetStr("mode").Plain()

				data, err := os.ReadFile(path)
				if err == nil {
					resultMap := NewRadMap()
					resultMap.SetPrimitiveInt64(constSizeBytes, int64(len(data)))

					switch strings.ToLower(mode) {
					case constText:
						// Normalize line endings for consistent cross-platform text handling
						resultMap.SetPrimitiveStr(constContent, NormalizeLineEndings(string(data)))
					case constBytes:
						byteList := NewRadList()
						for _, b := range data {
							byteList.Append(newRadValueInt64(int64(b)))
						}
						resultMap.SetPrimitiveList(constContent, byteList)
					default:
						return f.ReturnErrf(
							rl.ErrBugTypeCheck,
							"Bug! Invalid mode %q in %s; expected %q or %q",
							mode,
							FUNC_READ_FILE,
							constText,
							constBytes)
					}
					return f.Return(resultMap)
				} else if os.IsNotExist(err) {
					return f.Return(NewErrorStrf(err.Error()).SetCode(rl.ErrFileNoExist))
				} else if os.IsPermission(err) {
					return f.Return(NewErrorStrf(err.Error()).SetCode(rl.ErrFileNoPermission))
				} else {
					return f.Return(NewErrorStrf(err.Error()).SetCode(rl.ErrFileRead))
				}
			},
		},
		{
			Name: FUNC_WRITE_FILE,
			Execute: func(f FuncInvocation) RadValue {
				path := f.GetStr("_path").Plain()
				content := f.GetStr("_content").Plain()
				appendFlag := f.GetBool("append")

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
					resultMap.SetPrimitiveStr(constPath, NormalizePath(path))
					return f.Return(resultMap)
				} else if os.IsNotExist(err) {
					return f.Return(NewErrorStrf(err.Error()).SetCode(rl.ErrFileNoExist))
				} else if os.IsPermission(err) {
					return f.Return(NewErrorStrf(err.Error()).SetCode(rl.ErrFileNoPermission))
				} else {
					return f.Return(NewErrorStrf(err.Error()).SetCode(rl.ErrFileWrite))
				}
			},
		},
		{
			Name: FUNC_READ_STDIN,
			Execute: func(f FuncInvocation) RadValue {
				// Check if stdin has content (is piped)
				if !RIo.StdIn.HasContent() {
					return f.Return(RAD_NULL_VAL)
				}

				// Read all stdin content
				data, err := io.ReadAll(RIo.StdIn)
				if err != nil {
					return f.Return(NewErrorStrf("Failed to read from stdin: %v", err).SetCode(rl.ErrStdinRead))
				}

				// Normalize line endings for consistent cross-platform text handling
				return f.Return(NormalizeLineEndings(string(data)))
			},
		},
		{
			Name: FUNC_HAS_STDIN,
			Execute: func(f FuncInvocation) RadValue {
				return f.Return(RIo.StdIn.HasContent())
			},
		},
		{
			Name: FUNC_ROUND,
			Execute: func(f FuncInvocation) RadValue {
				num := f.GetFloat("_num")
				precision := f.GetInt("_decimals")
				if precision < 0 {
					return f.ReturnErrf(rl.ErrNumInvalidRange, "Precision must be non-negative, got %d", precision)
				}

				if precision == 0 {
					return f.Return(int64(math.Round(num)))
				}

				factor := math.Pow10(int(precision))
				rounded := math.Round(num*factor) / factor
				return f.Return(rounded)
			},
		},
		{
			Name: FUNC_CEIL,
			Execute: func(f FuncInvocation) RadValue {
				num := f.GetFloat("_num")
				return f.Return(int64(math.Ceil(num)))
			},
		},
		{
			Name: FUNC_FLOOR,
			Execute: func(f FuncInvocation) RadValue {
				num := f.GetFloat("_num")
				return f.Return(int64(math.Floor(num)))
			},
		},
		{
			Name: FUNC_MIN,
			Execute: func(f FuncInvocation) RadValue {
				nums, errVal := extractMinMaxNums(f, FUNC_MIN)
				if errVal != nil {
					return *errVal
				}

				minVal := math.MaxFloat64
				for _, val := range nums {
					minVal = math.Min(minVal, val)
				}
				return f.Return(minVal)
			},
		},
		{
			Name: FUNC_MAX,
			Execute: func(f FuncInvocation) RadValue {
				nums, errVal := extractMinMaxNums(f, FUNC_MAX)
				if errVal != nil {
					return *errVal
				}

				maxVal := -math.MaxFloat64
				for _, val := range nums {
					maxVal = math.Max(maxVal, val)
				}

				return f.Return(maxVal)
			},
		},
		{
			Name: FUNC_CLAMP,
			Execute: func(f FuncInvocation) RadValue {
				valNum := f.GetFloat("val")
				minNum := f.GetFloat("min")
				maxNum := f.GetFloat("max")

				if minNum > maxNum {
					return f.ReturnErrf(rl.ErrArgsContradict, "min must be <= max, got %f and %f", minNum, maxNum)
				}
				return f.Return(math.Min(math.Max(valNum, minNum), maxNum))
			},
		},
		{
			Name: FUNC_REVERSE,
			Execute: func(f FuncInvocation) RadValue {
				val := f.GetStr("_val")
				return f.Return(val.Reverse())
			},
		},
		{
			Name: FUNC_IS_DEFINED,
			Execute: func(f FuncInvocation) RadValue {
				name := f.GetStr("_var").Plain()
				val, ok := f.i.env.GetVar(name)
				if !ok {
					return f.Return(false)
				}
				return f.Return(val.Type() != rl.RadNullT)
			},
		},
		{
			Name: FUNC_HYPERLINK,
			Execute: func(f FuncInvocation) RadValue {
				text := f.GetArg("_val")
				link := f.GetStr("_link")
				switch coerced := text.Val.(type) {
				case RadString:
					return f.Return(coerced.Hyperlink(link))
				default:
					s := NewRadString(ToPrintable(text))
					s.SetSegmentsHyperlink(link)
					return f.Return(s)
				}
			},
		},
		{
			Name: FUNC_UUID_V4,
			Execute: func(f FuncInvocation) RadValue {
				id, _ := uuid.NewRandom()
				return f.Return(id.String())
			},
		},
		{
			Name: FUNC_UUID_V7,
			Execute: func(f FuncInvocation) RadValue {
				id, _ := uuid.NewV7()
				return f.Return(id.String())
			},
		},
		{
			Name: FUNC_GEN_FID,
			Execute: func(f FuncInvocation) RadValue {
				// defaults
				config := fid.NewConfig().
					WithTickSize(fid.Millisecond).
					WithNumRandomChars(6).
					WithAlphabet(fid.Base62Alphabet)

				alphabet := f.GetArg("alphabet")
				if !alphabet.IsNull() {
					config = config.WithAlphabet(alphabet.RequireStr(f.i, f.callNode).Plain())
				}

				tickSizeMs := f.GetArg("tick_size_ms")
				if !tickSizeMs.IsNull() {
					tickSize := tickSizeMs.RequireInt(f.i, f.callNode)
					config = config.WithTickSize(time.Duration(tickSize) * time.Millisecond)
				}

				numRandomCharsArg := f.GetArg("num_random_chars")
				if !numRandomCharsArg.IsNull() {
					numRandomChars := numRandomCharsArg.RequireInt(f.i, f.callNode)
					if numRandomChars < 0 {
						return f.ReturnErrf(
							rl.ErrNumInvalidRange,
							"Number of random chars must be non-negative, got %d",
							numRandomChars)
					}
					config = config.WithNumRandomChars(int(numRandomChars))
				}

				generator, err := fid.NewGenerator(config)
				if err != nil {
					return f.ReturnErrf(rl.ErrFid, "Error creating FID generator: %v", err)
				}

				id, err := generator.Generate()
				if err != nil {
					return f.ReturnErrf(rl.ErrFid, "Error generating FID: %v", err)
				}

				return f.Return(id)
			},
		},
		{
			Name: FUNC_GET_DEFAULT,
			Execute: func(f FuncInvocation) RadValue {
				radMap := f.GetMap("_map")
				key := f.GetArg("key")
				def := f.GetArg("default")

				value, ok := radMap.Get(key)
				if !ok {
					value = def
				}

				return f.Return(value)
			},
		},
		{
			Name: FUNC_GET_RAD_HOME,
			Execute: func(f FuncInvocation) RadValue {
				radHome := RadHomeInst.HomeDir
				return f.Return(NormalizePath(radHome))
			},
		},
		{
			Name: FUNC_GET_STASH_DIR,
			Execute: func(f FuncInvocation) RadValue {
				stashPath := RadHomeInst.GetStash()
				if stashPath == nil {
					return f.Return(errNoStashId(f.callNode))
				}

				subPath := f.GetStr("_sub_path").Plain()
				path := filepath.Join(*stashPath, subPath)

				return f.Return(NormalizePath(path))
			},
		},
		{
			Name: FUNC_LOAD_STATE,
			Execute: func(f FuncInvocation) RadValue {
				state, _, err := RadHomeInst.LoadState(f.i, f.callNode)
				if err != nil {
					return f.Return(err)
				}
				state.RequireMap(f.i, f.callNode)
				return f.Return(state)
			},
		},
		{
			Name: FUNC_SAVE_STATE,
			Execute: func(f FuncInvocation) RadValue {
				err := RadHomeInst.SaveState(f.i, f.callNode, f.GetArg("_state"))
				if err != nil {
					return f.Return(err)
				}
				return f.Return()
			},
		},
		{
			Name: FUNC_LOAD_STASH_FILE,
			Execute: func(f FuncInvocation) RadValue {
				pathStr := f.GetStr("_path").Plain()
				def := f.GetStr("_default").Plain()

				path, err := RadHomeInst.GetStashSub(pathStr, f.callNode)
				if err != nil {
					return f.Return(err)
				}

				output := NewRadMap()
				output.SetPrimitiveStr(constFullPath, NormalizePath(path))

				if !com.FileExists(path) {
					err := com.CreateFilePathAndWriteString(path, def)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to create file %q: %v", path, err)
						return f.Return(NewErrorStrf(errMsg))
					}

					output.SetPrimitiveStr(constContent, def)
					output.SetPrimitiveBool(constCreated, true)
					return f.Return(output) // signal not existed
				}

				loadResult := com.LoadFile(path)
				if loadResult.Error != nil {
					errMsg := fmt.Sprintf("Error loading file %q: %v", path, loadResult.Error)
					return f.Return(NewErrorStrf(errMsg))
				}

				output.SetPrimitiveStr(constContent, NormalizeLineEndings(loadResult.Content))
				output.SetPrimitiveBool(constCreated, false)
				return f.Return(output) // signal existed
			},
		},
		{
			Name: FUNC_WRITE_STASH_FILE,
			Execute: func(f FuncInvocation) RadValue {
				pathStr := f.GetStr("_path").Plain()
				content := f.GetStr("_content").Plain()

				path, err1 := RadHomeInst.GetStashSub(pathStr, f.callNode)
				if err1 != nil {
					return f.Return(err1)
				}

				err := com.CreateFilePathAndWriteString(path, content)
				if err != nil {
					errMsg := fmt.Sprintf("Error writing stash file %q: %v", path, err)
					return f.Return(NewErrorStrf(errMsg).SetCode(rl.ErrFileWrite))
				}

				return f.Return()
			},
		},
		{
			Name: FUNC_HASH,
			Execute: func(f FuncInvocation) RadValue {
				content := f.GetStr("_val").Plain()
				algo := f.GetStr("algo").Plain()

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
					errMsg := fmt.Sprintf("Unsupported hash algorithm %q; supported: %s, %s, %s, %s",
						algo, constSha1, constSha256, constSha512, constMd5)
					return f.Return(NewErrorStrf(errMsg))
				}
				return f.Return(newRadValueStr(digest))
			},
		},
		{
			Name: FUNC_ENCODE_BASE64,
			Execute: func(f FuncInvocation) RadValue {
				content := f.GetStr("_content").Plain()
				urlSafe := f.GetBool("url_safe")
				padding := f.GetBool("padding")

				encoder := base64.StdEncoding
				if urlSafe {
					encoder = base64.URLEncoding
				}
				if !padding {
					encoder = encoder.WithPadding(base64.NoPadding)
				}

				encoded := encoder.EncodeToString([]byte(content))
				return f.Return(newRadValueStr(encoded))
			},
		},
		{
			Name: FUNC_DECODE_BASE64,
			Execute: func(f FuncInvocation) RadValue {
				content := f.GetStr("_content").Plain()
				urlSafe := f.GetBool("url_safe")
				padding := f.GetBool("padding")

				encoder := base64.StdEncoding
				if urlSafe {
					encoder = base64.URLEncoding
				}
				if !padding {
					encoder = encoder.WithPadding(base64.NoPadding)
				}

				decodedBytes, err := encoder.DecodeString(content)
				if err != nil {
					return f.ReturnErrf(rl.ErrDecode, "Error decoding base64: %v", err)
				}
				decoded := string(decodedBytes)
				return f.Return(newRadValueStr(decoded))
			},
		},
		{
			Name: FUNC_ENCODE_BASE16,
			Execute: func(f FuncInvocation) RadValue {
				content := f.GetStr("_content").Plain()
				encoded := hex.EncodeToString([]byte(content))
				return f.Return(newRadValueStr(encoded))
			},
		},
		{
			Name: FUNC_DECODE_BASE16,
			Execute: func(f FuncInvocation) RadValue {
				content := f.GetStr("_content").Plain()
				decodedBytes, err := hex.DecodeString(content)
				if err != nil {
					return f.ReturnErrf(rl.ErrDecode, "Error decoding base16: %v", err)
				}
				decoded := string(decodedBytes)
				return f.Return(newRadValueStr(decoded))
			},
		},
		{
			Name: FUNC_MAP,
			Execute: func(f FuncInvocation) RadValue {
				coll := f.GetArg("_coll")
				fn := f.GetFn("_fn")

				switch coerced := coll.Val.(type) {
				case *RadList:
					outputList := NewRadList()
					for _, val := range coerced.Values {
						invocation := NewFnInvocation(
							f.i,
							f.callNode,
							fn.Name(),
							NewPosArgs(NewPosArg(f.callNode, val)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						outputList.Append(out)
					}
					return f.Return(outputList)
				case *RadMap:
					outputList := NewRadList()
					coerced.Range(func(key, value RadValue) bool {
						invocation := NewFnInvocation(
							f.i,
							f.callNode,
							fn.Name(),
							NewPosArgs(NewPosArg(f.callNode, key), NewPosArg(f.callNode, value)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						outputList.Append(out)
						return true // signal to keep going
					})
					return f.Return(outputList)
				default:
					panic(fmt.Sprintf("Bug! Expected either list or map %s", coll.Type().AsString()))
				}
			},
		},
		{
			Name: FUNC_FILTER,
			Execute: func(f FuncInvocation) RadValue {
				coll := f.GetArg("_coll")
				fn := f.GetFn("_fn")

				switch coerced := coll.Val.(type) {
				case *RadList:
					outputList := NewRadList()
					for _, val := range coerced.Values {
						invocation := NewFnInvocation(
							f.i,
							f.callNode,
							fn.Name(),
							NewPosArgs(NewPosArg(f.callNode, val)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						if out.RequireBool(f.i, f.callNode) {
							outputList.Append(val)
						}
					}
					return f.Return(outputList)

				case *RadMap:
					outputMap := NewRadMap()
					coerced.Range(func(key, value RadValue) bool {
						invocation := NewFnInvocation(
							f.i,
							f.callNode,
							fn.Name(),
							NewPosArgs(NewPosArg(f.callNode, key), NewPosArg(f.callNode, value)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						if out.RequireBool(f.i, f.callNode) {
							outputMap.Set(key, value)
						}
						return true
					})
					return f.Return(outputMap)

				default:
					panic(fmt.Sprintf("Bug! Expected either list or map %s", coll.Type().AsString()))
				}
			},
		},
		{
			Name: FUNC_FLAT_MAP,
			Execute: func(f FuncInvocation) RadValue {
				coll := f.GetArg("_coll")
				fnArg := f.GetArg("_fn")

				outputList := NewRadList()

				switch coerced := coll.Val.(type) {
				case *RadList:
					if fnArg.IsNull() {
						for i, val := range coerced.Values {
							list, ok := val.TryGetList()
							if !ok {
								return f.ReturnErrf(rl.ErrGenericRuntime,
									"%s requires all elements to be lists, but element at index %d is %s",
									FUNC_FLAT_MAP, i, val.Type().AsString())
							}
							for _, elem := range list.Values {
								outputList.Append(elem)
							}
						}
					} else {
						fn := fnArg.RequireFn(f.i, f.callNode)
						for i, val := range coerced.Values {
							invocation := NewFnInvocation(
								f.i,
								f.callNode,
								fn.Name(),
								NewPosArgs(NewPosArg(f.callNode, val)),
								NO_NAMED_ARGS_INPUT,
								fn.IsBuiltIn(),
							)
							out := fn.Execute(invocation)
							list, ok := out.TryGetList()
							if !ok {
								return f.ReturnErrf(rl.ErrGenericRuntime,
									"%s function must return a list, but returned %s for element at index %d",
									FUNC_FLAT_MAP, out.Type().AsString(), i)
							}
							for _, elem := range list.Values {
								outputList.Append(elem)
							}
						}
					}
					return f.Return(outputList)

				case *RadMap:
					if fnArg.IsNull() {
						return f.ReturnErrf(rl.ErrGenericRuntime,
							"%s on maps requires a function argument", FUNC_FLAT_MAP)
					}
					fn := fnArg.RequireFn(f.i, f.callNode)
					var err RadValue
					var hasErr bool
					coerced.Range(func(key, value RadValue) bool {
						invocation := NewFnInvocation(
							f.i,
							f.callNode,
							fn.Name(),
							NewPosArgs(NewPosArg(f.callNode, key), NewPosArg(f.callNode, value)),
							NO_NAMED_ARGS_INPUT,
							fn.IsBuiltIn(),
						)
						out := fn.Execute(invocation)
						list, ok := out.TryGetList()
						if !ok {
							err = f.ReturnErrf(rl.ErrGenericRuntime,
								"%s function must return a list, but returned %s for key %v",
								FUNC_FLAT_MAP, out.Type().AsString(), key.Val)
							hasErr = true
							return false // stop iteration
						}
						for _, elem := range list.Values {
							outputList.Append(elem)
						}
						return true
					})
					if hasErr {
						return err
					}
					return f.Return(outputList)

				default:
					panic(fmt.Sprintf("Bug! Expected either list or map %s", coll.Type().AsString()))
				}
			},
		},
		{
			Name: FUNC_LOAD,
			Execute: func(f FuncInvocation) RadValue {
				coll := f.GetArg("_map")
				key := f.GetArg("_key")
				loaderFn := f.GetFn("_loader")

				// parse named args
				reload := f.GetBool("reload")
				overrideVal := f.GetArg(namedArgOverride)

				if !overrideVal.IsNull() && reload {
					return f.ReturnErrf(rl.ErrMutualExclArgs,
						"Cannot provide values for both %q and %q", namedArgReload, namedArgOverride)
				}

				switch m := coll.Val.(type) {
				case *RadMap:
					// override wins
					if !overrideVal.IsNull() {
						m.Set(key, overrideVal)
						return f.Return(overrideVal)
					}

					// helper to call loader
					runLoader := func() RadValue {
						inv := NewFnInvocation(
							f.i, f.callNode,
							loaderFn.Name(),
							NewPosArgs(),
							NO_NAMED_ARGS_INPUT,
							loaderFn.IsBuiltIn(),
						)
						return loaderFn.Execute(inv)
					}

					// forced reload
					if reload {
						v := runLoader()
						m.Set(key, v)
						return f.Return(v)
					}

					// existing?
					if existing, ok := m.Get(key); ok {
						return f.Return(existing)
					}

					// load new
					v := runLoader()
					m.Set(key, v)
					return f.Return(v)

				default:
					panic(fmt.Sprintf("Bug! Expected map but got %s", coll.Type().AsString()))
				}
			},
		},
		{
			Name: FUNC_COLOR_RGB,
			Execute: func(f FuncInvocation) RadValue {
				valArg := f.GetArg("_val")

				red := f.GetInt("red")
				green := f.GetInt("green")
				blue := f.GetInt("blue")

				for _, color := range []int64{red, green, blue} {
					if color < 0 || color > 255 {
						return f.ReturnErrf(rl.ErrNumInvalidRange, "RGB values must be [0, 255]; got %d", color)
					}
				}

				switch coerced := valArg.Val.(type) {
				case RadString:
					str := coerced.DeepCopy()
					str.SetRgb64(red, green, blue)
					return f.Return(str)
				default:
					str := NewRadString(ToPrintable(valArg))
					str.SetRgb64(red, green, blue)
					return f.Return(str)
				}
			},
		},
		{
			Name: FUNC_GET_ARGS,
			Execute: func(f FuncInvocation) RadValue {
				// When a rad script is invoked, os.Args will look like:
				// [ "rad", "./script.rl", "arg1", "arg2" ]
				// Users will not expect or want the initial "rad", so we cut that out.
				args := os.Args[1:]
				return f.Return(args)
			},
		},
	}

	functions = append(functions, createTextAttrFunctions()...)
	functions = append(functions, createHttpFunctions()...)

	FunctionsByName = make(map[string]BuiltInFunc)
	for _, f := range functions {
		sig := rts.GetSignature(f.Name)
		if sig == nil {
			panic("Bug! Function " + f.Name + " does not have a signature defined")
		}
		f.Signature = sig
		FunctionsByName[f.Name] = f
	}
}

func createTextAttrFunctions() []BuiltInFunc {
	attrStrs := lo.Values(attrEnumToStrings)
	funcs := make([]BuiltInFunc, len(attrStrs))
	for idx, attrStr := range attrStrs {
		funcs[idx] = BuiltInFunc{
			Name: attrStr,
			Execute: func(f FuncInvocation) RadValue {
				attr := AttrFromString(f.i, f.callNode, attrStr)
				switch coerced := f.GetArg("_item").Val.(type) {
				case RadString:
					return f.Return(coerced.CopyWithAttr(attr))
				default:
					s := NewRadString(ToPrintable(coerced))
					s.SetAttr(attr)
					return f.Return(s)
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
			Name: httpFunc,
			Execute: func(f FuncInvocation) RadValue {
				url := f.GetStr("url").Plain()
				method := httpMethodFromFuncName(httpFunc)

				headers := make(map[string][]string)

				headersArg := f.GetArg(namedArgHeaders)
				if !headersArg.IsNull() {
					headerMap := headersArg.RequireMap(f.i, f.callNode)
					keys := headerMap.Keys()
					for _, key := range keys {
						value, _ := headerMap.Get(key)
						keyStr := key.RequireStr(f.i, f.callNode).Plain()
						switch coercedV := value.Val.(type) {
						case RadString:
							headers[keyStr] = []string{coercedV.Plain()}
						case *RadList:
							headers[keyStr] = coercedV.AsActualStringList(f.i, f.callNode)
						}
					}
				}

				var body *string
				bodyArg := f.GetArg(namedArgBody)
				jsonArg := f.GetArg(namedArgJson)

				// Check for mutually exclusive parameters
				if !bodyArg.IsNull() && !jsonArg.IsNull() {
					return f.ReturnErrf(rl.ErrMutualExclArgs, "Cannot specify both 'body' and 'json' parameters")
				}

				if !bodyArg.IsNull() {
					// Use body as-is (raw string)
					bodyStr := ToPrintableQuoteStr(bodyArg.Val, false)
					body = &bodyStr
				} else if !jsonArg.IsNull() {
					// Convert to JSON and set default Content-Type header, if no headers provided
					bodyStr := JsonToString(RadToJsonType(jsonArg))
					body = &bodyStr

					if headersArg.IsNull() {
						headers["Content-Type"] = []string{"application/json"}
					}
				}

				reqDef := NewRequestDef(method, url, headers, body)
				response := RReq.Request(reqDef)
				radMap := response.ToRadMap(f.i, f.callNode)
				return f.Return(radMap)
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

func runTrim(f FuncInvocation, trimFunc func(str RadString, chars string) RadString) RadValue {
	subject := f.GetStr("_subject")
	toTrim := f.GetStr("_to_trim").Plain()

	subject = trimFunc(subject, toTrim)
	return f.Return(subject)
}

// extractMinMaxNums extracts float values from min/max variadic arguments.
// Supports two calling patterns:
// - Single list argument: min([1, 2, 3]) - iterates over the list's elements
// - Multiple number arguments: min(1, 2, 3) - uses each argument as a number
func extractMinMaxNums(f FuncInvocation, funcName string) ([]float64, *RadValue) {
	args := f.GetList("_nums")

	if args.Len() == 0 {
		errVal := f.ReturnErrf(rl.ErrEmptyList, "Cannot find %s of empty list", funcName)
		return nil, &errVal
	}

	// Check for single-list-argument mode: min([1, 2, 3])
	if args.LenInt() == 1 {
		if innerList, ok := args.Values[0].Val.(*RadList); ok {
			if innerList.Len() == 0 {
				errVal := f.ReturnErrf(rl.ErrEmptyList, "Cannot find %s of empty list", funcName)
				return nil, &errVal
			}
			nums := make([]float64, 0, innerList.LenInt())
			for idx, item := range innerList.Values {
				val, ok := item.TryGetFloatAllowingInt()
				if !ok {
					errVal := f.ReturnErrf(
						rl.ErrBugTypeCheck,
						"%s() requires a list of numbers, got %q at index %d",
						funcName, TypeAsString(item), idx)
					return nil, &errVal
				}
				nums = append(nums, val)
			}
			return nums, nil
		}
	}

	// Multiple arguments mode: min(1, 2, 3)
	// Each argument must be a number (not a list)
	nums := make([]float64, 0, args.LenInt())
	for idx, item := range args.Values {
		// Check if any argument is a list (error: should use single-list mode)
		if _, isList := item.Val.(*RadList); isList {
			errVal := f.ReturnErrf(
				rl.ErrBugTypeCheck,
				"%s() with multiple arguments requires numbers, not lists. Use %s([...]) for a single list",
				funcName, funcName)
			return nil, &errVal
		}
		val, ok := item.TryGetFloatAllowingInt()
		if !ok {
			errVal := f.ReturnErrf(
				rl.ErrBugTypeCheck,
				"%s() requires numbers, got %q at argument %d",
				funcName, TypeAsString(item), idx+1)
			return nil, &errVal
		}
		nums = append(nums, val)
	}
	return nums, nil
}

func bugIncorrectTypes(funcName string) {
	panic(fmt.Sprintf("Bug! Switch cases should line up with %q definition", funcName))
}
