package core

import (
	"fmt"
	"strings"
)

type MockResponse struct {
	Pattern  string
	FilePath string
}

type MockResponseSlice []MockResponse

func (m *MockResponseSlice) String() string {
	var result []string
	for _, mock := range *m {
		result = append(result, fmt.Sprintf("%q %s", mock.Pattern, mock.FilePath))
	}
	return strings.Join(result, ", ")
}

func (m *MockResponseSlice) Set(value string) error {
	index := strings.LastIndex(value, ":")

	if index == -1 {
		return fmt.Errorf("invalid format: expected pattern:filePath")
	}

	*m = append(*m, MockResponse{Pattern: value[:index], FilePath: value[index+1:]})
	return nil
}

func (m *MockResponseSlice) Type() string {
	return "mockResponse"
}
