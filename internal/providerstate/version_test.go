package providerstate

import "testing"

func TestExtractSemanticVersionFromString(t *testing.T) {
	for _, tt := range []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Incomplete",
			input:    "v1.3",
			expected: "",
		},
		{
			name:     "SimpleWithV",
			input:    "v1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "SimpleWithoutV",
			input:    "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "Docker",
			input:    "v4.5.6-Docker-3.2",
			expected: "4.5.6",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual, _ := ExtractVersion(tt.input)
			if actual != tt.expected {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tt.expected, actual)
			}
		})
	}
}
