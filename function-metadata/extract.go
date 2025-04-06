package main

import (
	"os"
	"rad/core"
	"sort"
)

func main() {
	functions := core.FunctionsByName

	names := make([]string, 0, len(functions))
	for name := range functions {
		names = append(names, name)
	}

	sort.Strings(names)

	path := "functions.txt"

	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for idx, name := range names {
		file.WriteString(name)
		if idx < len(names)-1 {
			file.WriteString("\n")
		}
	}
}
