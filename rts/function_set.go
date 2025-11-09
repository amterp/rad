package rts

import (
	"embed"
	"strings"
	"sync"
)

//go:embed embedded/functions.txt
var embeddedFiles embed.FS

type FunctionSet struct {
	names map[string]bool
}

var builtInFunctions *FunctionSet
var loadOnce sync.Once

// GetBuiltInFunctions returns the singleton instance of built-in functions.
// This is thread-safe and loads the functions only once.
func GetBuiltInFunctions() *FunctionSet {
	loadOnce.Do(func() {
		builtInFunctions = loadNewFunctionSet()
	})
	return builtInFunctions
}

func loadNewFunctionSet() *FunctionSet {
	src, err := embeddedFiles.ReadFile("embedded/functions.txt")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(src), "\n")
	names := make(map[string]bool)
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			names[strings.TrimSpace(line)] = true
		}
	}
	return &FunctionSet{
		names: names,
	}
}

func (fs *FunctionSet) Contains(name string) bool {
	_, exists := fs.names[name]
	return exists
}
