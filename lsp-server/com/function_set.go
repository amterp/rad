package com

import (
	"embed"
	"strings"
)

//go:embed embedded/*
var embeddedFiles embed.FS

type FunctionSet struct {
	names map[string]bool
}

func LoadNewFunctionSet() *FunctionSet {
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

func (fs *FunctionSet) Len() int {
	return len(fs.names)
}
