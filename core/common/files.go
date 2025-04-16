package com

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func LoadJson(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

func CreateFilePathAndWriteJson(path string, jsonData interface{}) error {
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return err
	}
	return writeFileWithDir(path, jsonBytes)
}

func CreateFilePathAndWriteString(path string, str string) error {
	return writeFileWithDir(path, []byte(str))
}

func LoadFile(path string) LoadFileResult {
	data, err := os.ReadFile(path)
	if err != nil {
		return LoadFileResult{Error: fmt.Errorf("failed to read file: %w", err)}
	}
	sizeBytes := int64(len(data))
	return LoadFileResult{Content: string(data), SizeBytes: sizeBytes}
}

type LoadFileResult struct {
	Content   string
	SizeBytes int64
	Error     error
}

func writeFileWithDir(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
