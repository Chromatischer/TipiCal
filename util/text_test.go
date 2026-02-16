package util

import (
	"testing"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected []string
	}{
		{
			name:     "short text",
			input:    "Hello",
			width:    10,
			expected: []string{"Hello"},
		},
		{
			name:     "exact width",
			input:    "Hello World",
			width:    11,
			expected: []string{"Hello World"},
		},
		{
			name:     "wrap at space",
			input:    "Hello World",
			width:    7,
			expected: []string{"Hello", "World"},
		},
		{
			name:     "multiple wraps",
			input:    "The quick brown fox jumps",
			width:    10,
			expected: []string{"The quick", "brown fox", "jumps"},
		},
		{
			name:     "preserve newlines",
			input:    "Line 1\nLine 2",
			width:    20,
			expected: []string{"Line 1", "Line 2"},
		},
		{
			name:     "wrap and preserve newlines",
			input:    "This is a long line\nShort",
			width:    10,
			expected: []string{"This is a", "long line", "Short"},
		},
		{
			name:     "empty string",
			input:    "",
			width:    10,
			expected: []string{""},
		},
		{
			name:     "single word too long",
			input:    "Supercalifragilisticexpialidocious",
			width:    10,
			expected: []string{"Supercalif", "ragilistic", "expialidoc", "ious"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.input, tt.width)
			if len(result) != len(tt.expected) {
				t.Errorf("WrapText() returned %d lines, expected %d\nGot: %v\nExpected: %v",
					len(result), len(tt.expected), result, tt.expected)
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("Line %d: got %q, expected %q", i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestWrapTextWithEmoji(t *testing.T) {
	// Emojis are 1 rune but 2 display columns wide
	input := "🎉 Party time! 🎊"
	width := 12 // "🎉 Party" = 2+1+5 = 8 columns, should fit

	result := WrapText(input, width)

	// Should wrap since emojis are wide
	if len(result) < 2 {
		t.Errorf("Expected wrapping with emojis, got: %v", result)
	}
}
