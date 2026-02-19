package util

import (
	"strings"
	"testing"
)

func TestDetectURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []URLMatch
	}{
		{
			name:     "no URLs",
			input:    "Hello world",
			expected: nil,
		},
		{
			name:  "single URL",
			input: "Check https://example.com for details",
			expected: []URLMatch{
				{URL: "https://example.com", Start: 6, End: 25},
			},
		},
		{
			name:  "multiple URLs",
			input: "Visit https://example.com and https://test.org",
			expected: []URLMatch{
				{URL: "https://example.com", Start: 6, End: 25},
				{URL: "https://test.org", Start: 30, End: 46},
			},
		},
		{
			name:  "URL at start",
			input: "https://example.com is the site",
			expected: []URLMatch{
				{URL: "https://example.com", Start: 0, End: 19},
			},
		},
		{
			name:  "URL at end",
			input: "Go to https://example.com",
			expected: []URLMatch{
				{URL: "https://example.com", Start: 6, End: 25},
			},
		},
		{
			name:  "http URL",
			input: "Use http://example.com",
			expected: []URLMatch{
				{URL: "http://example.com", Start: 4, End: 22},
			},
		},
		{
			name:  "URL with path",
			input: "See https://example.com/path/to/page",
			expected: []URLMatch{
				{URL: "https://example.com/path/to/page", Start: 4, End: 36},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectURLs(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("DetectURLs() returned %d matches, expected %d", len(result), len(tt.expected))
				return
			}
			for i, match := range result {
				if match.URL != tt.expected[i].URL {
					t.Errorf("Match %d URL = %q, expected %q", i, match.URL, tt.expected[i].URL)
				}
				if match.Start != tt.expected[i].Start {
					t.Errorf("Match %d Start = %d, expected %d", i, match.Start, tt.expected[i].Start)
				}
				if match.End != tt.expected[i].End {
					t.Errorf("Match %d End = %d, expected %d", i, match.End, tt.expected[i].End)
				}
			}
		})
	}
}

func TestFormatWithHyperlinks(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		maxDisplayLen int
		checkContains []string
		checkNoOSC    bool
	}{
		{
			name:          "no URLs",
			input:         "Hello world",
			maxDisplayLen: 40,
			checkNoOSC:    true,
		},
		{
			name:          "single URL",
			input:         "Check https://example.com for details",
			maxDisplayLen: 40,
			checkContains: []string{"\x1b]8;;https://example.com\x07"},
		},
		{
			name:          "long URL shortened to domain",
			input:         "Visit https://us02web.zoom.us/w/86001173084?tk=abc123",
			maxDisplayLen: 30,
			checkContains: []string{
				"\x1b]8;;https://us02web.zoom.us/w/86001173084?tk=abc123\x07",
				"us02web.zoom.us/…",
			},
		},
		{
			name:          "multiple URLs",
			input:         "Visit https://example.com and https://test.org",
			maxDisplayLen: 40,
			checkContains: []string{
				"\x1b]8;;https://example.com\x07",
				"\x1b]8;;https://test.org\x07",
			},
		},
		{
			name:          "short URL not shortened",
			input:         "Go to https://a.co/path",
			maxDisplayLen: 40,
			checkContains: []string{"https://a.co/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatWithHyperlinks(tt.input, tt.maxDisplayLen)

			if tt.checkNoOSC {
				if strings.Contains(result, "\x1b]8;;") {
					t.Errorf("Expected no OSC 8 sequences, got: %q", result)
				}
				return
			}

			for _, check := range tt.checkContains {
				if !strings.Contains(result, check) {
					t.Errorf("Expected result to contain %q, got: %q", check, result)
				}
			}
		})
	}
}

func TestFormatHyperlinkDirect(t *testing.T) {
	url := "https://example.com"
	display := "example.com"
	result := formatHyperlinkDirect(url, display)

	expected := "\x1b]8;;https://example.com\x07example.com\x1b]8;;\x07"
	if result != expected {
		t.Errorf("formatHyperlinkDirect() = %q, expected %q", result, expected)
	}
}

func TestWrapTextANSI(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		width      int
		minLines   int
		maxLineLen int
	}{
		{
			name:       "plain text wrapping",
			input:      "Hello world this is a test",
			width:      10,
			minLines:   2,
			maxLineLen: 10,
		},
		{
			name:       "text with ANSI codes",
			input:      "\x1b[32mHello\x1b[0m world this is long",
			width:      12,
			minLines:   2,
			maxLineLen: 12,
		},
		{
			name:       "hyperlink preserved",
			input:      "Check \x1b]8;;https://example.com\x07link\x1b]8;;\x07 out",
			width:      10,
			minLines:   2,
			maxLineLen: 10,
		},
		{
			name:       "no wrapping needed",
			input:      "Short text",
			width:      20,
			minLines:   1,
			maxLineLen: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapTextANSI(tt.input, tt.width)
			if len(result) < tt.minLines {
				t.Errorf("WrapTextANSI() returned %d lines, expected at least %d: %v", len(result), tt.minLines, result)
			}

			for i, line := range result {
				visibleLen := DisplayWidthANSI(line)
				if visibleLen > tt.maxLineLen {
					t.Errorf("Line %d has visible width %d, max %d: %q", i, visibleLen, tt.maxLineLen, line)
				}
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello world", "Hello world"},
		{"\x1b[32mGreen\x1b[0m", "Green"},
		{"\x1b]8;;https://example.com\x07link\x1b]8;;\x07", "link"},
		{"\x1b[1mBold\x1b[0m and \x1b[4munderline\x1b[0m", "Bold and underline"},
	}

	for _, tt := range tests {
		result := StripANSI(tt.input)
		if result != tt.expected {
			t.Errorf("StripANSI(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
