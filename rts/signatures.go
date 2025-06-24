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
		newFnSignature(`pprint(_item: any) -> void`),
		newFnSignature(`debug(*_items: any, *, sep: str = " ", end: str = "\n") -> void`),
		newFnSignature(`exit(_code: int|bool = 0) -> void`),
		newFnSignature(`sleep(_duration: float|str, title: str?) -> void`),
		newFnSignature(`seed_random(_seed: int) -> void`),
		newFnSignature(`rand() -> float`),
		newFnSignature(`rand_int(_arg1: int = 9223372036854775807, _arg2: int?) -> int`),
		newFnSignature(`replace(_original: str, _find: str, _replace: str) -> str`),
		newFnSignature(`len(_val: list|str|map) -> int`),
		newFnSignature(`sort(_val: list|str, *, reverse: bool = false) -> list|str`),
		newFnSignature(`now(tz: str = "local") -> { "date": str, "year": int, "month": int, "day": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }`),
		newFnSignature(`parse_epoch(_epoch: int, *, tz: str = "local", unit: ["auto", "seconds", "milliseconds", "microseconds", "nanoseconds"] = "auto") -> { "date": str, "year": int, "month": int, "day": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }`),
		newFnSignature(`type_of(_var: any) -> ["int", "str", "list", "map", "float"]`),
		newFnSignature(`range(_start: float, _end: float?, _step: float = 1) -> float[]`), // todo weird to always have floats?
		newFnSignature(`join(_list: list, joiner: str = "", prefix: str = "", suffix: str = "") -> str`),
		newFnSignature(`split(_val: str, _sep: str) -> str[]`),
		newFnSignature(`lower(_val: str) -> str`),
		newFnSignature(`upper(_val: str) -> str`),
		newFnSignature(`starts_with(_val: str, _start: str) -> bool`),
		newFnSignature(`ends_with(_val: str, _end: str) -> bool`),
		newFnSignature(`pick(_options: str[], _filter: str?|str[]?, *, prompt: str = "Pick an option") -> str`),
		newFnSignature(`pick_kv(keys: str[], values: any[], _filter: str?|str[]?, *, prompt: str = "Pick an option") -> any`),
		newFnSignature(`pick_from_resource(path: str, _filter: str?, *, prompt: str = "Pick an option") -> any`),
		newFnSignature(`keys(_map: map) -> any[]`),
		newFnSignature(`values(_map: map) -> any[]`),
		newFnSignature(`truncate(_str: str, _len: int) -> str`),
		newFnSignature(`unique(_list: any[]) -> any[]`),
		newFnSignature(`confirm(prompt: str?) -> bool`),
		newFnSignature(`parse_json(_str: str) -> any`),
		newFnSignature(`parse_int(_str: str) -> int|error`),
		newFnSignature(`parse_float(_str: str) -> float|error`),
		newFnSignature(`abs(_num: float) -> float`),
		newFnSignature(`error(_msg: str) -> error`),
		newFnSignature(`input(prompt: str?, *, hint: str?, default: str?, secret: bool = false) -> str`),
		newFnSignature(`get_path(_path: str) -> { "exists": bool, "full_path": str, "base_name"?: str, "permissions"?: str, "type"?: str, "size_bytes"?: int }`),
		newFnSignature(`get_env(_var: str) -> str`),
		newFnSignature(`find_paths(_path: str, *, depth: int?, relative: ["target", "cwd", "absolute"] = "target") -> str[]`),
		newFnSignature(`delete_path(_path: str, *, relative: ["target", "cwd", "absolute"] = "target") -> bool`),
		newFnSignature(`count(_subject: str|any[], _inner: any) -> int`),
		newFnSignature(`zip(*_lists: list, *, fill: any?, strict: bool = false) -> list[]`),
		newFnSignature(`str(_var: any) -> str`),
		newFnSignature(`int(_var: any) -> int|error`),
		newFnSignature(`float(_var: any) -> float|error`),
		newFnSignature(`sum(_nums: float[]) -> float`),
		newFnSignature(`trim(_subject: str, to_trim: str = " ") -> str`),
		newFnSignature(`trim_prefix(_subject: str, to_trim: str = " ") -> str`),
		newFnSignature(`trim_suffix(_subject: str, to_trim: str = " ") -> str`),
		newFnSignature(`read_file(_path: str, *, mode: ["text", "bytes"] = "text") -> error|{ "size_bytes": int, "content": str|[int] }`),
		newFnSignature(`write_file(_path: str, _content: str, *, append: bool = false) -> error|{ "bytes_written": int, "path": str }`),
		newFnSignature(`round(_num: float, _decimals: int = 0) -> float`),
		newFnSignature(`ceil(_num: float) -> int`),
		newFnSignature(`floor(_num: float) -> int`),
		newFnSignature(`min(_num: float[]) -> float|error`),
		newFnSignature(`max(_num: float[]) -> float|error`),
		newFnSignature(`clamp(val: float, min: float, max: float) -> float`),
		newFnSignature(`reverse(_val: str|list) -> str|list`),
		newFnSignature(`is_defined(_var: str) -> bool`),
		newFnSignature(`hyperlink(_str: str, _link: str) -> str`),
		newFnSignature(`uuid_v4() -> str`),
		newFnSignature(`uuid_v7() -> str`),
		newFnSignature(`gen_fid(*, alphabet: str?, tick_size_ms: int?, num_random_chars: int?) -> str`),
		newFnSignature(`get_default(_map: map, key: any, default: any) -> any`),
		newFnSignature(`get_rad_home() -> str`),
		newFnSignature(`get_stash_dir(_sub_path: str?) -> error|str`),
		newFnSignature(`load_state() -> map`),
		newFnSignature(`save_state(_state: map) -> void`),
		newFnSignature(`load_stash_file(_path: str, _default: str = "") -> error|{ "full_path": str, "created": bool, "content"?: str }`),
		newFnSignature(`write_stash_file(_path: str, _content: str) -> error`),
		newFnSignature(`hash(_val: str, algo: ["sha1", "sha256", "sha512", "md5"] = "sha1") -> str`),
		newFnSignature(`encode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> str`),
		newFnSignature(`decode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> str`),
		newFnSignature(`encode_base16(_content: str) -> str`),
		newFnSignature(`decode_base16(_content: str) -> str`),
		newFnSignature(`map(_coll: map|list, _fn: fn(any) -> any | fn(any, any) -> any) -> map|list`),
		newFnSignature(`filter(_coll: map|list, _fn: fn(any) -> bool | fn(any, any) -> bool) -> map|list`),
		newFnSignature(`get_fid(*, alphabet: str?, tick_size_ms: int?, num_random_chars: int?) -> str`),
		newFnSignature(`get_default(_map: map, key: any, default: any) -> any`),
		newFnSignature(`get_rad_home() -> str`),
		newFnSignature(`get_stash_dir(_sub_path: str?) -> error|str`),
		newFnSignature(`load_state() -> map`),
		newFnSignature(`save_state(_state: map) -> void`),
		newFnSignature(`load_stash_file(_path: str, _default: str = "") -> error|{ "full_path": str, "created": bool, "content"?: str }`),
		newFnSignature(`write_stash_file(_path: str, _content: str) -> error`),
		newFnSignature(`load(_map: map, _key: any, _loader: fn() -> any, *, reload: bool = false, override: any?) -> any`),
		newFnSignature(`color_rgb(_str: any, red: int, green: int, blue: int) -> error|str`),
		newFnSignature(`colorize(_val: str, _enum: str[]) -> str`),
		newFnSignature(`http_get(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
		newFnSignature(`http_post(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
		newFnSignature(`http_put(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
		newFnSignature(`http_patch(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
		newFnSignature(`http_delete(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
		newFnSignature(`http_head(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
		newFnSignature(`http_options(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
		newFnSignature(`http_trace(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
		newFnSignature(`http_connect(url: str, *, body: any?, headers: map?) -> { "success": bool, "status_code"?: int, "error"?: str, "duration_seconds"?: float }`),
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
	}

	parser, err := NewRadParser()
	if err != nil {
		panic(fmt.Sprintf("Failed to create parser for fn signatures: %v", err))
	}

	FnSignaturesByName = make(map[string]FnSignature, len(signatures))

	// todo this is actually a bit slow. ~7 millis as of 2025-06-24. meaning every rad script takes at least 7 millis extra.\
	//  lazy load? or figure out a fast eager way?
	for _, sig := range signatures {
		fn := fmt.Sprintf("fn %s:\n    pass\n", sig.Signature)
		tree := parser.Parse(fn)
		typing := rl.NewTypingFnT(tree.Root().Child(0), tree.src)
		sig.Typing = typing
		sig.Name = typing.Name
		FnSignaturesByName[sig.Name] = sig
	}
}
