package utils

import "testing"

func TestFilterUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "some duplicates",
			input:    []string{"a", "b", "a", "c", "b", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "preserves order",
			input:    []string{"z", "a", "z", "m"},
			expected: []string{"z", "a", "m"},
		},
		{
			name:     "single element",
			input:    []string{"x"},
			expected: []string{"x"},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterUniqueStrings(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}

			for i, v := range tt.expected {
				if result[i] != v {
					t.Errorf("result[%d] = %q, want %q", i, result[i], v)
				}
			}
		})
	}
}

func TestValidatePorts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty/basic
		{name: "empty string", input: "", expected: ""},
		{name: "single port", input: "80", expected: "80"},
		{name: "single port with spaces", input: " 443 ", expected: "443"},

		// Multiple ports
		{name: "multiple ports", input: "80,443,8080", expected: "80,443,8080"},
		{name: "multiple with spaces", input: "80, 443, 8080", expected: "80,443,8080"},

		// Ranges with dash
		{name: "port range dash", input: "1000-2000", expected: "1000-2000"},
		{name: "port range with spaces", input: " 1000 - 2000 ", expected: "1000-2000"},

		// Ranges with colon (converted to dash)
		{name: "port range colon", input: "1000:2000", expected: "1000-2000"},

		// Mixed
		{name: "mixed ports and ranges", input: "80,443,1000-2000,8080", expected: "80,443,1000-2000,8080"},

		// Edge cases - valid
		{name: "port 1", input: "1", expected: "1"},
		{name: "port 65535", input: "65535", expected: "65535"},
		{name: "range 1-65535", input: "1-65535", expected: "1-65535"},

		// Invalid - filtered out
		{name: "port 0", input: "0", expected: ""},
		{name: "port 65536", input: "65536", expected: ""},
		{name: "negative port", input: "-1", expected: ""},
		{name: "non-numeric", input: "abc", expected: ""},
		{name: "range start >= end", input: "2000-1000", expected: ""},
		{name: "range equal", input: "1000-1000", expected: ""},
		{name: "invalid range format", input: "1000-2000-3000", expected: ""},
		{name: "range with invalid start", input: "abc-2000", expected: ""},
		{name: "range with invalid end", input: "1000-abc", expected: ""},
		{name: "range out of bounds", input: "0-65536", expected: ""},

		// Mixed valid and invalid
		{name: "mixed valid invalid", input: "80,invalid,443", expected: "80,443"},
		{name: "mixed valid invalid range", input: "80,2000-1000,443", expected: "80,443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePorts(tt.input)
			if result != tt.expected {
				t.Errorf("ValidatePorts(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkFilterUniqueStrings(b *testing.B) {
	input := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = string(rune('a' + (i % 26)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterUniqueStrings(input)
	}
}

func BenchmarkValidatePorts(b *testing.B) {
	input := "80,443,8080,1000-2000,3000-4000,5000,6000,7000-8000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidatePorts(input)
	}
}
