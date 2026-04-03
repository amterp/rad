package lstesting

import (
	"encoding/json"
	"strings"
)

// normalizeJSON pretty-prints a JSON message with sorted keys.
// Go's encoding/json sorts map keys alphabetically when marshaling
// map[string]interface{}, so unmarshal-then-marshal produces deterministic output.
func normalizeJSON(raw json.RawMessage) (string, error) {
	var obj interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

// normalizeMessages normalizes a slice of raw JSON messages into a single
// string with each message pretty-printed and separated by blank lines.
func normalizeMessages(messages []json.RawMessage) (string, error) {
	var parts []string
	for _, msg := range messages {
		normalized, err := normalizeJSON(msg)
		if err != nil {
			return "", err
		}
		parts = append(parts, normalized)
	}
	return strings.Join(parts, "\n\n"), nil
}
