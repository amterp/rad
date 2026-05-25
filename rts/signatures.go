package rts

import (
	"fmt"

	"github.com/amterp/rad/rts/rl"
)

//go:generate go run ../tools/gen-funcs-sigs -source ../docs/funcs -out signatures_gen.go
//go:generate go run ../tools/gen-funcs-go -source ../docs/funcs -target embedded_funcs
// gen-funcs-page emits the public reference page under docs-web/.
// It lives in the same //go:generate batch (rather than the
// Makefile) so a contributor who runs `go generate ./rts`
// regenerates *everything* derived from docs/funcs/ in one shot;
// otherwise the public docs page silently drifts whenever someone
// edits a per-function .md and skips Make.
//go:generate go run ../tools/gen-funcs-page -source ../docs/funcs -out ../docs-web/docs/reference/functions.md

type FnSignature struct {
	Name      string
	Signature string
	Typing    *rl.TypingFnT
	// IsInternal marks signatures that exist for the runtime's own use
	// (e.g. _rad_explain wiring up the CLI's `rad explain` flow). They
	// remain callable from a script - that's how the runtime invokes
	// them - but completion, hover, and other user-facing surfaces
	// should filter them out so the public API stays focused.
	IsInternal bool
}

func newFnSignature(signature string) FnSignature {
	return FnSignature{
		Signature: signature,
	}
}

func newInternalFnSignature(signature string) FnSignature {
	return FnSignature{
		Signature:  signature,
		IsInternal: true,
	}
}

var FnSignaturesByName map[string]FnSignature

func GetSignature(name string) *FnSignature {
	if sig, ok := FnSignaturesByName[name]; ok {
		return &sig
	}
	return nil
}

// internalSignatures are the runtime-only `_rad_*` builtins that
// don't have public docs and can't be sourced from docs/funcs/.
// Hand-maintained; keep this list small. Internal sigs are
// filtered out of completion / hover so users never see them.
var internalSignatures = []FnSignature{
	newInternalFnSignature(`_rad_get_stash_id(*_)`),
	newInternalFnSignature(`_rad_delete_stash(*_)`),
	newInternalFnSignature(`_rad_run_check(*_)`),
	newInternalFnSignature(`_rad_check_from_logs(_duration: str, _verbose: bool) -> void`),
	newInternalFnSignature(`_rad_explain(_code: str) -> str?`),
	newInternalFnSignature(`_rad_explain_list() -> str[]`),
}

func init() {
	// Public signatures come from the generated signatures_gen.go
	// (sourced from docs/funcs/*.md). Internal _rad_* signatures
	// live in internalSignatures above. Both go through the same
	// Rad-parser init below to populate Typing.
	signatures := make([]FnSignature, 0, len(publicSignatures)+len(internalSignatures))
	for _, s := range publicSignatures {
		signatures = append(signatures, newFnSignature(s))
	}
	signatures = append(signatures, internalSignatures...)

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

		if _, ok := FnSignaturesByName[typing.FnName]; ok {
			panic(fmt.Sprintf("Duplicate function signature found: %s", typing.FnName))
		}

		sig.Typing = typing
		sig.Name = typing.FnName
		FnSignaturesByName[sig.Name] = sig
	}

	// Pre-convert CST defaults to AST so the interpreter never has to
	// do on-the-fly CST->AST conversion at call time.
	for name, sig := range FnSignaturesByName {
		for i := range sig.Typing.Params {
			param := &sig.Typing.Params[i]
			if param.Default != nil && param.DefaultAST == nil {
				param.DefaultAST = &rl.ASTDefault{
					Node: ConvertExpr(param.Default.Node, param.Default.Src, "<builtin>"),
					Src:  param.Default.Src,
				}
			}
		}
		// Must reassign: range copies the struct value, so mutations
		// through param pointers don't update the map entry.
		FnSignaturesByName[name] = sig
	}
}
