package core

import (
	"path/filepath"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts/rl"
)

const ENV_RAD_HOME = "RAD_HOME"

var RadHomeInst *RadHome

type RadHome struct {
	HomeDir string
	StashId *string
}

func NewRadHome(home string) *RadHome {
	return &RadHome{
		HomeDir: home,
	}
}

func (r *RadHome) SetStashId(id string) {
	r.StashId = &id
}

func (r *RadHome) GetStash() *string {
	if r.StashId == nil {
		return nil
	}

	path := r.GetStashForId(*r.StashId)
	return &path
}

func (r *RadHome) GetStashForId(id string) string {
	return NormalizePath(filepath.Join(r.HomeDir, "stashes", id))
}

func (r *RadHome) GetStashSub(subPath string, node rl.Node) (string, *RadError) {
	stash := r.GetStash()
	if stash == nil {
		return "", errNoStashId(node)
	}

	return NormalizePath(filepath.Join(*stash, "files", subPath)), nil
}

func (r *RadHome) LoadState(i *Interpreter, node rl.Node) (RadValue, bool, *RadError) {
	if r.StashId == nil {
		return RadValue{}, false, errNoStashId(node)
	}

	path := r.resolveScriptStatePath()
	if !com.FileExists(path) {
		return newRadValueMap(NewRadMap()), false, nil
	}

	data, err := com.LoadJson(path)
	if err != nil {
		i.emitErrorf(rl.ErrFileRead, node, "Failed to load state: %s", err)
	}

	return ConvertToNativeTypes(i, node, data), true, nil
}

func (r *RadHome) SaveState(i *Interpreter, node rl.Node, value RadValue) *RadError {
	if r.StashId == nil {
		return errNoStashId(node)
	}

	path := r.resolveScriptStatePath()

	json := RadToJsonType(value)
	err := com.CreateFilePathAndWriteJson(path, json)
	if err != nil {
		i.emitErrorf(rl.ErrFileWrite, node, "Failed to save state: %s", err)
	}
	return nil
}

func (r *RadHome) resolveScriptStatePath() string {
	stashHome := *r.GetStash()
	return filepath.Join(stashHome, "state.json")
}

func errNoStashId(node rl.Node) *RadError {
	return NewErrorStrf("Script ID is not set. Set the '%s' macro in the file header.", MACRO_STASH_ID).
		SetCode(rl.ErrNoStashId).
		SetSpan(nodeSpanPtr(node))
}
