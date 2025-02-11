package core

import (
	"embed"
	"fmt"
	"strings"

	"github.com/amterp/rts"
)

//go:embed embedded/*
var embeddedFiles embed.FS

const (
	EmbCmdNew = "new"
)

type EmbeddedCmd struct {
	Name        string
	Src         string
	Description string
}

var CmdsByName map[string]EmbeddedCmd

func GetEmbeddedCommandSrc(name string) *string {
	cmd, ok := CmdsByName[name]
	if !ok {
		return nil
	}
	return &cmd.Src
}

func init() {
	embeddedCommands := []EmbeddedCmd{
		{
			Name:        EmbCmdNew,
			Src:         getEmbeddedSrc("new"),
			Description: getFileHeaderLine("new"),
		},
	}

	CmdsByName = make(map[string]EmbeddedCmd)
	for _, cmd := range embeddedCommands {
		CmdsByName[cmd.Name] = cmd
	}
}

func getEmbeddedSrc(name string) string {
	src, err := embeddedFiles.ReadFile("embedded/" + name)
	if err != nil {
		panic(fmt.Sprintf("Failed to read embedded file %s: %s", name, err)) // can't use RP cause not yet loaded
	}
	return string(src)
}

func getFileHeaderLine(fileName string) string {
	src := getEmbeddedSrc(fileName) // todo we're reading it *twice* per command, a little wasteful
	parser, err := rts.NewRslParser()
	if err != nil {
		panic(fmt.Sprintf("Failed to create parser for embedded file %s: %s", fileName, err))
	}
	defer parser.Close()
	tree := parser.Parse(src)
	fh, err := tree.FindFileHeader()
	if err != nil {
		panic(fmt.Sprintf("Failed to find file header in embedded file %s: %s", fileName, err))
	}
	firstLine := strings.Split(fh.Contents, "\n")[0]
	if !strings.HasSuffix(firstLine, ".") {
		// Heuristic for the first line being a decent standalone description, for usage in rad help message.
		panic(fmt.Sprintf("First line in header of embedded file %s isn't a complete sentence: %s", fileName, firstLine))
	}
	return firstLine
}
