package core

import (
	"path/filepath"
)

// Pure path string manipulation - no filesystem access, so these work on paths
// that don't (yet) exist, unlike get_path.

var FuncBaseName = BuiltInFunc{
	Name: FUNC_BASE_NAME,
	Execute: func(f FuncInvocation) RadValue {
		path := f.GetStr("_path").Plain()
		return f.Return(NormalizePath(filepath.Base(path)))
	},
}

var FuncDirName = BuiltInFunc{
	Name: FUNC_DIR_NAME,
	Execute: func(f FuncInvocation) RadValue {
		path := f.GetStr("_path").Plain()
		return f.Return(NormalizePath(filepath.Dir(path)))
	},
}

var FuncJoinPaths = BuiltInFunc{
	Name: FUNC_JOIN_PATHS,
	Execute: func(f FuncInvocation) RadValue {
		parts := f.GetList("_parts")
		strs := make([]string, 0, parts.LenInt())
		for _, v := range parts.Values {
			strs = append(strs, v.RequireStr(f.i, f.callNode).Plain())
		}
		return f.Return(NormalizePath(filepath.Join(strs...)))
	},
}
