package pg

import (
	"reflect"
	"testing"
)

func TestParseSchemas(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single schema",
			input:    "public",
			expected: []string{"public"},
		},
		{
			name:     "multiple schemas",
			input:    "public,auth,api",
			expected: []string{"public", "auth", "api"},
		},
		{
			name:     "with spaces",
			input:    "public, auth , api",
			expected: []string{"public", "auth", "api"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only commas",
			input:    ",,,",
			expected: nil,
		},
		{
			name:     "trailing comma",
			input:    "public,auth,",
			expected: []string{"public", "auth"},
		},
		{
			name:     "leading comma",
			input:    ",public,auth",
			expected: []string{"public", "auth"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSchemas(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseSchemas(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
