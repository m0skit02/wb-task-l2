package unpack

import "testing"

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasErr   bool
	}{
		{"a4bc2d5e", "aaaabccddddde", false},
		{"abcd", "abcd", false},
		{"45", "", true},
		{"", "", false},
		{"qwe\\4\\5", "qwe45", false},
		{"qwe\\45", "qwe44444", false},
	}

	for _, tt := range tests {
		result, err := Unpack(tt.input)
		if tt.hasErr && err == nil {
			t.Errorf("expected error for input %q", tt.input)
			continue
		}
		if !tt.hasErr && err != nil {
			t.Errorf("unexpected error for input %q: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("for input %q expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}
