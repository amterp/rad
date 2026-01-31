package core

import (
	"embed"
	"strings"
)

//go:embed error_docs/*.md
var errorDocFiles embed.FS

// GetErrorDoc returns the documentation for an error code, or empty string if not found.
// The code should be just the numeric part (e.g., "10001" not "RAD10001").
func GetErrorDoc(code string) string {
	// Try to read the file for this error code
	filename := "error_docs/" + code + ".md"
	content, err := errorDocFiles.ReadFile(filename)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// ListErrorCodes returns all documented error codes.
func ListErrorCodes() []string {
	entries, err := errorDocFiles.ReadDir("error_docs")
	if err != nil {
		return nil
	}

	var codes []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".md") {
			code := strings.TrimSuffix(name, ".md")
			codes = append(codes, code)
		}
	}
	return codes
}
