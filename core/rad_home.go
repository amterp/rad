package core

import (
	"path/filepath"
	com "rad/core/common"

	ts "github.com/tree-sitter/go-tree-sitter"
)

const RAD_HOME_DIR = "RAD_HOME_DIR"

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

func (r *RadHome) SetScriptId(id string) {
	r.StashId = &id
}

func (r *RadHome) GetStash() *string {
	if r.StashId == nil {
		return nil
	}

	path := filepath.Join(r.HomeDir, "stash", *r.StashId)
	return &path
}

func (r *RadHome) GetStashSub(subPath string, i *Interpreter, node *ts.Node) string {
	stash := r.GetStash()
	if stash == nil {
		errMissingScriptId(i, node)
		panic(UNREACHABLE)
	}

	return filepath.Join(*stash, "files", subPath)
}

func (r *RadHome) LoadState(i *Interpreter, node *ts.Node) (RslValue, bool) {
	if r.StashId == nil {
		errMissingScriptId(i, node)
	}

	path := r.resolveScriptStatePath()
	if !com.FileExists(path) {
		return newRslValueMap(NewRslMap()), false
	}

	data, err := com.LoadJson(path)
	if err != nil {
		i.errorf(node, "Failed to load state: %s", err)
	}

	return ConvertToNativeTypes(i, node, data), true
}

func (r *RadHome) SaveState(i *Interpreter, node *ts.Node, value RslValue) {
	if r.StashId == nil {
		errMissingScriptId(i, node)
	}

	path := r.resolveScriptStatePath()

	json := RslToJsonType(value)
	err := com.CreateFilePathAndWriteJson(path, json)
	if err != nil {
		i.errorf(node, "Failed to save state: %s", err)
	}
}

func (r *RadHome) resolveScriptStatePath() string {
	stashHome := *r.GetStash()
	return filepath.Join(stashHome, "state.json")
}
