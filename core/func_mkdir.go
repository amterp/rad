package core

import (
	"os"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/rl"
)

var FuncMkdir = BuiltInFunc{
	Name: FUNC_MKDIR,
	Execute: func(f FuncInvocation) RadValue {
		path := com.ExpandTilde(f.GetStr("_path").Plain())

		newResult := func(created bool) RadValue {
			resultMap := NewRadMap()
			resultMap.SetPrimitiveStr(constPath, NormalizePath(path))
			resultMap.SetPrimitiveBool(constCreated, created)
			return f.Return(resultMap)
		}

		if stat, err := os.Stat(path); err == nil {
			if stat.IsDir() {
				return newResult(false)
			}
			return f.ReturnErrf(rl.ErrFileWrite, "Cannot create directory %q: path exists and is not a directory", NormalizePath(path))
		}

		// 0755 matches Rad's convention for internally created directories.
		if err := os.MkdirAll(path, 0755); err != nil {
			if os.IsPermission(err) {
				return f.Return(NewErrorStr(err.Error()).SetCode(rl.ErrFileNoPermission))
			}
			return f.Return(NewErrorStr(err.Error()).SetCode(rl.ErrFileWrite))
		}

		return newResult(true)
	},
}
