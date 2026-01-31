package rts

import (
	"fmt"

	"github.com/amterp/rad/rts/rl"
)

type FnSignature struct {
	Name      string
	Signature string
	Typing    *rl.TypingFnT
}

func newFnSignature(signature string) FnSignature {
	return FnSignature{
		Signature: signature,
	}
}

var FnSignaturesByName map[string]FnSignature

func GetSignature(name string) *FnSignature {
	if sig, ok := FnSignaturesByName[name]; ok {
		return &sig
	}
	return nil
}

func init() {
	signatures := []FnSignature{
		newFnSignature(`print(*_items: any, *, sep: str = " ", end: str = "\n") -> void`),
		newFnSignature(`print_err(*_items: any, *, sep: str = " ", end: str = "\n") -> void`),
		newFnSignature(`pprint(_item: any?) -> void`),
		newFnSignature(`debug(*_items: any, *, sep: str = " ", end: str = "\n") -> void`),
		newFnSignature(`exit(_code: int|bool = 0) -> void`),
		newFnSignature(`sleep(_duration: int|float|str, *, title: str?) -> void`),
		newFnSignature(`seed_random(_seed: int) -> void`),
		newFnSignature(`rand() -> float`),
		newFnSignature(`rand_int(_arg1: int = 9223372036854775807, _arg2: int?) -> int`),
		newFnSignature(`replace(_original: str, _find: str, _replace: str) -> str`),
		newFnSignature(`len(_val: str|list|map) -> int`),
		newFnSignature(`sort(_primary: list|str, *_others: list|str, *, reverse: bool = false) -> list|str`),
		newFnSignature(`now(*, tz: str = "local") -> error|{ "date": str, "year": int, "month": int, "day": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }`),
		newFnSignature(`parse_epoch(_epoch: int|float, *, tz: str = "local", unit: ["auto", "seconds", "milliseconds", "microseconds", "nanoseconds"] = "auto") -> error|{ "date": str, "year": int, "month": int, "day": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }`),
		newFnSignature(`type_of(_var: any) -> ["int", "str", "list", "map", "float"]`),
		newFnSignature(`range(_arg1: float|int, _arg2: float?|int?, _step: float|int = 1) -> float[]|int[]`),
		newFnSignature(`join(_list: list, sep: str = "", prefix: str = "", suffix: str = "") -> str`),
		newFnSignature(`split(_val: str, _sep: str) -> str[]`),
		newFnSignature(`lower(_val: str) -> str`),
		newFnSignature(`upper(_val: str) -> str`),
		newFnSignature(`starts_with(_val: str, _start: str) -> bool`),
		newFnSignature(`ends_with(_val: str, _end: str) -> bool`),
		newFnSignature(`pick(_options: str[], _filter: str?|str[]?, *, prompt: str = "Pick an option", prefer_exact: bool = false) -> str`),
		newFnSignature(`pick_kv(keys: str[], values: any[], _filter: str?|str[]?, *, prompt: str = "Pick an option", prefer_exact: bool = false) -> any`),
		newFnSignature(`pick_from_resource(path: str, _filter: str?, *, prompt: str = "Pick an option", prefer_exact: bool = true) -> any`),
		newFnSignature(`multipick(_options: str[], *, prompt: str?, min: int = 0, max: int?) -> str[]`),
		newFnSignature(`keys(_map: map) -> any[]`),
		newFnSignature(`values(_map: map) -> any[]`),
		newFnSignature(`truncate(_str: str, _len: int) -> error|str`),
		newFnSignature(`unique(_list: any[]) -> any[]`),
		newFnSignature(`confirm(prompt: str = "Confirm? [y/n] > ") -> error|bool`),
		newFnSignature(`parse_json(_str: str) -> any|error`),
		newFnSignature(`parse_int(_str: str) -> int|error`),
		newFnSignature(`parse_float(_str: str) -> float|error`),
		newFnSignature(`abs(_num: int|float) -> int|float`),
		newFnSignature(`pow(_base: float, _exponent: float) -> float`),
		newFnSignature(`error(_msg: str) -> error`),
		newFnSignature(`input(prompt: str = "> ", *, hint: str = "", default: str = "", secret: bool = false) -> error|str`),
		newFnSignature(`get_path(_path: str) -> { "exists": bool, "full_path": str, "base_name"?: str, "permissions"?: str, "type"?: str, "size_bytes"?: int, "modified_millis"?: int, "accessed_millis"?: int }`),
		newFnSignature(`get_env(_var: str) -> str`),
		newFnSignature(`find_paths(_path: str, *, depth: int = -1, relative: ["target", "cwd", "absolute"] = "target") -> error|str[]`),
		newFnSignature(`delete_path(_path: str) -> bool`),
		newFnSignature(`count(_str: str, _substr: str) -> int`),
		newFnSignature(`zip(*_lists: list, *, fill: any?, strict: bool = false) -> error|list[]`),
		newFnSignature(`str(_var: any) -> str`),
		newFnSignature(`int(_var: any) -> int|error`),
		newFnSignature(`float(_var: any) -> float|error`),
		newFnSignature(`sum(_nums: float[]) -> error|float`),
		newFnSignature(`trim(_subject: str, _chars: str = " \t\n") -> str`),
		newFnSignature(`trim_prefix(_subject: str, _prefix: str) -> str`),
		newFnSignature(`trim_suffix(_subject: str, _suffix: str) -> str`),
		newFnSignature(`trim_left(_subject: str, _chars: str = " \t\n") -> str`),
		newFnSignature(`trim_right(_subject: str, _chars: str = " \t\n") -> str`),
		newFnSignature(`read_file(_path: str, *, mode: ["text", "bytes"] = "text") -> error|{ "size_bytes": int, "content": str|int[] }`),
		newFnSignature(`write_file(_path: str, _content: str, *, append: bool = false) -> error|{ "bytes_written": int, "path": str }`),
		newFnSignature(`read_stdin() -> str?|error`),
		newFnSignature(`has_stdin() -> bool`),
		newFnSignature(`round(_num: float, _decimals: int = 0) -> error|int|float`),
		newFnSignature(`ceil(_num: float) -> int`),
		newFnSignature(`floor(_num: float) -> int`),
		newFnSignature(`min(*_nums: float|float[]) -> float|error`),
		newFnSignature(`max(*_nums: float|float[]) -> float|error`),
		newFnSignature(`matches(_str: str, _pattern: str, *, partial: bool = false) -> bool|error`),
		newFnSignature(`clamp(val: float, min: float, max: float) -> error|float`),
		newFnSignature(`reverse(_val: str|list) -> str|list`),
		newFnSignature(`is_defined(_var: str) -> bool`),
		newFnSignature(`hyperlink(_val: any, _link: str) -> str`),
		newFnSignature(`uuid_v4() -> str`),
		newFnSignature(`uuid_v7() -> str`),
		newFnSignature(`gen_fid(*, alphabet: str?, tick_size_ms: int?, num_random_chars: int?) -> error|str`),
		newFnSignature(`get_stash_dir(_sub_path: str = "") -> error|str`),
		newFnSignature(`load_state() -> error|map`),
		newFnSignature(`save_state(_state: map) -> error?`),
		newFnSignature(`load_stash_file(_path: str, _default: str = "") -> error|{ "full_path": str, "created": bool, "content"?: str }`),
		newFnSignature(`write_stash_file(_path: str, _content: str) -> error?`),
		newFnSignature(`hash(_val: str, algo: ["sha1", "sha256", "sha512", "md5"] = "sha1") -> str`),
		newFnSignature(`encode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> str`),
		newFnSignature(`decode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> error|str`),
		newFnSignature(`encode_base16(_content: str) -> str`),
		newFnSignature(`decode_base16(_content: str) -> error|str`),
		newFnSignature(`map(_coll: map|list, _fn: fn(any) -> any | fn(any, any) -> any) -> map|list`),
		newFnSignature(`filter(_coll: map|list, _fn: fn(any) -> bool | fn(any, any) -> bool) -> map|list`),
		newFnSignature(`flat_map(_coll: map|list, _fn: any?) -> list`),
		newFnSignature(`get_rad_home() -> str`),
		newFnSignature(`load(_map: map, _key: any, _loader: fn() -> any, *, reload: bool = false, override: any?) -> error|any`),
		newFnSignature(`color_rgb(_val: any, red: int, green: int, blue: int) -> error|str`),
		newFnSignature(`colorize(_val: any, _enum: any[], *, skip_if_single: bool = false) -> str`),
		newFnSignature(`http_get(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`http_post(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`http_put(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`http_patch(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`http_delete(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`http_head(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`http_options(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`http_trace(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`http_connect(url: str, *, body: any?, json: any?, headers: map?) -> { "success": bool, "status_code"?: int, "body"?: any, "error"?: str, "duration_seconds": float }`),
		newFnSignature(`get_args() -> str[]`),
		newFnSignature(`plain(_item: any) -> str`),
		newFnSignature(`black(_item: any) -> str`),
		newFnSignature(`red(_item: any) -> str`),
		newFnSignature(`green(_item: any) -> str`),
		newFnSignature(`yellow(_item: any) -> str`),
		newFnSignature(`blue(_item: any) -> str`),
		newFnSignature(`magenta(_item: any) -> str`),
		newFnSignature(`cyan(_item: any) -> str`),
		newFnSignature(`white(_item: any) -> str`),
		newFnSignature(`orange(_item: any) -> str`),
		newFnSignature(`pink(_item: any) -> str`),
		newFnSignature(`bold(_item: any) -> str`),
		newFnSignature(`italic(_item: any) -> str`),
		newFnSignature(`underline(_item: any) -> str`),

		// Internal signatures
		newFnSignature(`_rad_get_stash_id(*_)`),
		newFnSignature(`_rad_delete_stash(*_)`),
		newFnSignature(`_rad_run_check(*_)`),
		newFnSignature(`_rad_check_from_logs(_duration: str, _verbose: bool) -> void`),
		newFnSignature(`_rad_explain(_code: str) -> str?`),
		newFnSignature(`_rad_explain_list() -> str[]`),
	}

	parser, err := NewRadParser()
	if err != nil {
		panic(fmt.Sprintf("Failed to create parser for fn signatures: %v", err))
	}

	FnSignaturesByName = make(map[string]FnSignature, len(signatures))

	// todo this is actually a bit slow. ~7 millis as of 2025-06-24. meaning every rad script takes at least 7 millis extra.
	//  lazy load? or figure out a fast eager way?
	for _, sig := range signatures {
		fn := fmt.Sprintf("fn %s:\n    pass\n", sig.Signature)
		tree := parser.Parse(fn)

		if invalidNodes := tree.FindInvalidNodes(); len(invalidNodes) > 0 {
			panic(fmt.Sprintf("Invalid function signature syntax: %s", sig.Signature))
		}

		typing := rl.NewTypingFnT(tree.Root().Child(0), tree.src)

		if _, ok := FnSignaturesByName[typing.Name]; ok {
			panic(fmt.Sprintf("Duplicate function signature found: %s", typing.Name))
		}

		sig.Typing = typing
		sig.Name = typing.Name
		FnSignaturesByName[sig.Name] = sig
	}
}
