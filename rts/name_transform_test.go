package rts

import "testing"

func TestToExternalName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"force_push", "force-push", "single underscore conversion"},
		{"verbose_mode", "verbose-mode", "single underscore conversion"},
		{"multi_word_flag", "multi-word-flag", "multiple underscores"},
		{"already-dashed", "already-dashed", "already has dashes"},
		{"mixed_dash-name", "mixed-dash-name", "mixed underscore and dash"},
		{"simple", "simple", "no transformation needed"},
		{"", "", "empty string"},
		{"single_", "single-", "trailing underscore"},
		{"_leading", "-leading", "leading underscore"},
		{"__double", "--double", "double underscore"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := ToExternalName(test.input)
			if result != test.expected {
				t.Errorf("ToExternalName(%q) = %q, expected %q",
					test.input, result, test.expected)
			}
		})
	}
}
